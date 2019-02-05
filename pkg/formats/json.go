package formats

import (
	"bytes"
	"encoding/json"

	"github.com/alecthomas/chroma/quick"
)

type jsonEncoding struct{}

func (jsonEncoding) MarshalJSONBytes(jsonBytes []byte) ([]byte, error) {
	// It's already JSON, silly!
	return jsonBytes, nil
}

func (jsonEncoding) UnmarshalJSONBytes(jsonBytes []byte) ([]byte, error) {
	// It's already JSON, silly!
	return jsonBytes, nil
}

func (jsonEncoding) PrettyPrint(jsonBytes []byte) ([]byte, error) {
	var i interface{}
	err := json.Unmarshal(jsonBytes, &i)
	if err != nil {
		return nil, err
	}

	return json.MarshalIndent(i, "", "  ")
}

func (jsonEncoding) Color(jsonBytes []byte) ([]byte, error) {
	var b bytes.Buffer
	if err := quick.Highlight(&b, string(jsonBytes), "json", ChromaFormatter(), ChromaStyle()); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func (jsonEncoding) NewDecoder(jsonBytes []byte) ToJSONDecoder {
	decoder := json.NewDecoder(bytes.NewBuffer(jsonBytes))
	return &jsonDecoder{decoder}
}

type jsonDecoder struct {
	decoder *json.Decoder
}

func (d *jsonDecoder) MarshalJSONBytes() ([]byte, error) {
	var tmp interface{}
	err := d.decoder.Decode(&tmp)
	if err != nil {
		return nil, err
	}

	b, err := json.Marshal(tmp)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func init() {
	Register("json", jsonEncoding{})
	Register("js", jsonEncoding{})
	Register("javascript", jsonEncoding{})
}
