package formats

import "github.com/clbanning/mxj"

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

func init() {
	ByName["xml"] = xmlEncoding{}
	ByName["svg"] = xmlEncoding{}
}
