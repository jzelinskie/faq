package formats

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/BurntSushi/toml"
	"github.com/alecthomas/chroma/quick"
)

var (
	_ Encoding = tomlEncoding{}
	_ Decoder  = &tomlDecoder{}
	_ Encoder  = &tomlEncoder{}
)

type tomlEncoding struct{}

func (tomlEncoding) NewDecoder(r io.Reader) Decoder {
	return &tomlDecoder{r, false}
}

func (e tomlEncoding) NewEncoder(w io.Writer) Encoder {
	return &tomlEncoder{w}
}

type tomlDecoder struct {
	r    io.Reader
	read bool
}

func (d *tomlDecoder) MarshalJSONBytes() ([]byte, error) {
	if d.read {
		return nil, io.EOF
	}
	tomlBytes, err := ioutil.ReadAll(d.r)
	if err != nil {
		return nil, err
	}
	d.read = true
	var obj interface{}
	err = toml.Unmarshal(tomlBytes, &obj)
	if err != nil {
		return nil, err
	}
	return json.Marshal(obj)
}

type tomlEncoder struct {
	w io.Writer
}

func (e tomlEncoder) UnmarshalJSONBytes(jsonBytes []byte, color, pretty bool) error {
	out, err := internalEncode(e, jsonBytes, color, pretty)
	if err != nil {
		return err
	}
	fmt.Fprintln(e.w, string(out))
	return nil
}

func (tomlEncoder) unmarshalJSONBytes(jsonBytes []byte) ([]byte, error) {
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

func (tomlEncoder) prettyPrint(tomlBytes []byte) ([]byte, error) { return tomlBytes, nil }

func (tomlEncoder) color(tomlBytes []byte) ([]byte, error) {
	var b bytes.Buffer
	if err := quick.Highlight(&b, string(tomlBytes), "toml", ChromaFormatter(), ChromaStyle()); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func init() {
	Register("toml", tomlEncoding{})
}
