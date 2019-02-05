package formats

import (
	"bytes"
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
	decoder := json.NewDecoder(bytes.NewBuffer(jsonBytes))
	decoder.UseNumber()

	var obj interface{}
	err := decoder.Decode(&obj)
	if err != nil {
		return nil, err
	}
	return bencode.EncodeBytes(obj)
}

func (bencodeEncoding) PrettyPrint(bencodeBytes []byte) ([]byte, error) { return bencodeBytes, nil }
func (bencodeEncoding) Color(bencodeBytes []byte) ([]byte, error)       { return bencodeBytes, nil }

func init() {
	Register("bencode", bencodeEncoding{})
	Register("torrent", bencodeEncoding{})
}
