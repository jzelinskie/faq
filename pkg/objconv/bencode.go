package objconv

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/zeebo/bencode"
)

var (
	_ Encoding = bencodeEncoding{}
	_ Decoder  = &bencodeDecoder{}
	_ Encoder  = &bencodeEncoder{}
)

type bencodeEncoding struct{}

func (bencodeEncoding) NewDecoder(r io.Reader) Decoder {
	return &bencodeDecoder{r, false}
}

func (e bencodeEncoding) NewEncoder(w io.Writer) Encoder {
	return &bencodeEncoder{w}
}

type bencodeDecoder struct {
	r    io.Reader
	read bool
}

func (d *bencodeDecoder) MarshalJSONBytes() ([]byte, error) {
	if d.read {
		return nil, io.EOF
	}
	bencodeBytes, err := ioutil.ReadAll(d.r)
	if err != nil {
		return nil, err
	}
	d.read = true

	var obj interface{}
	err = bencode.DecodeBytes(bencodeBytes, &obj)
	if err != nil {
		return nil, err
	}
	return json.Marshal(obj)
}

type bencodeEncoder struct {
	w io.Writer
}

func (e bencodeEncoder) UnmarshalJSONBytes(jsonBytes []byte, color, pretty bool) error {
	out, err := internalEncode(e, jsonBytes, color, pretty)
	if err != nil {
		return err
	}
	fmt.Fprintln(e.w, string(out))
	return nil
}

func (bencodeEncoder) unmarshalJSONBytes(jsonBytes []byte) ([]byte, error) {
	decoder := json.NewDecoder(bytes.NewBuffer(jsonBytes))
	decoder.UseNumber()

	var obj interface{}
	err := decoder.Decode(&obj)
	if err != nil {
		return nil, err
	}
	return bencode.EncodeBytes(obj)
}

func (bencodeEncoder) prettyPrint(bencodeBytes []byte) ([]byte, error) { return bencodeBytes, nil }
func (bencodeEncoder) color(bencodeBytes []byte) ([]byte, error)       { return bencodeBytes, nil }

func init() {
	Register("bencode", bencodeEncoding{})
	Register("torrent", bencodeEncoding{})
}
