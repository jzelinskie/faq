package objconv

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/alecthomas/chroma/quick"
)

var (
	_ Encoding = jsonEncoding{}
	_ Decoder  = &jsonDecoder{}
	_ Encoder  = &jsonEncoder{}
)

type jsonEncoding struct{}

func (jsonEncoding) NewDecoder(r io.Reader) Decoder {
	decoder := json.NewDecoder(r)
	return &jsonDecoder{decoder}
}

func (jsonEncoding) NewEncoder(w io.Writer) Encoder {
	return &jsonEncoder{w}
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

type jsonEncoder struct {
	w io.Writer
}

func (e *jsonEncoder) UnmarshalJSONBytes(input []byte, color, pretty bool) error {
	out, err := internalEncode(e, input, color, pretty)
	if err != nil {
		return err
	}
	fmt.Fprintln(e.w, string(out))
	return nil
}

func (jsonEncoder) unmarshalJSONBytes(jsonBytes []byte) ([]byte, error) {
	// It's already JSON, silly!
	return jsonBytes, nil
}

func (jsonEncoder) prettyPrint(jsonBytes []byte) ([]byte, error) {
	var i interface{}
	err := json.Unmarshal(jsonBytes, &i)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	err = enc.Encode(i)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (jsonEncoder) color(jsonBytes []byte) ([]byte, error) {
	var b bytes.Buffer
	if err := quick.Highlight(&b, string(jsonBytes), "json", ChromaFormatter(), ChromaStyle()); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func init() {
	Register("json", jsonEncoding{})
	Register("js", jsonEncoding{})
	Register("javascript", jsonEncoding{})
}
