package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/Azure/draft/pkg/linguist"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"

	"github.com/jzelinskie/faq/formats"
	"github.com/jzelinskie/faq/jq"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "faq [flags] [filter string] [files...]",
		Short: "format agnostic querier",
		Long: `faq is a tool intended to be a more flexible "jq", supporting additional formats.
The additional formats are converted into JSON and processed with libjq.

faq is pronounced "fah queue".

Supported formats:
- BSON
- Bencode
- JSON
- TOML
- XML
- YAML
`,
		DisableFlagsInUseLine: true,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if debug, _ := cmd.Flags().GetBool("debug"); debug {
				logrus.SetLevel(logrus.DebugLevel)
			}
			return nil
		},
		RunE: runCmdFunc,
	}

	rootCmd.Flags().Bool("debug", false, "enable debug logging")
	rootCmd.Flags().StringP("input-format", "f", "auto", "input format")
	rootCmd.Flags().StringP("output-format", "o", "auto", "output format")
	rootCmd.Flags().BoolP("raw-output", "r", false, "output raw strings, not JSON texts")
	rootCmd.Flags().BoolP("color-output", "c", true, "colorize the output")
	rootCmd.Flags().BoolP("monochrome-output", "m", false, "monochrome (don't colorize the output)")
	rootCmd.Flags().BoolP("pretty-output", "p", true, "pretty-printed output")
	rootCmd.Flags().BoolP("slurp", "s", false, "read (slurp) all inputs into an array; apply filter to it")

	rootCmd.Flags().MarkHidden("debug")

	rootCmd.Execute()
}

type flags struct {
	inputFormat  string
	outputFormat string
	raw          bool
	color        bool
	monochrome   bool
	pretty       bool
	slurp        bool
}

func runCmdFunc(cmd *cobra.Command, args []string) error {
	var flags flags
	flags.inputFormat, _ = cmd.Flags().GetString("input-format")
	flags.outputFormat, _ = cmd.Flags().GetString("output-format")
	flags.raw, _ = cmd.Flags().GetBool("raw-output")
	flags.color, _ = cmd.Flags().GetBool("color-output")
	flags.pretty, _ = cmd.Flags().GetBool("pretty-output")
	flags.monochrome, _ = cmd.Flags().GetBool("monochrome-output")
	flags.slurp, _ = cmd.Flags().GetBool("slurp")
	if runtime.GOOS == "windows" {
		flags.monochrome = true
	}

	return runFaq(args, flags)
}

func runFaq(args []string, flags flags) error {
	// Check to see execution is in an interactive terminal and set the args
	// and flags as such.
	program := ""
	paths := []string{}

	// Determine the jq program and arguments if a unix being used or not.
	if !terminal.IsTerminal(int(os.Stdin.Fd())) {
		switch {
		case len(args) == 0:
			program = "."
			paths = []string{"/dev/stdin"}
		case len(args) == 1:
			program = args[0]
			paths = []string{"/dev/stdin"}
		case len(args) > 1:
			program = args[0]
			paths = args[1:]
		default:
			return fmt.Errorf("not enough arguments provided")
		}
	} else {
		if len(args) < 2 {
			return fmt.Errorf("not enough arguments provided")
		}
		program = args[0]
		paths = args[1:]
	}

	// verify all files exist, and open them:
	var fileInfos []*fileInfo
	for _, path := range paths {
		fileInfo, err := openFile(path, flags)
		if err != nil {
			return err
		}
		fileInfos = append(fileInfos, fileInfo)
	}

	return runFaq2(os.Stdout, fileInfos, program, flags)
}

func runFaq2(outputWriter io.Writer, fileInfos []*fileInfo, program string, flags flags) error {
	// If stdout isn't an interactive tty, then default to monochrome.
	if !terminal.IsTerminal(int(os.Stdout.Fd())) {
		flags.monochrome = true
	}

	// Handle each file path provided.
	if flags.slurp {
		if flags.outputFormat == "" {
			return fmt.Errorf("must specify --output-format when using --slurp")
		}
		encoder, ok := formatByName(flags.outputFormat)
		if !ok {
			return fmt.Errorf("invalid --output-format %s", flags.outputFormat)
		}
		err := slurpFiles(outputWriter, fileInfos, program, encoder, flags)
		if err != nil {
			return err
		}
	} else {
		err := processFiles(outputWriter, fileInfos, program, flags)
		if err != nil {
			return err
		}
	}

	return nil
}

// processFiles takes a list of files, and for each, attempts to convert it
// to a JSON value and runs the jq program it
func processFiles(outputWriter io.Writer, fileInfos []*fileInfo, program string, flags flags) error {
	for _, fileInfo := range fileInfos {
		decoder, err := determineDecoder(flags.inputFormat, fileInfo)
		if err != nil {
			return err
		}
		data, err := fileInfo.MarshalJSONBytes(decoder)
		if err != nil {
			return err
		}
		encoder, err := determineEncoder(flags.outputFormat, decoder)
		if err != nil {
			return err
		}
		err = runJQ(outputWriter, program, data, encoder, flags)
		if err != nil {
			return err
		}
	}

	return nil
}

