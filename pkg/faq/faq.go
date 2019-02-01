package faq

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/Azure/draft/pkg/linguist"

	"github.com/jzelinskie/faq/pkg/formats"
	"github.com/jzelinskie/faq/pkg/jq"
)

// Flags are the configuration flags for faq
type Flags struct {
	Debug        bool
	InputFormat  string
	OutputFormat string
	ProgramFile  string
	Raw          bool
	Color        bool
	Monochrome   bool
	Pretty       bool
	Slurp        bool
	ProvideNull  bool
	Args         []string
	Jsonargs     []interface{}
	Kwargs       map[string]string
	Jsonkwargs   map[string]interface{}
	PrintVersion bool
}

var errInvalidOutputFormat = errors.New("invalid output format")

// RunFaq takes a list of files, a jq program, and flags, and runs the
// specified jq program against the files provided, writing results to
// outputWriter.
func RunFaq(outputWriter io.Writer, files []File, program string, flags Flags) error {
	if flags.ProvideNull {
		encoder, ok := formats.ByName(flags.OutputFormat)
		if !ok {
			return errInvalidOutputFormat
		}
		err := runJQ(outputWriter, program, nil, encoder, flags)
		if err != nil {
			return err
		}
		return nil
	}

	// Handle each file path provided.
	if flags.Slurp {
		encoder, ok := formats.ByName(flags.OutputFormat)
		if !ok {
			return errInvalidOutputFormat
		}
		err := slurpFiles(outputWriter, files, program, encoder, flags)
		if err != nil {
			return err
		}
	} else {
		err := processFiles(outputWriter, files, program, flags)
		if err != nil {
			return err
		}
	}

	return nil
}

