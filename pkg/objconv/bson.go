package objconv

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/globalsign/mgo/bson"
)

var (
	_ Encoding = bsonEncoding{}
	_ Decoder  = &bsonDecoder{}
	_ Encoder  = &bsonEncoder{}
)

type bsonEncoding struct{}

func (bsonEncoding) NewDecoder(r io.Reader) Decoder {
	return &bsonDecoder{r, false}
}

func (e bsonEncoding) NewEncoder(w io.Writer) Encoder {
	return &bsonEncoder{w}
}

type bsonDecoder struct {
	r    io.Reader
	read bool
}

func (d *bsonDecoder) MarshalJSONBytes() ([]byte, error) {
	if d.read {
		return nil, io.EOF
	}
	bsonBytes, err := ioutil.ReadAll(d.r)
	if err != nil {
		return nil, err
	}
	d.read = true

	var obj interface{}
	err = bson.Unmarshal(bsonBytes, &obj)
	if err != nil {
		return nil, err
	}
	return json.Marshal(obj)
}

type bsonEncoder struct {
	w io.Writer
}

func (e bsonEncoder) UnmarshalJSONBytes(jsonBytes []byte, color, pretty bool) error {
	out, err := internalEncode(e, jsonBytes, color, pretty)
	if err != nil {
		return err
	}
	fmt.Fprintln(e.w, string(out))
	return nil
}

func (bsonEncoder) unmarshalJSONBytes(jsonBytes []byte) ([]byte, error) {
	var obj interface{}
	err := json.Unmarshal(jsonBytes, &obj)
	if err != nil {
		return nil, err
	}
	return bson.Marshal(obj)
}

func (bsonEncoder) prettyPrint(bsonBytes []byte) ([]byte, error) { return bsonBytes, nil }
func (bsonEncoder) color(bsonBytes []byte) ([]byte, error)       { return bsonBytes, nil }

func init() {
	Register("bson", bsonEncoding{})
}
