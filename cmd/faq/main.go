package main

import (
	"os"

	"github.com/jzelinskie/faq/pkg/flagutil"
	"github.com/spf13/cobra"
)

func main() {
	var flags flags

	stringKwargsFlag := flagutil.NewKwargStringFlag(&flags.Kwargs)
	jsonKwargsFlag := flagutil.NewKwargJSONFlag(&flags.Jsonkwargs)
	stringPositionalArgsFlag := flagutil.NewPositionalArgStringFlag(&flags.Args)
	jsonPositionalArgsFlag := flagutil.NewPositionalArgJSONFlag(&flags.Jsonargs)

	var rootCmd = &cobra.Command{
		Use:   "faq [flags] [filter string] [files...]",
		Short: "format agnostic querier",
		Long: `faq is a tool intended to be a more flexible jq, supporting additional formats.
The additional formats are converted into JSON and processed with libjq.

Supported formats:
- BSON
- Bencode
- JSON
- Property Lists
- TOML
- XML
- YAML

$FAQ_FORMATTER can be set to terminal, terminal16m, json, tokens, html.
$FAQ_STYLE can be set to any of the following themes:
https://xyproto.github.io/splash/docs/

How do you pronounce faq? "Fuck you".
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

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
