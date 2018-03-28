package main

import (
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Azure/draft/pkg/linguist"
	"github.com/BurntSushi/toml"
	"github.com/clbanning/mxj"
	"github.com/ghodss/yaml"
	"github.com/globalsign/mgo/bson"
	"github.com/zeebo/bencode"
)

type unmarshaler func([]byte) (interface{}, error)

var formats = map[string]unmarshaler{
	"pem":     pemUnmarshal,
	"bencode": bencodeUnmarshal,
	"bson":    bsonUnmarshal,
	"json":    jsonUnmarshal,
	"toml":    tomlUnmarshal,
	"xml":     xmlUnmarshal,
	"yaml":    yamlUnmarshal,
}

var aliases = map[string]string{
	"javascript": "json",
	"lisp":       "sexp",
}

// extensions are used to override Linguist's auto-detect by mapping file-type
// extensions to supported formats.
var extensions = map[string]string{
	"bson":    "bson",
	"pem":     "pem",
	"toml":    "toml",
	"torrent": "bencode",
}

func detectFormat(fileBytes []byte, path string) string {
	if ext := filepath.Ext(path); ext != "" {
		if format, ok := extensions[ext[1:]]; ok {
			return format
		}
	}

	format := strings.ToLower(linguist.LanguageByContents(fileBytes, linguist.LanguageHints(path)))
	if alias, ok := aliases[format]; ok {
		format = alias
	} else if format == "coq" {
		// This is what linguist says when it has no idea what it's talking about.
		// For now, just fallback to JSON.
		format = "json"
	}
	return format
}

func unmarshal(name string, contents []byte) (interface{}, error) {
	fn, ok := formats[name]
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

func xmlUnmarshal(fileBytes []byte) (interface{}, error) {
	return mxj.NewMapXml(fileBytes, true)
}

func tomlUnmarshal(fileBytes []byte) (interface{}, error) {
	var obj interface{}
	err := toml.Unmarshal(fileBytes, &obj)
	if err != nil {
		return nil, err
	}
	return obj, nil
}

func bsonUnmarshal(fileBytes []byte) (interface{}, error) {
	var obj interface{}
	err := bson.Unmarshal(fileBytes, &obj)
	if err != nil {
		return nil, err
	}
	return obj, nil
}

func pemUnmarshal(fileBytes []byte) (interface{}, error) {
	block, _ := pem.Decode(fileBytes)
	switch {
	case block == nil:
		return nil, errors.New("failed to decode pem")
	case block.Type == "PUBLIC KEY":
		key, _ := x509.ParsePKIXPublicKey(block.Bytes)
		return json.Marshal(key)
	}
	return nil, errors.New("unknown pem format")
}
