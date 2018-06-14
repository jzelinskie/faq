package formats

import (
	"encoding/json"

	"github.com/zeebo/bencode"
)

type bencodeEncoding struct{}

func (bencodeEncoding) MarshalJSONBytes(bencodedBytes []byte) ([]byte, error) {
	var obj interface{}
	err := bencode.DecodeBytes(bencodedBytes, &obj)
	if err != nil {
		return nil, err
	}
	return json.Marshal(obj)
}

func (bencodeEncoding) UnmarshalJSONBytes(jsonBytes []byte) ([]byte, error) {
	var obj interface{}
	err := json.Unmarshal(jsonBytes, &obj)
	if err != nil {
		return nil, err
	}
	return bencode.EncodeBytes(obj)
}

func (bencodeEncoding) Raw(bencodeBytes []byte) ([]byte, error)         { return bencodeBytes, nil }
func (bencodeEncoding) PrettyPrint(bencodeBytes []byte) ([]byte, error) { return bencodeBytes, nil }
func (bencodeEncoding) Color(bencodeBytes []byte) ([]byte, error)       { return bencodeBytes, nil }

func init() {
	ByName["bencode"] = bencodeEncoding{}
	ByName["torrent"] = bencodeEncoding{}
}
