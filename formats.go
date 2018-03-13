package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/zeebo/bencode"
)

type unmarshaler func([]byte) (interface{}, error)

var formats = map[string]unmarshaler{
	"bencode": bencodeUnmarshal,
	"json":    jsonUnmarshal,
	"yaml":    yamlUnmarshal,
}

func unmarshal(name string, contents []byte) (interface{}, error) {
	fn, ok := formats[strings.ToLower(name)]
	if !ok {
		return nil, fmt.Errorf("no supported format found named %s", name)
	}
	return fn(contents)
}

func bencodeUnmarshal(fileBytes []byte) (interface{}, error) {
	var obj interface{}
	err := bencode.DecodeBytes(fileBytes, &obj)
	if err != nil {
		return nil, err
	}
	return obj, nil
}

func jsonUnmarshal(fileBytes []byte) (interface{}, error) {
	var obj interface{}
	err := json.Unmarshal(fileBytes, &obj)
	if err != nil {
		return nil, err
	}
	return obj, nil
}

func yamlUnmarshal(fileBytes []byte) (interface{}, error) {
	var obj interface{}
	err := yaml.Unmarshal(fileBytes, &obj)
	if err != nil {
		return nil, err
	}
	return obj, nil
}
