package formats

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

// ByName is a mapping from dynamically registered encoding names to Encoding
// implementations.
var ByName = map[string]Encoding{}
