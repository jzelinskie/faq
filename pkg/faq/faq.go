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
	"github.com/sirupsen/logrus"
)

// ProcessEachFile takes a list of files, and for each, attempts to convert it
// to a JSON value and runs ExecuteProgram against each.
func ProcessEachFile(inputFormat string, files []File, program string, programArgs ProgramArguments, outputWriter io.Writer, outputEncoding formats.Encoding, outputConf OutputConfig, rawOutput bool) error {
	encoder := outputEncoding.NewEncoder(outputWriter)
	for fileNum, file := range files {
		decoderEncoding, err := determineEncoding(inputFormat, file)
		if err != nil {
			return err
		}

		fileBytes, err := file.Contents()
		if err != nil {
			return err
		}

		decoder := decoderEncoding.NewDecoder(bytes.NewBuffer(fileBytes))

		if len(bytes.TrimSpace(fileBytes)) != 0 {
			itemNum := 1
			for {
				data, err := decoder.MarshalJSONBytes()
				if err == io.EOF {
					break
				}
				if err != nil {
					return fmt.Errorf("failed to jsonify file at %s: `%s`", file.Path(), err)
				}

				logrus.Debugf("file: %s (item %d), jsonified:\n%s", file.Path(), fileNum, string(data))

				err = processInput(&data, program, programArgs, encoder, outputConf, rawOutput)
				if err != nil {
					return err
				}
				itemNum++
			}
		}
	}

	return nil
}

// SlurpAllFiles takes a list of files, and for each, attempts to convert it to
// a JSON value and appends each JSON value to an array, and passes that array
// as the input ExecuteProgram.
func SlurpAllFiles(inputFormat string, files []File, program string, programArgs ProgramArguments, outputWriter io.Writer, encoding formats.Encoding, outputConf OutputConfig, rawOutput bool) error {
	data, err := combineJSONFilesToJSONArray(files, inputFormat)
	if err != nil {
		return err
	}

	var paths []string
	for _, f := range files {
		paths = append(paths, f.Path())
	}
	logrus.Debugf("files: %q, jsonified:\n%s", paths, string(data))

	encoder := encoding.NewEncoder(outputWriter)
	return processInput(&data, program, programArgs, encoder, outputConf, rawOutput)
}

// ProcessInput takes input, a single JSON value, and runs program via libjq
// against it, writing the results to outputWriter.
func ProcessInput(input *[]byte, program string, programArgs ProgramArguments, outputWriter io.Writer, encoding formats.Encoding, outputConf OutputConfig, rawOutput bool) error {
	encoder := encoding.NewEncoder(outputWriter)
	return processInput(input, program, programArgs, encoder, outputConf, rawOutput)
}

func processInput(input *[]byte, program string, programArgs ProgramArguments, encoder formats.Encoder, outputConf OutputConfig, rawOutput bool) error {
	outputs, err := ExecuteProgram(input, program, programArgs, rawOutput)
	if err != nil {
		return err
	}

	for _, output := range outputs {
		err := encoder.UnmarshalJSONBytes([]byte(output), outputConf.Color, outputConf.Pretty)
		if err != nil {
			return err
		}
	}

	return nil
}

// ExecuteProgram takes input, a single JSON value, and runs program via libjq
// against it, returning the results.
func ExecuteProgram(input *[]byte, program string, programArgs ProgramArguments, rawOutput bool) ([]string, error) {
	if input == nil {
		input = new([]byte)
		*input = []byte("null")
	}

	args, err := marshalJqArgs(*input, programArgs)
	if err != nil {
		return nil, err
	}

	return jq.Exec(program, args, *input, rawOutput)
}

func combineJSONFilesToJSONArray(files []File, inputFormat string) ([]byte, error) {
	var buf bytes.Buffer

	// append the first array bracket
	buf.WriteRune('[')

	// iterate over each file, appending it's contents to an array
	for i, file := range files {
		encoding, err := determineEncoding(inputFormat, file)
		if err != nil {
			return nil, err
		}

		fileBytes, err := file.Contents()
		if err != nil {
			return nil, err
		}

		decoder := encoding.NewDecoder(bytes.NewBuffer(fileBytes))
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
	}
	// append the last array bracket
	buf.WriteRune(']')

	return buf.Bytes(), nil
}

// OutputConfig contains configuration for out to print out values
type OutputConfig struct {
	Pretty bool
	Color  bool
}

// ProgramArguments contains the arguments to a JQ program
type ProgramArguments struct {
	Args       []string
	Jsonargs   []interface{}
	Kwargs     map[string]string
	Jsonkwargs map[string]interface{}
}

func marshalJqArgs(jsonBytes []byte, jqArgs ProgramArguments) ([]byte, error) {
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

func determineEncoding(format string, file File) (formats.Encoding, error) {
	var encoding formats.Encoding
	var err error
	if format == "auto" {
		encoding, err = detectFormat(file)
	} else {
		var ok bool
		encoding, ok = formats.ByName(format)
		if !ok {
			err = fmt.Errorf("no supported format found named %s", format)
		}
	}
	if err != nil {
		return nil, err
	}

	return encoding, nil
}

var yamlSeparator = []byte("---")

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
						if len(line[i:]) >= 3 && bytes.Equal(line[i:i+3], yamlSeparator) {
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