// processFiles takes a list of files, and for each, attempts to convert it
// to a JSON value and runs the jq program against it
func processFiles(outputWriter io.Writer, files []File, program string, flags Flags) error {
	for _, file := range files {
		decoder, err := determineDecoder(flags.InputFormat, file)
		if err != nil {
			return err
		}

		fileBytes, err := file.Contents()
		if err != nil {
			return err
		}

		if len(bytes.TrimSpace(fileBytes)) != 0 {
			err := convertInputAndRun(outputWriter, decoder, fileBytes, file.Path(), program, flags)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func convertInputAndRun(outputWriter io.Writer, decoder formats.Encoding, fileBytes []byte, path, program string, flags Flags) error {
	encoder, err := determineEncoder(flags.OutputFormat, decoder)
	if err != nil {
		return err
	}

	if streamable, ok := decoder.(formats.Streamable); ok {
		decoder := streamable.NewDecoder(fileBytes)
		for {
			data, err := decoder.MarshalJSONBytes()
			if err == io.EOF {
				break
			}
			if err != nil {
				return fmt.Errorf("failed to jsonify file at %s: `%s`", path, err)
			}

			err = runJQ(outputWriter, program, data, encoder, flags)
			if err != nil {
				return err
			}
		}
		return nil
	}

	data, err := decoder.MarshalJSONBytes(fileBytes)
	if err != nil {
		return fmt.Errorf("failed to jsonify file at %s: `%s`", path, err)
	}

	err = runJQ(outputWriter, program, data, encoder, flags)
	if err != nil {
		return err
	}
	return nil
}

func combineJSONFilesToJSONArray(files []File, inputFormat string) ([]byte, error) {
	var buf bytes.Buffer

	// append the first array bracket
	buf.WriteRune('[')

	// iterate over each file, appending it's contents to an array
	for i, file := range files {
		decoder, err := determineDecoder(inputFormat, file)
		if err != nil {
			return nil, err
		}

		fileBytes, err := file.Contents()
		if err != nil {
			return nil, err
		}

		if streamable, ok := decoder.(formats.Streamable); ok {
			decoder := streamable.NewDecoder(fileBytes)
			var dataList [][]byte
			for {
				data, err := decoder.MarshalJSONBytes()
				if err == io.EOF {
					break
				}
				if err != nil {
					return nil, fmt.Errorf("failed to jsonify file at %s: `%s`", file.Path(), err)
				}
				if len(bytes.TrimSpace(data)) != 0 {
					dataList = append(dataList, data)
				}
			}
			// for each json value in dataList, write it, plus a comma after
			// it, as long it isn't the last item in dataList
			for j, data := range dataList {
				buf.Write(data)
				if j != len(dataList)-1 {
					buf.WriteRune(',')
				}
			}
			// append a comma between each file
			if len(dataList) != 0 && i != len(files)-1 {
				buf.WriteRune(',')
			}
		} else {
			data, err := decoder.MarshalJSONBytes(fileBytes)
			if err != nil {
				return nil, fmt.Errorf("failed to jsonify file at %s: `%s`", file.Path(), err)
			}
			if len(bytes.TrimSpace(data)) != 0 {
				buf.Write(data)
				if i != len(files)-1 {
					buf.WriteRune(',')
				}
			}
		}

	}
	// append the last array bracket
	buf.WriteRune(']')

	return buf.Bytes(), nil
}

// slurpFiles takes a list of files, and for each, attempts to convert it to
// a JSON value and appends each JSON value to an array, and passes that array
// as the input to the jq program.
func slurpFiles(outputWriter io.Writer, files []File, program string, encoder formats.Encoding, flags Flags) error {
	data, err := combineJSONFilesToJSONArray(files, flags.InputFormat)
	if err != nil {
		return err
	}

	err = runJQ(outputWriter, program, data, encoder, flags)
	if err != nil {
		return err
	}

	return nil
}

func runJQ(outputWriter io.Writer, program string, input []byte, encoder formats.Encoding, flags Flags) error {
	if flags.ProvideNull {
		input = []byte("null")
	}

	inputArgs := programArguments{flags.Args, flags.Jsonargs, flags.Kwargs, flags.Jsonkwargs}
	args, err := parseArgs(input, inputArgs)
	if err != nil {
		return err
	}

	outputs, err := jq.Exec(program, args, input)
	if err != nil {
		return err
	}

	for _, output := range outputs {
		err := printValue(output, outputWriter, encoder, flags)
		if err != nil {
			return err
		}
	}

	return nil
}

func printValue(jqOutput string, outputWriter io.Writer, encoder formats.Encoding, flags Flags) error {
	output, err := encoder.UnmarshalJSONBytes([]byte(jqOutput))
	if err != nil {
		return fmt.Errorf("failed to encode jq program output as %s: %s", flags.OutputFormat, err)
	}

	if flags.Pretty {
		output, err = encoder.PrettyPrint(output)
		if err != nil {
			return fmt.Errorf("failed to encode jq program output as pretty %s: %s", flags.OutputFormat, err)
		}
	}

	if flags.Raw {
		output, err = encoder.Raw(output)
		if err != nil {
			return fmt.Errorf("failed to encode jq program output as raw %s: %s", flags.OutputFormat, err)
		}
	} else if flags.Color && !flags.Monochrome {
		output, err = encoder.Color(output)
		if err != nil {
			return fmt.Errorf("failed to encode jq program output as color %s: %s", flags.OutputFormat, err)
		}
	}

	fmt.Fprintln(outputWriter, string(output))
	return nil
}

type programArguments struct {
	Args       []string
	Jsonargs   []interface{}
	Kwargs     map[string]string
	Jsonkwargs map[string]interface{}
}

func parseArgs(jsonBytes []byte, jqArgs programArguments) ([]byte, error) {
	var positionalArgsArray []interface{}
	programArgs := make(map[string]interface{})
	namedArgs := make(map[string]interface{})

	for _, value := range jqArgs.Args {
		positionalArgsArray = append(positionalArgsArray, value)
	}
	positionalArgsArray = append(positionalArgsArray, jqArgs.Jsonargs...)
	for key, value := range jqArgs.Kwargs {
		programArgs[key] = value
		namedArgs[key] = value
	}
	for key, value := range jqArgs.Jsonkwargs {
		programArgs[key] = value
		namedArgs[key] = value
	}

	programArgs["ARGS"] = map[string]interface{}{
		"positional": positionalArgsArray,
		"named":      namedArgs,
	}

	return json.Marshal(programArgs)
}

func determineDecoder(inputFormat string, file File) (formats.Encoding, error) {
	var decoder formats.Encoding
	var err error
	if inputFormat == "auto" {
		decoder, err = detectFormat(file)
	} else {
		var ok bool
		decoder, ok = formats.ByName(inputFormat)
		if !ok {
			err = fmt.Errorf("no supported format found named %s", inputFormat)
		}
	}
	if err != nil {
		return nil, err
	}

	return decoder, nil
}

func determineEncoder(outputFormat string, decoder formats.Encoding) (formats.Encoding, error) {
	var encoder formats.Encoding
	var ok bool
	if outputFormat == "auto" {
		encoder = decoder
	} else {
		encoder, ok = formats.ByName(outputFormat)
		if !ok {
			return nil, fmt.Errorf("no supported format found named %s", outputFormat)
		}
	}

	return encoder, nil
}

func detectFormat(file File) (formats.Encoding, error) {
	if ext := filepath.Ext(file.Path()); ext != "" {
		if format, ok := formats.ByName(ext[1:]); ok {
			return format, nil
		}
	}

	fileBytes, err := file.Contents()
	if err != nil {
		return nil, err
	}

	format := strings.ToLower(linguist.Analyse(fileBytes, linguist.LanguageHints(file.Path())))

	// If linguist doesn't detect care about then try to take a better guess.
	if _, ok := formats.ByName(format); !ok {
		// Look for either {, <, or --- at the beginning of the file to detect
		// json/xml/yaml.
		scanner := bufio.NewScanner(bytes.NewReader(fileBytes))
		keepScanning := true
		for keepScanning && scanner.Scan() {
			line := scanner.Bytes()
			for i, b := range line {
				// Go through each byte until we find a non-whitespace
				// character.
				if !unicode.IsSpace(rune(b)) {
					// If it's either character we're looking for, set the
					// correct format.
					if b == '{' {
						format = "json"
					} else if b == '<' {
						format = "xml"
					} else if b == '-' {
						// If we run into a -, then check if there is a yaml
						// document separator ---.
						if len(line[i:]) >= 3 && bytes.Equal(line[i:i+3], []byte("---")) {
							format = "yaml"
						}
					}
					// Break here because if the first non-whitespace character
					// isn't what we're looking for, then we didn't find what
					// we're looking for.
					keepScanning = false
					break
				}
			}
		}
	}

	// Go isn't smart enough to do this in one line.
	enc, ok := formats.ByName(format)
	if !ok {
		return nil, errors.New("failed to detect format of the input")
	}
	return enc, nil
}
