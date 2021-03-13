package formats

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"reflect"

	"github.com/alecthomas/chroma/quick"
	"github.com/clbanning/mxj/v2"
)

var (
	_ Encoding = xmlEncoding{}
	_ Decoder  = &xmlDecoder{}
	_ Encoder  = &xmlEncoder{}
)

type xmlEncoding struct{}

func (xmlEncoding) NewDecoder(r io.Reader) Decoder {
	return &xmlDecoder{r, false}
}

func (e xmlEncoding) NewEncoder(w io.Writer) Encoder {
	return &xmlEncoder{w}
}

type xmlDecoder struct {
	r    io.Reader
	read bool
}

func (d *xmlDecoder) MarshalJSONBytes() ([]byte, error) {
	mxj.XMLEscapeChars(false)
	if d.read {
		return nil, io.EOF
	}
	xmlBytes, err := ioutil.ReadAll(d.r)
	if err != nil {
		return nil, err
	}
	d.read = true

	xmap, err := mxj.NewMapXml(xmlBytes, true)
	if err != nil {
		return nil, err
	}
	return xmap.Json()
}

type xmlEncoder struct {
	w io.Writer
}

func (e xmlEncoder) UnmarshalJSONBytes(jsonBytes []byte, color, pretty bool) error {
	var tmp interface{}
	err := json.Unmarshal(jsonBytes, &tmp)
	if err != nil {
		return err
	}
	if reflect.ValueOf(tmp).Kind() != reflect.Map {
		newObj := map[string]interface{}{"root": tmp}
		jsonBytes, err = json.Marshal(newObj)
		if err != nil {
			return err
		}
	}
	out, err := internalEncode(e, jsonBytes, color, pretty)
	if err != nil {
		return err
	}
	fmt.Fprintln(e.w, string(out))
	return nil
}

func (xmlEncoder) unmarshalJSONBytes(jsonBytes []byte) ([]byte, error) {
	mxj.XMLEscapeChars(true)
	xmap, err := mxj.NewMapJson(jsonBytes)
	if err != nil {
		return nil, err
	}
	return xmap.Xml()
}

func (xmlEncoder) prettyPrint(xmlBytes []byte) ([]byte, error) {
	xmap, err := mxj.NewMapXml(xmlBytes, true)
	if err != nil {
		return nil, err
	}

	return xmap.XmlIndent("", "  ")
}

func (xmlEncoder) color(xmlBytes []byte) ([]byte, error) {
	var b bytes.Buffer
	if err := quick.Highlight(&b, string(xmlBytes), "xml", ChromaFormatter(), ChromaStyle()); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func init() {
	Register("rss", xmlEncoding{})
	Register("svg", xmlEncoding{})
	Register("wsdl", xmlEncoding{})
	Register("wsf", xmlEncoding{})
	Register("xml", xmlEncoding{})
	Register("xsd", xmlEncoding{})
	Register("xsl", xmlEncoding{})
	Register("xslt", xmlEncoding{})
}