func combineJSONFiles(fileInfos []*fileInfo, inputFormat string) ([]byte, error) {
	// we ignore the errors because byte.Buffers generally do not return
	// error on write, and instead only panic when they cannot grow the
	// underlying slice.
	var buf bytes.Buffer

	// append the first array bracket
	buf.WriteRune('[')

	// iterate over each file, appending it's contents to an array
	for i, fileInfo := range fileInfos {
		// only handle files with content

		decoder, err := determineDecoder(inputFormat, fileInfo)
		if err != nil {
			return nil, err
		}

		data, err := fileInfo.MarshalJSONBytes(decoder)
		if err != nil {
			return nil, err
		}
		if len(bytes.TrimSpace(data)) != 0 {
			buf.Write(data)
			// append the comma if it isn't the last item in the array
			if i != len(fileInfos)-1 {
				buf.WriteRune(',')
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
func slurpFiles(outputWriter io.Writer, fileInfos []*fileInfo, program string, encoder formats.Encoding, flags flags) error {
	data, err := combineJSONFiles(fileInfos, flags.inputFormat)
	if err != nil {
		return err
	}

	err = runJQ(outputWriter, program, data, encoder, flags)
	if err != nil {
		return err
	}

	return nil
}

type fileInfo struct {
	path   string
	reader io.Reader
	data   []byte
	read   bool
}

func (info *fileInfo) MarshalJSONBytes(decoder formats.Encoding) ([]byte, error) {
	fileBytes, err := info.GetContents()
	if err != nil {
		return nil, err
	}
	data, err := decoder.MarshalJSONBytes(fileBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to jsonify file at %s: `%s`", info.path, err)
	}
	return data, err
}

func (info *fileInfo) GetContents() ([]byte, error) {
	if !info.read {
		if readCloser, ok := info.reader.(io.ReadCloser); ok {
			defer readCloser.Close()
		}
		var err error
		info.data, err = ioutil.ReadAll(info.reader)
		if err != nil {
			return nil, fmt.Errorf("failed to read file at %s: `%s`", info.path, err)
		}

		info.read = true
	}
	return info.data, nil
}

func openFile(path string, flags flags) (*fileInfo, error) {
	path = os.ExpandEnv(path)
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file at %s: `%s`", path, err)
	}

	return &fileInfo{path: path, reader: file}, nil
}

func runJQ(outputWriter io.Writer, program string, data []byte, encoder formats.Encoding, flags flags) error {
	resultJvs, err := execJQProgram(program, data)
	if err != nil {
		return err
	}

	for _, resultJv := range resultJvs {
		err := printJV(outputWriter, resultJv, encoder, flags)
		if err != nil {
			return err
		}
	}

	return nil
}

func printJV(outputWriter io.Writer, jv *jq.Jv, encoder formats.Encoding, flags flags) error {
	resultBytes := []byte(jv.Dump(jq.JvPrintNone))
	output, err := encoder.UnmarshalJSONBytes(resultBytes)
	if err != nil {
		return fmt.Errorf("failed to encode jq program output as %s: %s", flags.outputFormat, err)
	}

	if flags.pretty {
		output, err = encoder.PrettyPrint(output)
		if err != nil {
			return fmt.Errorf("failed to encode jq program output as pretty %s: %s", flags.outputFormat, err)
		}
	}

	if flags.raw {
		output, err = encoder.Raw(output)
		if err != nil {
			return fmt.Errorf("failed to encode jq program output as raw %s: %s", flags.outputFormat, err)
		}
	} else if flags.color && !flags.monochrome {
		output, err = encoder.Color(output)
		if err != nil {
			return fmt.Errorf("failed to encode jq program output as color %s: %s", flags.outputFormat, err)
		}
	}

	fmt.Fprintln(outputWriter, string(output))
	return nil
}

func execJQProgram(program string, jsonBytes []byte) ([]*jq.Jv, error) {
	libjq, err := jq.New()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize libjq: %s", err)
	}
	defer libjq.Close()

	fileJv, err := jq.JvFromJSONBytes(jsonBytes)
	if err != nil {
		return nil, fmt.Errorf("unable to convert to json value from bytes: %s", err)
	}

	errs := libjq.Compile(program, jq.JvArray())
	for _, err := range errs {
		if err != nil {
			return nil, fmt.Errorf("failed to compile jq program: %s", err)
		}
	}

	resultJvs, err := libjq.Execute(fileJv)
	if err != nil {
		return nil, fmt.Errorf("failed to execute jq program: %s", err)
	}

	return resultJvs, nil
}

func determineDecoder(inputFormat string, fileInfo *fileInfo) (formats.Encoding, error) {
	var decoder formats.Encoding
	var err error
	if inputFormat == "auto" {
		decoder, err = detectFormat(fileInfo)
	} else {
		var ok bool
		decoder, ok = formatByName(inputFormat)
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
		encoder, ok = formatByName(outputFormat)
		if !ok {
			return nil, fmt.Errorf("no supported format found named %s", outputFormat)
		}
	}

	return encoder, nil
}

func detectFormat(fileInfo *fileInfo) (formats.Encoding, error) {
	if ext := filepath.Ext(fileInfo.path); ext != "" {
		if format, ok := formatByName(ext[1:]); ok {
			return format, nil
		}
	}

	fileBytes, err := fileInfo.GetContents()
	if err != nil {
		return nil, err
	}
	format := linguist.LanguageByContents(fileBytes, linguist.LanguageHints(fileInfo.path))
	format = strings.ToLower(format)

	// This is what linguist says when it has no idea what it's talking about.
	// For now, just fallback to JSON.
	if format == "coq" {
		format = "json"
	}

	// Go isn't smart enough to do this in one line.
	enc, ok := formats.ByName[format]
	if !ok {
		return nil, errors.New("failed to detect format of the input")
	}
	return enc, nil
}

func formatByName(name string) (formats.Encoding, bool) {
	if format, ok := formats.ByName[strings.ToLower(name)]; ok {
		return format, true
	}
	return nil, false
}
