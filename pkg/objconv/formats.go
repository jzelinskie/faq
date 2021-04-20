package objconv

import (
	"bytes"
	"fmt"
	"io"
	"strings"
)

// Decoder is able to decode a format that is isomorphic with JSON.
type Decoder interface {
	// MarshalJSONBytes returns a JSON value per invocation. It should return
	// io.EOF when it reaches the end of it's stream input stream.
	MarshalJSONBytes() ([]byte, error)
}

// Encoder is able to encode JSON input to a format that is isomorphic with JSON.
type Encoder interface {
	UnmarshalJSONBytes(input []byte, color, pretty bool) error
}

// Encoding represents any format that is isomorphic with JSON.
type Encoding interface {
	NewDecoder(io.Reader) Decoder
	NewEncoder(io.Writer) Encoder
}

var nameToFormat = map[string]Encoding{}

// Register maps an encoding name to an Encoding implementation
func Register(name string, format Encoding) {
	nameToFormat[name] = format
}

// ByName is a mapping from dynamically registered encoding names to Encoding
// implementations.
func ByName(name string) (Encoding, bool) {
	format, ok := nameToFormat[strings.ToLower(name)]
	return format, ok
}

// ToName maps an encoder to a registered encoding name.
func ToName(format Encoding) string {
	for name, f := range nameToFormat {
		if f == format {
			return name
		}
	}
	return ""
}

type internalEncoder interface {
	unmarshalJSONBytes([]byte) ([]byte, error)
	prettyPrint([]byte) ([]byte, error)
	color([]byte) ([]byte, error)
}

func internalEncode(encoder internalEncoder, input []byte, color, pretty bool) ([]byte, error) {
	ret, err := encoder.unmarshalJSONBytes(input)
	if err != nil {
		return nil, fmt.Errorf("failed to encode as: %s", err)
	}

	if pretty {
		ret, err = encoder.prettyPrint(ret)
		if err != nil {
			return nil, fmt.Errorf("failed to encode as pretty: %s", err)
		}
	}
	if color {
		ret, err = encoder.color(ret)
		if err != nil {
			return nil, fmt.Errorf("failed to encode as color: %s", err)
		}
	}
	return bytes.TrimSuffix(ret, []byte("\n")), nil
}
