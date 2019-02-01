package commands

import (
	"fmt"
	"os"
	"runtime"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"

	"github.com/jzelinskie/faq/pkg/faq"
	"github.com/jzelinskie/faq/pkg/flagutil"
	"github.com/jzelinskie/faq/pkg/version"
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
	var flags faq.Flags

	stringKwargsFlag := flagutil.NewKwargStringFlag(&flags.Kwargs)
	jsonKwargsFlag := flagutil.NewKwargStringFlag(&flags.Jsonkwargs)
	stringPositionalArgsFlag := flagutil.NewPositionalArgStringFlag(&flags.Args)
	jsonPositionalArgsFlag := flagutil.NewPositionalArgBytesFlag(&flags.Jsonargs)

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
	rootCmd.Flags().BoolVarP(&flags.Raw, "raw-output", "r", false, "output raw strings, not JSON texts")
	rootCmd.Flags().BoolVarP(&flags.Color, "color-output", "c", true, "colorize the output")
	rootCmd.Flags().BoolVarP(&flags.Monochrome, "monochrome-output", "m", false, "monochrome (don't colorize the output)")
	rootCmd.Flags().BoolVarP(&flags.Pretty, "pretty-output", "p", true, "pretty-printed output")
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

func runCmdFunc(cmd *cobra.Command, args []string, flags faq.Flags) error {
	if flags.Debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	if flags.PrintVersion {
		fmt.Println(version.Version)
		return nil
	}

	if runtime.GOOS == "windows" {
		flags.Monochrome = true
	}

	// Check to see execution is in an interactive terminal and set the args
	// and flags as such.
	program := ""
	paths := []string{}

	// Determine the jq program and arguments if a unix being used or not.
	if !terminal.IsTerminal(int(os.Stdin.Fd())) {
		switch {
		case flags.ProvideNull && len(args) == 0:
			program = "."
		case flags.ProvideNull && len(args) == 1:
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
		case flags.ProvideNull && len(args) >= 1:
			program = args[0]
		case !flags.ProvideNull && len(args) >= 2:
			program = args[0]
			paths = args[1:]
		default:
			return fmt.Errorf("not enough arguments provided")
		}
	}

	// Verify all files exist, and open them.
	var fileInfos []faq.File
	for _, path := range paths {
		fileInfo, err := faq.OpenFile(path, flags)
		if err != nil {
			return err
		}
		fileInfos = append(fileInfos, fileInfo)
	}

	return faq.RunFaq(os.Stdout, fileInfos, program, flags)
}
