package faq

import (
	"bytes"
	"strconv"
	"testing"

	"github.com/jzelinskie/faq/pkg/formats"
)

func TestProcessEachFile(t *testing.T) {
	testCases := []struct {
		name              string
		program           string
		inputFileContents []string
		inputFormat       string
		outputFormat      string
		expectedOutput    string
		raw               bool
	}{
		{
			name:              "empty file simple program",
			program:           ".",
			inputFileContents: []string{},
			inputFormat:       "json",
			outputFormat:      "json",
		},
		{
			name:    "single file empty object simple program",
			program: ".",
			inputFileContents: []string{
				`{}`,
			},
			expectedOutput: "{}\n",
			inputFormat:    "json",
			outputFormat:   "json",
		},
		{
			name:    "single file multiple simple object json stream simple program",
			program: ".",
			inputFileContents: []string{
				"{}\n{}\n",
			},
			expectedOutput: "{}\n{}\n",
			inputFormat:    "json",
			outputFormat:   "json",
		},
		{
			name:    "single file multiple complex object json stream simple program",
			program: ".",
			inputFileContents: []string{
				`{}
{
}
{
   "foo": true
}
`,
			},
			expectedOutput: `{}
{}
{"foo":true}
`,
			inputFormat:  "json",
			outputFormat: "json",
		},
		{
			name:    "single file empty string simple program",
			program: ".",
			inputFileContents: []string{
				`""`,
			},
			expectedOutput: `""` + "\n",
			inputFormat:    "json",
			outputFormat:   "json",
		},
		{
			name:    "single file yaml stream simple program",
			program: ".",
			inputFileContents: []string{
				`---
foo: true
---
bar: false`,
			},
			expectedOutput: `{"foo":true}
{"bar":false}
`,

			inputFormat:  "yaml",
			outputFormat: "json",
		},
		{
			// TODO: this should return "" as the output probably
			name:    "FIXME single file yaml single empty yaml stream simple program",
			program: ".",
			inputFileContents: []string{
				`---
`,
			},
			expectedOutput: `null
`,
			inputFormat:  "yaml",
			outputFormat: "json",
		},
		{
			// TODO: this should return "" as the output probably
			name:    "FIXME single file yaml multiple empty yaml stream simple program",
			program: ".",
			inputFileContents: []string{
				`---
---
---
`,
			},
			expectedOutput: "null\nnull\nnull\n",
			inputFormat:    "yaml",
			outputFormat:   "json",
		},
		{
			name:    "single file yaml stream with extra newlines simple program",
			program: ".",
			inputFileContents: []string{
				// these extra newlines are intentionally here to ensure
				// they're not treated specially
				`


---

foo: true

---

bar: false

`,
			},
			expectedOutput: `{"foo":true}
{"bar":false}
`,
			inputFormat:  "yaml",
			outputFormat: "json",
		},
		{
			name:    "single file bool simple program",
			program: ".",
			inputFileContents: []string{
				`true`,
			},
			expectedOutput: "true\n",
			inputFormat:    "json",
			outputFormat:   "json",
		},
		{
			name:    "multiple file simple program",
			program: ".",
			inputFileContents: []string{
				`{}`,
				`true`,
			},
			expectedOutput: "{}\ntrue\n",
			inputFormat:    "json",
			outputFormat:   "json",
		},
		{
			name:    "single file empty input raw output",
			program: ".",
			inputFileContents: []string{
				``,
			},
			expectedOutput: "",
			inputFormat:    "json",
			outputFormat:   "json",
			raw:            true,
		},
		{
			name:    "single file empty string input raw output",
			program: ".",
			inputFileContents: []string{
				`""`,
			},
			expectedOutput: "\n",
			inputFormat:    "json",
			outputFormat:   "json",
			raw:            true,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			var files []File
			for i, fileContent := range testCase.inputFileContents {
				files = append(files, &FileInfo{
					path: "test-path-" + string(i),
					read: true,
					data: []byte(fileContent),
				})
			}

			encoding, ok := formats.ByName(testCase.outputFormat)
			if !ok {
				t.Errorf("invalid format: %s", testCase.outputFormat)
			}

			var outputBuf bytes.Buffer
			err := ProcessEachFile(testCase.inputFormat, files, testCase.program, ProgramArguments{}, &outputBuf, encoding, OutputConfig{}, testCase.raw)
			if err != nil {
				t.Errorf("expected no err, got %#v", err)
			}

			output := outputBuf.String()
			if output != testCase.expectedOutput {
				t.Errorf("incorrect output expected=%s, got=%s", testCase.expectedOutput, output)
			}
		})
	}
}

