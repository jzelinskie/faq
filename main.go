package main

import (
	"errors"
	"fmt"
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

	rootCmd.Flags().MarkHidden("debug")

	rootCmd.Execute()
}

func runCmdFunc(cmd *cobra.Command, args []string) error {
	inputFormat, _ := cmd.Flags().GetString("input-format")
	outputFormat, _ := cmd.Flags().GetString("output-format")
	raw, _ := cmd.Flags().GetBool("raw-output")
	color, _ := cmd.Flags().GetBool("color-output")
	prettyPrint, _ := cmd.Flags().GetBool("pretty-output")
	monochrome, _ := cmd.Flags().GetBool("monochrome-output")
	if runtime.GOOS == "windows" {
		monochrome = true
	}

	// Check to see execution is in an interactive terminal and set the args
	// and flags as such.
	program := ""
	pathArgs := []string{}

	// If stdout isn't an interactive tty, then default to monochrome.
	if !terminal.IsTerminal(int(os.Stdout.Fd())) {
		monochrome = true
	}

	// Determine the jq program and arguments if a unix being used or not.
	if !terminal.IsTerminal(int(os.Stdin.Fd())) {
		switch {
		case len(args) == 0:
			program = "."
			pathArgs = []string{"/dev/stdin"}
		case len(args) == 1:
			program = args[0]
			pathArgs = []string{"/dev/stdin"}
		case len(args) > 1:
			program = args[0]
			pathArgs = args[1:]
		default:
			return fmt.Errorf("not enough arguments provided")
		}
	} else {
		program = args[0]
		pathArgs = args[1:]
	}

	for _, pathArg := range pathArgs {
		libjq, err := jq.New()
		if err != nil {
			return fmt.Errorf("failed to initialize libjq: %s", err)
		}

		// Sucks these won't close until runCmdFunc exits.
		defer libjq.Close()

		path := os.ExpandEnv(pathArg)
		fileBytes, err := ioutil.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read file at %s: `%s`", path, err)
		}

		// If there was no input, there's no output!
		if len(fileBytes) == 0 {
			continue
		}

		var decoder formats.Encoding
		var ok bool
		if inputFormat == "auto" {
			decoder, ok = detectFormat(fileBytes, path)
			if !ok {
				return errors.New("failed to detect format of the input")
			}
		} else {
			decoder, ok = formats.ByName[strings.ToLower(inputFormat)]
			if !ok {
				return fmt.Errorf("no supported format found named %s", inputFormat)
			}
		}

		jsonifiedFile, err := decoder.MarshalJSONBytes(fileBytes)
		if err != nil {
			return fmt.Errorf("failed to jsonify file at %s: `%s`", path, err)
		}

		fileJv, err := jq.JvFromJSONBytes(jsonifiedFile)
		if err != nil {
			panic("failed to convert jsonified file into jv")
		}

		errs := libjq.Compile(program, jq.JvArray())
		for _, err := range errs {
			if err != nil {
				return fmt.Errorf("failed to compile jq program for file at %s: %s", path, err)
			}
		}

		resultJvs, err := libjq.Execute(fileJv)
		if err != nil {
			return fmt.Errorf("failed to execute jq program for file at %s: %s", path, err)
		}

		// Determine the encoding for the file output.
		var encoder formats.Encoding
		if outputFormat == "auto" {
			encoder = decoder
		} else {
			encoder, ok = formats.ByName[strings.ToLower(outputFormat)]
			if !ok {
				return fmt.Errorf("no supported format found named %s", outputFormat)
			}
		}

		// Print the final output.
		for _, resultJv := range resultJvs {
			resultBytes := []byte(resultJv.Dump(jq.JvPrintNone))
			output, err := encoder.UnmarshalJSONBytes(resultBytes)
			if err != nil {
				return fmt.Errorf("failed to encode jq program output as %s: %s", outputFormat, err)
			}

			if prettyPrint {
				output, err = encoder.PrettyPrint(output)
				if err != nil {
					return fmt.Errorf("failed to encode jq program output as pretty %s: %s", outputFormat, err)
				}
			}

			if raw {
				output, err = encoder.Raw(output)
				if err != nil {
					return fmt.Errorf("failed to encode jq program output as raw %s: %s", outputFormat, err)
				}
			} else if color && !monochrome {
				output, err = encoder.Color(output)
				if err != nil {
					return fmt.Errorf("failed to encode jq program output as color %s: %s", outputFormat, err)
				}
			}

			fmt.Println(string(output))
		}
	}

	return nil
}

func detectFormat(fileBytes []byte, path string) (formats.Encoding, bool) {
	if ext := filepath.Ext(path); ext != "" {
		if format, ok := formats.ByName[ext[1:]]; ok {
			return format, true
		}
	}

	format := linguist.LanguageByContents(fileBytes, linguist.LanguageHints(path))
	format = strings.ToLower(format)

	// This is what linguist says when it has no idea what it's talking about.
	// For now, just fallback to JSON.
	if format == "coq" {
		format = "json"
	}

	// Go isn't smart enough to do this in one line.
	enc, ok := formats.ByName[format]
	return enc, ok
}
