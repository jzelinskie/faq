package formats

// Encoding represents any format that is isomorphic with JSON.
type Encoding interface {
	MarshalJSONBytes([]byte) ([]byte, error)
	UnmarshalJSONBytes([]byte) ([]byte, error)
	Raw([]byte) ([]byte, error)
	PrettyPrint([]byte) ([]byte, error)
	Color([]byte) ([]byte, error)
}

// ByName is a mapping from dynamically registered encoding names to Encoding
// implementations.
var ByName = map[string]Encoding{}
