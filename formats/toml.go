package formats

import (
	"bytes"
	"encoding/json"

	"github.com/BurntSushi/toml"
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

func init() {
	ByName["toml"] = tomlEncoding{}
}
