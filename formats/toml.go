package formats

import (
	"bytes"
	"encoding/json"

	"github.com/BurntSushi/toml"
	"github.com/alecthomas/chroma/quick"
)

type tomlEncoding struct{}

func (tomlEncoding) MarshalJSONBytes(tomlBytes []byte) ([]byte, error) {
	var obj interface{}
	err := toml.Unmarshal(tomlBytes, &obj)
	if err != nil {
		return nil, err
	}
	return json.Marshal(obj)
}

func (tomlEncoding) UnmarshalJSONBytes(jsonBytes []byte) ([]byte, error) {
	var obj interface{}
	err := json.Unmarshal(jsonBytes, &obj)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := toml.NewEncoder(&buf).Encode(obj); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (tomlEncoding) Raw(tomlBytes []byte) ([]byte, error)         { return tomlBytes, nil }
func (tomlEncoding) PrettyPrint(tomlBytes []byte) ([]byte, error) { return tomlBytes, nil }

func (tomlEncoding) Color(tomlBytes []byte) ([]byte, error) {
	var b bytes.Buffer
	if err := quick.Highlight(&b, string(tomlBytes), "toml", ChromaFormatter(), ChromaStyle()); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func init() {
	ByName["toml"] = tomlEncoding{}
}
