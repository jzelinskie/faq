package formats

import (
	"encoding/json"

	"github.com/clbanning/mxj"
)

type xmlEncoding struct{}

func (xmlEncoding) MarshalJSONBytes(xmlBytes []byte) ([]byte, error) {
	xmap, err := mxj.NewMapXml(xmlBytes)
	if err != nil {
		return nil, err
	}
	return xmap.Json()
}

func (xmlEncoding) UnmarshalJSONBytes(jsonBytes []byte) ([]byte, error) {
	var obj interface{}
	err := json.Unmarshal(jsonBytes, &obj)
	if err != nil {
		return nil, err
	}
	return mxj.AnyXml(obj)
}

func init() {
	ByName["xml"] = xmlEncoding{}
}
