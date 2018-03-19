package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"runtime"

	"github.com/Sirupsen/logrus"
	"github.com/ashb/jqrepl/jq"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
)

func main() {
	var rootCmd = &cobra.Command{
		Use: "faq [flags] [filter string] [files...]",
		DisableFlagsInUseLine: true,
		Short: "format agnostic querier",
		Long:  "faq is like `jq`, but for a variety of object-like data formats",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if debug, _ := cmd.Flags().GetBool("debug"); debug {
				logrus.SetLevel(logrus.DebugLevel)
			}
			return nil
		},
		RunE: runCmdFunc,
	}

	rootCmd.Flags().Bool("debug", false, "enable debug logging")
	rootCmd.Flags().StringP("format", "f", "auto", "object format (e.g. json, yaml, bencode)")
	rootCmd.Flags().BoolP("raw", "r", false, "output raw strings, not JSON texts")
	rootCmd.Flags().BoolP("ascii-output", "a", false, "force output to be ascii instead of UTF-8")
	rootCmd.Flags().BoolP("color-output", "C", true, "colorize the output")
	rootCmd.Flags().BoolP("monochrome-output", "M", false, "monochrome (don't colorize the output)")
	rootCmd.Flags().BoolP("sort-keys", "S", false, "sort keys of objects on output")
	rootCmd.Flags().BoolP("compact", "c", false, "compact instead of pretty-printed output")
	rootCmd.Flags().BoolP("tab", "t", false, "use tabs for indentation")

	rootCmd.Flags().MarkHidden("debug")

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

	raw, _ := cmd.Flags().GetBool("raw")
	ascii, _ := cmd.Flags().GetBool("ascii-output")
	color, _ := cmd.Flags().GetBool("color-output")
	sortKeys, _ := cmd.Flags().GetBool("sort-keys")
	compact, _ := cmd.Flags().GetBool("compact")
	tab, _ := cmd.Flags().GetBool("tab")
	monochrome, _ := cmd.Flags().GetBool("monochrome-output")
	if runtime.GOOS == "windows" {
		monochrome = true
	}

	var flags jq.JvPrintFlags
	if sortKeys {
		flags = flags | jq.JvPrintSorted
	}
	if ascii {
		flags = flags | jq.JvPrintAscii
	}
	if stdoutIsTTY {
		flags = flags | jq.JvPrintIsATty
	}
	if color {
		flags = flags | jq.JvPrintColour
	}
	if monochrome {
		flags = flags &^ jq.JvPrintColour
	}
	if !compact {
		flags = flags | jq.JvPrintPretty
	}
	if tab {
		flags = flags | jq.JvPrintTab
	} else {
		flags = flags | jq.JvPrintSpace1
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

		if formatName == "auto" {
			formatName = detectFormat(fileBytes, path)
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
			if raw {
				printRaw(resultJv, ascii, flags)
			} else {
				fmt.Println(resultJv.Dump(flags))
			}
		}
	}

	return nil
}

func printRaw(resultJv *jq.Jv, ascii bool, flags jq.JvPrintFlags) {
	if ascii && (resultJv.Kind() == jq.JV_KIND_STRING) {
		fmt.Println(resultJv.Dump(jq.JvPrintAscii))
	} else if resultJv.Kind() == jq.JV_KIND_STRING {
		resultStr, err := resultJv.String()
		if err != nil {
			panic("failed to convert string jv into a Go string")
		}
		fmt.Println(resultStr)
	} else {
		fmt.Println(resultJv.Dump(flags))
	}
}
