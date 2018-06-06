package formats

import (
	"bytes"

	"github.com/alecthomas/chroma/quick"
	"github.com/clbanning/mxj"
)

type xmlEncoding struct{}

func (xmlEncoding) MarshalJSONBytes(xmlBytes []byte) ([]byte, error) {
	xmap, err := mxj.NewMapXml(xmlBytes, true)
	if err != nil {
		return nil, err
	}
	return xmap.Json()
}

func (xmlEncoding) UnmarshalJSONBytes(jsonBytes []byte) ([]byte, error) {
	xmap, err := mxj.NewMapJson(jsonBytes)
	if err != nil {
		return nil, err
	}
	return xmap.Xml()
}

func (xmlEncoding) Raw(xmlBytes []byte) ([]byte, error) { return xmlBytes, nil }
func (xmlEncoding) Color(xmlBytes []byte) ([]byte, error) {
	var b bytes.Buffer
	if err := quick.Highlight(&b, string(xmlBytes), "xml", ChromaFormatter(), ChromaStyle()); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func init() {
	ByName["rss"] = xmlEncoding{}
	ByName["svg"] = xmlEncoding{}
	ByName["wsdl"] = xmlEncoding{}
	ByName["wsf"] = xmlEncoding{}
	ByName["xml"] = xmlEncoding{}
	ByName["xsd"] = xmlEncoding{}
	ByName["xsl"] = xmlEncoding{}
	ByName["xslt"] = xmlEncoding{}
}
