package commands

import (
	"fmt"
	"io/ioutil"
	"os"
	"runtime"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"

	"github.com/jzelinskie/faq/pkg/faq"
	"github.com/jzelinskie/faq/pkg/flagutil"
	"github.com/jzelinskie/faq/pkg/formats"
	"github.com/jzelinskie/faq/pkg/version"
)

// NewFaqCommand returns a cobra.Command that
func NewFaqCommand() *cobra.Command {
	var flags flags

	stringKwargsFlag := flagutil.NewKwargStringFlag(&flags.Kwargs)
	jsonKwargsFlag := flagutil.NewKwargJSONFlag(&flags.Jsonkwargs)
	stringPositionalArgsFlag := flagutil.NewPositionalArgStringFlag(&flags.Args)
	jsonPositionalArgsFlag := flagutil.NewPositionalArgJSONFlag(&flags.Jsonargs)

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

	rootCmd.Flags().BoolVar(&flags.Debug, "debug", false, "enable debug logging")
	rootCmd.Flags().StringVarP(&flags.InputFormat, "input-format", "f", "auto", "input format")
	rootCmd.Flags().StringVarP(&flags.OutputFormat, "output-format", "o", "auto", "output format")
	rootCmd.Flags().StringVarP(&flags.ProgramFile, "program-file", "F", "", "If specified, read the file provided as the jq program for faq.")
	rootCmd.Flags().BoolVarP(&flags.Raw, "raw-output", "r", false, "output raw strings, not JSON texts")
	rootCmd.Flags().BoolVarP(&flags.Color, "color-output", "C", true, "colorize the output")
	rootCmd.Flags().BoolVarP(&flags.Monochrome, "monochrome-output", "M", false, "monochrome (don't colorize the output)")
	rootCmd.Flags().BoolVarP(&flags.Pretty, "pretty-output", "p", true, "pretty-printed output")
	rootCmd.Flags().BoolVarP(&flags.Compact, "compact-output", "c", false, "compact output (don't pretty print the output)")
	rootCmd.Flags().BoolVarP(&flags.Slurp, "slurp", "s", false, "read (slurp) all inputs into an array; apply filter to it")
	rootCmd.Flags().BoolVarP(&flags.ProvideNull, "null-input", "n", false, "use `null` as the single input value")
	rootCmd.Flags().Var(stringPositionalArgsFlag, "args", `Takes a value and adds it to the position arguments list. Values are always strings. Positional arguments are available as $ARGS.positional[]. Specify --args multiple times to pass additional arguments.`)
	rootCmd.Flags().Var(jsonPositionalArgsFlag, "jsonargs", `Takes a value and adds it to the position arguments list. Values are parsed as JSON values. Positional arguments are available as $ARGS.positional[]. Specify --jsonargs multiple times to pass additional arguments.`)
	rootCmd.Flags().Var(stringKwargsFlag, "kwargs", `Takes a key=value pair, setting $key to <value>: --kwargs foo=bar sets $foo to "bar". Values are always strings. Named arguments are also available as $ARGS.named[]. Specify --kwargs multiple times to add more arguments.`)
	rootCmd.Flags().Var(jsonKwargsFlag, "jsonkwargs", `Takes a key=value pair, setting $key to the JSON value of <value>: --kwargs foo={"fizz": "buzz"} sets $foo to the json object {"fizz": "buzz"}. Values are parsed as JSON values. Named arguments are also available as $ARGS.named[]. Specify --jsonkwargs multiple times to add more arguments.`)
	rootCmd.Flags().BoolVarP(&flags.PrintVersion, "version", "v", false, "Print the version and exit.")

	_ = rootCmd.Flags().MarkHidden("debug")
	return rootCmd
}

