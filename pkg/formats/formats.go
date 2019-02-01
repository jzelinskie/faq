package formats

import (
	"strings"
)

// Encoding represents any format that is isomorphic with JSON.
type Encoding interface {
	MarshalJSONBytes([]byte) ([]byte, error)
	UnmarshalJSONBytes([]byte) ([]byte, error)
	Raw([]byte) ([]byte, error)
	PrettyPrint([]byte) ([]byte, error)
	Color([]byte) ([]byte, error)
}

// ToJSONDecoder is a decoder that reads and decodes values that are isomorphic
// with JSON, and produces JSON encoded output
type ToJSONDecoder interface {
	MarshalJSONBytes() ([]byte, error)
}

// Streamable represents any format that is decodable as a stream and
// isomorphic with JSON.
type Streamable interface {
	NewDecoder([]byte) ToJSONDecoder
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
