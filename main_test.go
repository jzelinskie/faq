package main

import (
	"bytes"
	"testing"
)

func TestRunFaq(t *testing.T) {
	testCases := []struct {
		name              string
		program           string
		inputFileContents []string
		flags             flags
		expectedOutput    string
	}{
		{
			name:              "empty file simple program",
			program:           ".",
			inputFileContents: []string{},
		},
		{
			name:    "single file empty object simple program",
			program: ".",
			inputFileContents: []string{
				`{}`,
			},
			expectedOutput: "{}\n",
			flags: flags{
				inputFormat:  "json",
				outputFormat: "json",
			},
		},
		{
			name:    "single file multiple simple object json stream simple program",
			program: ".",
			inputFileContents: []string{
				"{}\n{}\n",
			},
			expectedOutput: "{}\n{}\n",
			flags: flags{
				inputFormat:  "json",
				outputFormat: "json",
			},
		},
		{
			name:    "single file multiple complex object json stream simple program",
			program: ".",
			inputFileContents: []string{
				`{}
{
}
{
   "bar": 2,
   "foo": true
}
`,
			},
			expectedOutput: `{}
{}
{"bar":2,"foo":true}
`,
			flags: flags{
				inputFormat:  "json",
				outputFormat: "json",
			},
		},
		{
			name:    "single file empty string simple program",
			program: ".",
			inputFileContents: []string{
				`""`,
			},
			expectedOutput: `""` + "\n",
			flags: flags{
				inputFormat:  "json",
				outputFormat: "json",
			},
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
			flags: flags{
				inputFormat:  "yaml",
				outputFormat: "json",
			},
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
			flags: flags{
				inputFormat:  "yaml",
				outputFormat: "json",
			},
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
			flags: flags{
				inputFormat:  "yaml",
				outputFormat: "json",
			},
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
			flags: flags{
				inputFormat:  "yaml",
				outputFormat: "json",
			},
		},
		{
			name:    "single file bool simple program",
			program: ".",
			inputFileContents: []string{
				`true`,
			},
			expectedOutput: "true\n",
			flags: flags{
				inputFormat:  "json",
				outputFormat: "json",
			},
		},
		{
			name:    "multiple file simple program",
			program: ".",
			inputFileContents: []string{
				`{}`,
				`true`,
			},
			expectedOutput: "{}\ntrue\n",
			flags: flags{
				inputFormat:  "json",
				outputFormat: "json",
			},
		},
		{
			name:    "slurp single file empty",
			program: ".",
			inputFileContents: []string{
				``,
			},
			expectedOutput: "[]\n",
			flags: flags{
				inputFormat:  "json",
				outputFormat: "json",
				slurp:        true,
			},
		},
		{
			name:    "slurp multiple file empty",
			program: ".",
			inputFileContents: []string{
				``,
				``,
			},
			expectedOutput: "[]\n",
			flags: flags{
				inputFormat:  "json",
				outputFormat: "json",
				slurp:        true,
			},
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
			flags: flags{
				inputFormat:  "json",
				outputFormat: "json",
				slurp:        true,
			},
		},
		{
			name:    "slurp multiple file json stream",
			program: ".",
			inputFileContents: []string{
				`
{}
{}
{
	"bar": 2,
	"foo": true
}
`,
				``,   // empty files are ignored
				`""`, // an empty string is valid
				`true`,
			},
			expectedOutput: `[{},{},{"bar":2,"foo":true},"",true]` + "\n",
			flags: flags{
				inputFormat:  "json",
				outputFormat: "json",
				slurp:        true,
			},
		},
		{
			name:    "slurp multiple file yaml stream simple program",
			program: ".",
			inputFileContents: []string{
				`---
foo: true
---
bar: false`,
				`---
fizz: buzz
---
cats: dogs
`,
			},
			expectedOutput: `[{"foo":true},{"bar":false},{"fizz":"buzz"},{"cats":"dogs"}]
`,
			flags: flags{
				inputFormat:  "yaml",
				outputFormat: "json",
				slurp:        true,
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			var fileInfos []*fileInfo
			for i, fileContent := range testCase.inputFileContents {
				fileInfos = append(fileInfos, &fileInfo{
					path: "test-path-" + string(i),
					read: true,
					data: []byte(fileContent),
				})
			}
			var outputBuf bytes.Buffer
			err := runFaq(&outputBuf, fileInfos, testCase.program, testCase.flags)
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