func runCmdFunc(cmd *cobra.Command, args []string, flags flags) error {
	if flags.Debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	if flags.PrintVersion {
		fmt.Println(version.Version)
		return nil
	}

	outputFile := os.Stdout
	var color bool
	// If monochrome is true, disable color, as it takes higher precedence then
	// --color-output.
	// If we're running in Windows, disable color, since it usually doesn't
	// handle colors correctly.
	// If the output isn't a TTY, and color hasn't been explicitly set via the
	// flag, disable color.
	// otherwise, use to the flags values to determine if color is enabled.
	if flags.Monochrome || runtime.GOOS == "windows" || !terminal.IsTerminal(int(outputFile.Fd())) && !cmd.Flags().Changed("color-output") {
		color = false
	} else {
		color = flags.Color && !flags.Monochrome
	}

	// Check to see execution is in an interactive terminal and set the args
	// and flags as such.
	var program string

	var files []faq.File
	if flags.ProgramFile != "" {
		programBytes, err := ioutil.ReadFile(flags.ProgramFile)
		if err != nil {
			return fmt.Errorf("unable to read --program-file %s: err %v", flags.ProgramFile, err)
		}
		program = string(programBytes)
	} else if len(args) == 0 {
		program = "."
	} else if len(args) >= 1 {
		program = args[0]
		args = args[1:]
	}

	if flags.ProvideNull {
		if flags.InputFormat == "auto" {
			flags.InputFormat = "json"
		}
		// Set output format to json if not explicitly set.
		if !cmd.Flags().Changed("output-format") {
			flags.OutputFormat = "json"
		}
	} else {
		if len(args) == 0 {
			files = []faq.File{faq.NewFile("/dev/stdin", os.Stdin)}
		} else if len(args) != 0 {
			// Verify all files exist, and open them.
			for _, path := range args {
				file, err := faq.OpenFile(path)
				if err != nil {
					return err
				}
				defer file.Close()
				files = append(files, file)
			}
		}

	}

	programArgs := faq.ProgramArguments{
		Args:       flags.Args,
		Jsonargs:   flags.Jsonargs,
		Kwargs:     flags.Kwargs,
		Jsonkwargs: flags.Jsonkwargs,
	}
	outputConf := faq.OutputConfig{
		Pretty: !flags.Compact && flags.Pretty,
		Color:  color,
	}

	if flags.ProvideNull {
		// If --output-format is auto, and we're taking a null input, we just
		// default to JSON output
		if flags.OutputFormat == "auto" {
			flags.OutputFormat = "json"
		}
		encoding, ok := formats.ByName(flags.OutputFormat)
		if !ok {
			return fmt.Errorf("invalid --output-format %s", flags.OutputFormat)
		}
		err := faq.ProcessInput(nil, program, programArgs, outputFile, encoding, outputConf, flags.Raw)
		if err != nil {
			return err
		}
		return nil
	}

	if flags.Slurp {
		if flags.OutputFormat == "" {
			return fmt.Errorf("must specify --output-format when using --slurp")
		}
		encoding, ok := formats.ByName(flags.OutputFormat)
		if !ok {
			return fmt.Errorf("invalid --output-format %s", flags.OutputFormat)
		}
		err := faq.SlurpAllFiles(flags.InputFormat, files, program, programArgs, outputFile, encoding, outputConf, flags.Raw)
		if err != nil {
			return err
		}
	} else {
		// If --output-format is auto, then use --input-format as the default
		// output-format, otherwise try to detect the format of the input file
		// and use that as the output format.
		if flags.OutputFormat == "auto" && flags.InputFormat != "auto" {
			flags.OutputFormat = flags.InputFormat
		}
		encoding, newFile, err := faq.DetermineEncoding(flags.OutputFormat, files[0])
		if err != nil {
			return fmt.Errorf("invalid --output-format %s: %v", flags.OutputFormat, err)
		}
		files[0] = newFile
		err = faq.ProcessEachFile(flags.InputFormat, files, program, programArgs, outputFile, encoding, outputConf, flags.Raw)
		if err != nil {
			return err
		}
	}

	return nil
}

// Flags are the configuration flags for faq
type flags struct {
	Debug        bool
	InputFormat  string
	OutputFormat string
	ProgramFile  string
	Raw          bool
	Color        bool
	Monochrome   bool
	Pretty       bool
	Compact      bool
	Slurp        bool
	ProvideNull  bool
	Args         []string
	Jsonargs     []interface{}
	Kwargs       map[string]string
	Jsonkwargs   map[string]interface{}
	PrintVersion bool
}
