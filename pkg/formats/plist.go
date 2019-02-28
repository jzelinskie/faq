package formats

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/alecthomas/chroma/quick"
	"github.com/go-xmlfmt/xmlfmt"
	"howett.net/plist"
)

var (
	_ Encoding = plistEncoding{}
	_ Decoder  = &plistDecoder{}
	_ Encoder  = &plistEncoder{}
)

type plistEncoding struct{}

func (plistEncoding) NewDecoder(r io.Reader) Decoder {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil
	}
	decoder := plist.NewDecoder(bytes.NewReader(b))
	return &plistDecoder{decoder, false}
}

func (e plistEncoding) NewEncoder(w io.Writer) Encoder {
	return &plistEncoder{w}
}

type plistDecoder struct {
	decoder *plist.Decoder
	read    bool
}

func (d *plistDecoder) MarshalJSONBytes() ([]byte, error) {
	if d.read {
		return nil, io.EOF
	}
	var tmp interface{}
	err := d.decoder.Decode(&tmp)
	if err != nil {
		return nil, err
	}
	d.read = true

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

type plistEncoder struct {
	w io.Writer
}

func (e *plistEncoder) UnmarshalJSONBytes(jsonBytes []byte, color, pretty bool) error {
	out, err := internalEncode(e, jsonBytes, color, pretty)
	if err != nil {
		return err
	}
	fmt.Fprintln(e.w, string(out))
	return nil
}

func (plistEncoder) unmarshalJSONBytes(jsonBytes []byte) ([]byte, error) {
	var tmp interface{}
	err := json.NewDecoder(bytes.NewReader(jsonBytes)).Decode(&tmp)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	err2 := plist.NewEncoder(&buf).Encode(tmp)
	if err2 != nil {
		return nil, err2
	}
	return buf.Bytes(), nil
}

func (plistEncoder) prettyPrint(plistXMLBytes []byte) ([]byte, error) {
	return []byte(xmlfmt.FormatXML(string(plistXMLBytes), "", "  ")), nil
}

func (plistEncoder) color(plistXMLBytes []byte) ([]byte, error) {
	var b bytes.Buffer
	if err := quick.Highlight(&b, string(plistXMLBytes), "xml", ChromaFormatter(), ChromaStyle()); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func init() {
	Register("plist", plistEncoding{})
}
