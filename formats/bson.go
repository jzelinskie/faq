package formats

import (
	"encoding/json"

	"github.com/globalsign/mgo/bson"
)

type bsonEncoding struct{}

func (bsonEncoding) MarshalJSONBytes(bsonBytes []byte) ([]byte, error) {
	var obj interface{}
	err := bson.Unmarshal(bsonBytes, &obj)
	if err != nil {
		return nil, err
	}
	return json.Marshal(obj)
}

func (bsonEncoding) UnmarshalJSONBytes(jsonBytes []byte) ([]byte, error) {
	var obj interface{}
	err := json.Unmarshal(jsonBytes, &obj)
	if err != nil {
		return nil, err
	}
	return bson.Marshal(obj)
}

func init() {
	ByName["bson"] = bsonEncoding{}
}
