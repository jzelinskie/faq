package objconv

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/alecthomas/chroma/quick"
	"github.com/ghodss/yaml"
	goyaml "gopkg.in/yaml.v2"
)

var (
	_ Encoding = yamlEncoding{}
	_ Decoder  = &yamlDecoder{}
	_ Encoder  = &yamlEncoder{}
)

const yamlSeparator = "---"

type yamlEncoding struct{}

func (yamlEncoding) NewDecoder(r io.Reader) Decoder {
	decoder := goyaml.NewDecoder(r)
	return &yamlDecoder{decoder}
}

func (e yamlEncoding) NewEncoder(w io.Writer) Encoder {
	return &yamlEncoder{w, false}
}

type yamlDecoder struct {
	decoder *goyaml.Decoder
}

func (d *yamlDecoder) MarshalJSONBytes() ([]byte, error) {
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

type yamlEncoder struct {
	w        io.Writer
	writeSep bool
}

func (e *yamlEncoder) UnmarshalJSONBytes(jsonBytes []byte, color, pretty bool) error {
	if e.writeSep {
		fmt.Fprintln(e.w, yamlSeparator)
	}
	out, err := internalEncode(e, jsonBytes, color, pretty)
	if err != nil {
		return err
	}
	fmt.Fprintln(e.w, string(out))

	e.writeSep = true
	return nil
}

func (yamlEncoder) unmarshalJSONBytes(jsonBytes []byte) ([]byte, error) {
	return yaml.JSONToYAML(jsonBytes)
}

func (yamlEncoder) prettyPrint(yamlBytes []byte) ([]byte, error) { return yamlBytes, nil }

func (yamlEncoder) color(yamlBytes []byte) ([]byte, error) {
	var b bytes.Buffer
	if err := quick.Highlight(&b, string(yamlBytes), "yaml", ChromaFormatter(), ChromaStyle()); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func init() {
	Register("yaml", yamlEncoding{})
	Register("yml", yamlEncoding{})
}
