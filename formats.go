package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/zeebo/bencode"
)

type unmarshaler func([]byte) (interface{}, error)

var formats = map[string]unmarshaler{
	"bencode": bencodeUnmarshal,
	"json":    jsonUnmarshal,
}

func unmarshal(name string, contents []byte) (interface{}, error) {
	fn, ok := formats[strings.ToLower(name)]
	if !ok {
		return nil, fmt.Errorf("no supported format found named %s", name)
	}
	return fn(contents)
}

func bencodeUnmarshal(fileBytes []byte) (interface{}, error) {
	var bencodeObject interface{}
	err := bencode.DecodeBytes(fileBytes, &bencodeObject)
	if err != nil {
		return nil, err
	}
	return bencodeObject, nil
}

func jsonUnmarshal(fileBytes []byte) (interface{}, error) {
	var jsonObject interface{}
	err := json.Unmarshal(fileBytes, &jsonObject)
	if err != nil {
		return nil, err
	}
	return jsonObject, nil
}