func TestSlurpAllFiles(t *testing.T) {
	testCases := []struct {
		name              string
		program           string
		inputFileContents []string
		inputFormat       string
		outputFormat      string
		expectedOutput    string
		raw               bool
	}{

		{
			name:    "slurp single file empty",
			program: ".",
			inputFileContents: []string{
				``,
			},
			expectedOutput: "[]\n",
			inputFormat:    "json",
			outputFormat:   "json",
		},
		{
			name:    "slurp multiple file empty",
			program: ".",
			inputFileContents: []string{
				``,
				``,
			},
			expectedOutput: "[]\n",
			inputFormat:    "json",
			outputFormat:   "json",
		},
		{
			name:    "slurp multiple file simple",
			program: ".",
			inputFileContents: []string{
				`{}`,
				``,   // empty files are ignored
				`""`, // an empty string is valid
				`true`,
			},
			expectedOutput: `[{},"",true]` + "\n",
			inputFormat:    "json",
			outputFormat:   "json",
		},
		{
			name:    "slurp multiple file json stream",
			program: ".",
			inputFileContents: []string{
				`
 {}
 {}
 {
	 "bar": 2
 }
 `,
				``,   // empty files are ignored
				`""`, // an empty string is valid
				`true`,
			},
			expectedOutput: `[{},{},{"bar":2},"",true]` + "\n",
			inputFormat:    "json",
			outputFormat:   "json",
		},
		{
			name:    "slurp multiple file yaml stream simple program",
			program: ".",
			inputFileContents: []string{
				`---
foo: true
---
bar: false
`,
				`---
fizz: buzz
---
cats: dogs
`,
			},
			expectedOutput: `[{"foo":true},{"bar":false},{"fizz":"buzz"},{"cats":"dogs"}]
`,
			inputFormat:  "yaml",
			outputFormat: "json",
		},
		{
			name:    "single file empty input raw output",
			program: ".",
			inputFileContents: []string{
				``,
			},
			expectedOutput: "[]\n",
			inputFormat:    "json",
			outputFormat:   "json",
			raw:            true,
		},
		{
			name:    "single file empty string input raw output",
			program: ".",
			inputFileContents: []string{
				`""`,
			},
			expectedOutput: "[\"\"]\n",
			inputFormat:    "json",
			outputFormat:   "json",
			raw:            true,
		},
	}
	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			var files []File
			for i, fileContent := range testCase.inputFileContents {
				files = append(files, &FileInfo{
					path: "test-path-" + strconv.Itoa(i),
					read: true,
					data: []byte(fileContent),
				})
			}
			encoder, _ := formats.ByName(testCase.outputFormat)
			var outputBuf bytes.Buffer
			err := SlurpAllFiles(testCase.inputFormat, files, testCase.program, ProgramArguments{}, &outputBuf, encoder, OutputConfig{}, testCase.raw)
			if err != nil {
				t.Errorf("expected no err, got %#v", err)
			}

			output := outputBuf.String()
			if output != testCase.expectedOutput {
				t.Errorf("incorrect output expected=%s, got=%s", testCase.expectedOutput, output)
			}
		})
	}
}

func TestExecuteProgram(t *testing.T) {
	testCases := []struct {
		name           string
		program        string
		input          *[]byte
		inputFormat    string
		outputFormat   string
		programArgs    ProgramArguments
		expectedOutput string
		raw            bool
	}{
		{
			name:           "null input simple program",
			program:        ".",
			input:          nil,
			expectedOutput: "null\n",
			inputFormat:    "json",
			outputFormat:   "json",
		},
	}
	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			encoder, _ := formats.ByName(testCase.outputFormat)
			var outputBuf bytes.Buffer
			err := ProcessInput(testCase.input, testCase.program, testCase.programArgs, &outputBuf, encoder, OutputConfig{}, testCase.raw)
			if err != nil {
				t.Errorf("expected no err, got %#v", err)
			}

			output := outputBuf.String()
			if output != testCase.expectedOutput {
				t.Errorf("incorrect output expected=%s, got=%s", testCase.expectedOutput, output)
			}
		})
	}
}
