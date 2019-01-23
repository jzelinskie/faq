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
			err := runFaq2(&outputBuf, fileInfos, testCase.program, testCase.flags)
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
