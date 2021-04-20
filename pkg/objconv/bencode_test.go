package objconv

import (
	"bytes"
	"strings"
	"testing"
)

func TestBencodeMarshal(t *testing.T) {
	var table = []struct {
		input  string
		output string
		err    error
	}{
		{"d2:hi2:hie", `{"hi":"hi"}`, nil},
	}

	for _, tt := range table {
		t.Run("", func(t *testing.T) {
			outputBytes, err := bencodeEncoding{}.NewDecoder(strings.NewReader(tt.input)).MarshalJSONBytes()
			if err != tt.err {
				t.Errorf("unexpected error: %s instead of %s", err, tt.err)
			}
			if string(outputBytes) != tt.output {
				t.Errorf("unexpected output: %s instead of %s", outputBytes, tt.output)
			}
		})
	}
}

func TestBencodeUnmarshal(t *testing.T) {
	var table = []struct {
		input  string
		output string
		err    error
	}{
		{`{"hi":"hi"}`, "d2:hi2:hie\n", nil},
	}

	for _, tt := range table {
		t.Run("", func(t *testing.T) {
			var buf bytes.Buffer
			err := bencodeEncoding{}.NewEncoder(&buf).UnmarshalJSONBytes([]byte(tt.input), false, false)
			if err != tt.err {
				t.Errorf("unexpected error: %s instead of %s", err, tt.err)
			}
			output := buf.String()
			if output != tt.output {
				t.Errorf("unexpected output: %s instead of %s", output, tt.output)
			}
		})
	}
}
