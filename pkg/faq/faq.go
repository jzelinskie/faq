package faq

import (
	"bytes"
	"encoding/json"
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
	"github.com/jzelinskie/faq/pkg/flagutil"
	"github.com/jzelinskie/faq/pkg/jq"
)

var (
	version string
)

// ExecuteFaqCmd executes the faq commandline program
func ExecuteFaqCmd() {
	faqCmd := NewFaqCommand()
	err := faqCmd.Execute()
	if err != nil {
		fmt.Printf("error executing %s: %v\n", faqCmd.Name(), err)
		return
	}
}

// NewFaqCommand returns a cobra.Command that
func NewFaqCommand() *cobra.Command {
	var flags flags

	stringKwargsFlag := flagutil.NewKwargStringFlag(&flags.kwargs)
	jsonKwargsFlag := flagutil.NewKwargStringFlag(&flags.jsonkwargs)
	stringPositionalArgsFlag := flagutil.NewPositionalArgStringFlag(&flags.args)
	jsonPositionalArgsFlag := flagutil.NewPositionalArgBytesFlag(&flags.jsonargs)

	var rootCmd = &cobra.Command{
		Use:   "faq [flags] [filter string] [files...]",
		Short: "format agnostic querier",
		Long: `faq is a tool intended to be a more flexible "jq", supporting additional formats.
The additional formats are converted into JSON and processed with libjq.

Supported formats:
- BSON
- Bencode
- JSON
- TOML
- XML
- YAML

$FAQ_FORMATTER can be set to terminal, terminal16m, json, tokens, html.
$FAQ_STYLE can be set to any of the following themes:
https://xyproto.github.io/splash/docs/

How do you pronounce "faq"? Fuck you.
`,
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCmdFunc(cmd, args, flags)
		},
	}

	rootCmd.Flags().BoolVar(&flags.debug, "debug", false, "enable debug logging")
	rootCmd.Flags().StringVarP(&flags.inputFormat, "input-format", "f", "auto", "input format")
	rootCmd.Flags().StringVarP(&flags.outputFormat, "output-format", "o", "auto", "output format")
	rootCmd.Flags().BoolVarP(&flags.raw, "raw-output", "r", false, "output raw strings, not JSON texts")
	rootCmd.Flags().BoolVarP(&flags.color, "color-output", "c", true, "colorize the output")
	rootCmd.Flags().BoolVarP(&flags.monochrome, "monochrome-output", "m", false, "monochrome (don't colorize the output)")
	rootCmd.Flags().BoolVarP(&flags.pretty, "pretty-output", "p", true, "pretty-printed output")
	rootCmd.Flags().BoolVarP(&flags.slurp, "slurp", "s", false, "read (slurp) all inputs into an array; apply filter to it")
	rootCmd.Flags().BoolVarP(&flags.provideNull, "null-input", "n", false, "use `null` as the single input value")
	rootCmd.Flags().Var(stringPositionalArgsFlag, "args", `Takes a value and adds it to the position arguments list. Values are always strings. Positional arguments are available as $ARGS.positional[]. Specify --args multiple times to pass additional arguments.`)
	rootCmd.Flags().Var(jsonPositionalArgsFlag, "jsonargs", `Takes a value and adds it to the position arguments list. Values are parsed as JSON values. Positional arguments are available as $ARGS.positional[]. Specify --jsonargs multiple times to pass additional arguments.`)
	rootCmd.Flags().Var(stringKwargsFlag, "kwargs", `Takes a key=value pair, setting $key to <value>: --kwargs foo=bar sets $foo to "bar". Values are always strings. Named arguments are also available as $ARGS.named[]. Specify --kwargs multiple times to add more arguments.`)
	rootCmd.Flags().Var(jsonKwargsFlag, "jsonkwargs", `Takes a key=value pair, setting $key to the JSON value of <value>: --kwargs foo={"fizz": "buzz"} sets $foo to the json object {"fizz": "buzz"}. Values are parsed as JSON values. Named arguments are also available as $ARGS.named[]. Specify --jsonkwargs multiple times to add more arguments.`)
	rootCmd.Flags().BoolVarP(&flags.printVersion, "version", "v", false, "Print the version and exit.")

	_ = rootCmd.Flags().MarkHidden("debug")
	return rootCmd
}

type flags struct {
	debug        bool
	inputFormat  string
	outputFormat string
	raw          bool
	color        bool
	monochrome   bool
	pretty       bool
	slurp        bool
	provideNull  bool
	args         []string
	jsonargs     [][]byte
	kwargs       map[string][]byte
	jsonkwargs   map[string][]byte
	printVersion bool
}

