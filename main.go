package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/ashb/jqrepl/jq"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "faq",
		Short: "format agnostic querier",
		Long:  "faq is like `sed`, but for object-like data using libjq.",
		RunE:  runCmdFunc,
	}

	rootCmd.Flags().Bool("debug", false, "enable debug logging")
	rootCmd.Flags().StringP("format", "f", "json", "object format (e.g. json, yaml, bencode)")
	rootCmd.Flags().BoolP("raw", "r", false, "output raw strings, not JSON texts")
	rootCmd.Execute()
}

func runCmdFunc(cmd *cobra.Command, args []string) error {
	stdoutIsTTY := terminal.IsTerminal(int(os.Stdout.Fd()))
	stdinIsTTY := terminal.IsTerminal(int(os.Stdin.Fd()))
	program := "."
	pathArgs := []string{}
	if stdoutIsTTY && stdinIsTTY && len(args) > 1 {
		program = args[0]
		pathArgs = args[1:]
	} else if !stdoutIsTTY || !stdinIsTTY {
		pathArgs = []string{"/dev/stdin"}
	} else {
		return fmt.Errorf("not enough arguments provided")
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

		formatName, err := cmd.Flags().GetString("format")
		if err != nil {
			panic("failed to find format flag")
		}

		unmarshaledFile, err := unmarshal(formatName, fileBytes)
		if err != nil {
			return fmt.Errorf("failed to unmarshal file at %s: `%s`", path, err)
		}

		fileJv, err := jq.JvFromInterface(unmarshaledFile)
		if err != nil {
			panic("failed to reflect a jv from unmarshalled file")
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

		for _, resultJv := range resultJvs {
			if ok, _ := cmd.Flags().GetBool("raw"); ok {
				printRaw(resultJv, true)
			} else {
				flags := jq.JvPrintPretty | jq.JvPrintSpace1 | jq.JvPrintColour
				fmt.Println(resultJv.Dump(flags))
			}
		}
	}

	return nil
}

func printRaw(resultJv *jq.Jv, ascii bool) {
	if ascii && (resultJv.Kind() == jq.JV_KIND_STRING) {
		fmt.Println(resultJv.Dump(jq.JvPrintAscii))
	}
}
