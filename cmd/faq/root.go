package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"

	"github.com/jzelinskie/faq/internal/faq"
	"github.com/jzelinskie/faq/internal/version"
	"github.com/jzelinskie/faq/pkg/objconv"
)

func runCmdFunc(cmd *cobra.Command, args []string, flags flags) error {
	if flags.Debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	if flags.PrintVersion {
		fmt.Printf(version.UsageVersion())
		return nil
	}

	if len(args) == 0 && terminal.IsTerminal(int(os.Stdin.Fd())) {
		return errors.New("no arguments provided")
	}

	outputFile := os.Stdout

	// If monochrome is true, disable color, as it takes higher precedence then
	// --color-output.
	// If we're running in Windows, disable color, since it usually doesn't
	// handle colors correctly.
	// If the output isn't a TTY, and color hasn't been explicitly set via the
	// flag, disable color.
	// otherwise, use to the flags values to determine if color is enabled.
	var color bool
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
		Pretty: !flags.Raw && !flags.Compact && flags.Pretty,
		Color:  !flags.Raw && color,
	}

	if flags.ProvideNull {
		// If --output-format is auto, and we're taking a null input, we just
		// default to JSON output
		if flags.OutputFormat == "auto" {
			flags.OutputFormat = "json"
		}
		encoding, ok := objconv.ByName(flags.OutputFormat)
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
		encoding, ok := objconv.ByName(flags.OutputFormat)
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
