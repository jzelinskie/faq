package formats

import "github.com/ghodss/yaml"

type yamlEncoding struct{}

func (yamlEncoding) MarshalJSONBytes(yamlBytes []byte) ([]byte, error) {
	return yaml.YAMLToJSON(yamlBytes)
}

func (yamlEncoding) UnmarshalJSONBytes(jsonBytes []byte) ([]byte, error) {
	return yaml.JSONToYAML(jsonBytes)
}

func init() {
	ByName["yaml"] = yamlEncoding{}
	ByName["yml"] = yamlEncoding{}
}
