package formats

import (
	"bytes"
	"encoding/json"

	"github.com/alecthomas/chroma/quick"
	"github.com/ghodss/yaml"
	goyaml "gopkg.in/yaml.v2"
)

type yamlEncoding struct{}

func (yamlEncoding) MarshalJSONBytes(yamlBytes []byte) ([]byte, error) {
	return yaml.YAMLToJSON(yamlBytes)
}

func (yamlEncoding) UnmarshalJSONBytes(jsonBytes []byte) ([]byte, error) {
	return yaml.JSONToYAML(jsonBytes)
}

func (yamlEncoding) Raw(yamlBytes []byte) ([]byte, error)         { return yamlBytes, nil }
func (yamlEncoding) PrettyPrint(yamlBytes []byte) ([]byte, error) { return yamlBytes, nil }

func (yamlEncoding) Color(yamlBytes []byte) ([]byte, error) {
	var b bytes.Buffer
	if err := quick.Highlight(&b, string(yamlBytes), "yaml", ChromaFormatter(), ChromaStyle()); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func (yamlEncoding) NewDecoder(yamlBytes []byte) ToJSONDecoder {
	decoder := goyaml.NewDecoder(bytes.NewBuffer(yamlBytes))
	return &yamlToJSONDecoder{decoder}
}

type yamlToJSONDecoder struct {
	decoder *goyaml.Decoder
}

func (d *yamlToJSONDecoder) MarshalJSONBytes() ([]byte, error) {
	var tmp interface{}
	err := d.decoder.Decode(&tmp)
	if err != nil {
		return nil, err
	}

	jsonObj, err := convertToJSONableObject(tmp, nil)
	if err != nil {
		return nil, err
	}

	b, err := json.Marshal(jsonObj)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func init() {
	Register("yaml", yamlEncoding{})
	Register("yml", yamlEncoding{})
}