func runCmdFunc(cmd *cobra.Command, args []string, flags flags) error {
	if flags.debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	if flags.printVersion {
		fmt.Println(version)
		return nil
	}

	if runtime.GOOS == "windows" {
		flags.monochrome = true
	}

	// Check to see execution is in an interactive terminal and set the args
	// and flags as such.
	program := ""
	paths := []string{}

	// Determine the jq program and arguments if a unix being used or not.
	if !terminal.IsTerminal(int(os.Stdin.Fd())) {
		switch {
		case flags.provideNull && len(args) == 0:
			program = "."
		case flags.provideNull && len(args) == 1:
			program = args[0]
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
		switch {
		case flags.provideNull && len(args) >= 1:
			program = args[0]
		case !flags.provideNull && len(args) >= 2:
			program = args[0]
			paths = args[1:]
		default:
			return fmt.Errorf("not enough arguments provided")
		}
	}

	// Verify all files exist, and open them.
	var fileInfos []*fileInfo
	for _, path := range paths {
		fileInfo, err := openFile(path, flags)
		if err != nil {
			return err
		}
		fileInfos = append(fileInfos, fileInfo)
	}

	return runFaq(os.Stdout, fileInfos, program, flags)
}

func runFaq(outputWriter io.Writer, fileInfos []*fileInfo, program string, flags flags) error {
	// If stdout isn't an interactive tty, then default to monochrome.
	if !terminal.IsTerminal(int(os.Stdout.Fd())) {
		flags.monochrome = true
	}

	if flags.provideNull {
		encoder, ok := formatByName(flags.outputFormat)
		if !ok {
			return fmt.Errorf("invalid --output-format %s", flags.outputFormat)
		}
		err := runJQ(outputWriter, program, nil, encoder, flags)
		if err != nil {
			return err
		}
		return nil
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
// to a JSON value and runs the jq program against it
func processFiles(outputWriter io.Writer, fileInfos []*fileInfo, program string, flags flags) error {
	for _, fileInfo := range fileInfos {
		decoder, err := determineDecoder(flags.inputFormat, fileInfo)
		if err != nil {
			return err
		}

		fileBytes, err := fileInfo.GetContents()
		if err != nil {
			return err
		}

		if len(bytes.TrimSpace(fileBytes)) != 0 {
			err := convertInputAndRun(outputWriter, decoder, fileBytes, fileInfo.path, program, flags)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func convertInputAndRun(outputWriter io.Writer, decoder formats.Encoding, fileBytes []byte, path, program string, flags flags) error {
	encoder, err := determineEncoder(flags.outputFormat, decoder)
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

func combineJSONFilesToJSONArray(fileInfos []*fileInfo, inputFormat string) ([]byte, error) {
	var buf bytes.Buffer

	// append the first array bracket
	buf.WriteRune('[')

	// iterate over each file, appending it's contents to an array
	for i, fileInfo := range fileInfos {
		decoder, err := determineDecoder(inputFormat, fileInfo)
		if err != nil {
			return nil, err
		}

		fileBytes, err := fileInfo.GetContents()
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
					return nil, fmt.Errorf("failed to jsonify file at %s: `%s`", fileInfo.path, err)
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
			if len(dataList) != 0 && i != len(fileInfos)-1 {
				buf.WriteRune(',')
			}
		} else {
			data, err := decoder.MarshalJSONBytes(fileBytes)
			if err != nil {
				return nil, fmt.Errorf("failed to jsonify file at %s: `%s`", fileInfo.path, err)
			}
			if len(bytes.TrimSpace(data)) != 0 {
				buf.Write(data)
				if i != len(fileInfos)-1 {
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
func slurpFiles(outputWriter io.Writer, fileInfos []*fileInfo, program string, encoder formats.Encoding, flags flags) error {
	data, err := combineJSONFilesToJSONArray(fileInfos, flags.inputFormat)
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

func runJQ(outputWriter io.Writer, program string, input []byte, encoder formats.Encoding, flags flags) error {
	if flags.provideNull {
		input = []byte("null")
	}

	args, err := parseArgs(input, flags)
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

func printValue(jqOutput string, outputWriter io.Writer, encoder formats.Encoding, flags flags) error {
	output, err := encoder.UnmarshalJSONBytes([]byte(jqOutput))
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

func parseArgs(jsonBytes []byte, flags flags) ([]byte, error) {
	var positionalArgsArray []interface{}
	for _, arg := range flags.args {
		positionalArgsArray = append(positionalArgsArray, arg)
	}

	for _, value := range flags.jsonargs {
		var i interface{}
		err := json.Unmarshal(value, &i)
		if err != nil {
			return nil, fmt.Errorf("unable to decode JSON arg: %v", err)
		}
		positionalArgsArray = append(positionalArgsArray, i)
	}

	programArgs := make(map[string]interface{}, 0)
	namedArgs := make(map[string]interface{}, 0)

	for key, value := range flags.kwargs {
		programArgs[key] = string(value)
		namedArgs[key] = string(value)
	}

	for key, jsonValue := range flags.jsonkwargs {
		var i interface{}
		err := json.Unmarshal(jsonValue, &i)
		if err != nil {
			return nil, fmt.Errorf("unable to decode JSON kwarg: %s", err)
		}

		programArgs[key] = i
		namedArgs[key] = i
	}

	programArgs["ARGS"] = map[string]interface{}{
		"positional": positionalArgsArray,
		"named":      namedArgs,
	}

	return json.Marshal(programArgs)
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
