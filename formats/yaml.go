package formats

import (
	"bytes"

	"github.com/alecthomas/chroma/quick"
	"github.com/ghodss/yaml"
)

type yamlEncoding struct{}

func (yamlEncoding) MarshalJSONBytes(yamlBytes []byte) ([]byte, error) {
	return yaml.YAMLToJSON(yamlBytes)
}

func (yamlEncoding) UnmarshalJSONBytes(jsonBytes []byte) ([]byte, error) {
	return yaml.JSONToYAML(jsonBytes)
}

func (yamlEncoding) Raw(yamlBytes []byte) ([]byte, error) { return yamlBytes, nil }

func (yamlEncoding) Color(yamlBytes []byte) ([]byte, error) {
	var b bytes.Buffer
	if err := quick.Highlight(&b, string(yamlBytes), "yaml", ChromaFormatter(), ChromaStyle()); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func init() {
	ByName["yaml"] = yamlEncoding{}
	ByName["yml"] = yamlEncoding{}
}
