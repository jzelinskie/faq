package main

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Azure/draft/pkg/linguist"
	"github.com/ghodss/yaml"
	"github.com/zeebo/bencode"
)

type unmarshaler func([]byte) (interface{}, error)

var formats = map[string]unmarshaler{
	"bencode": bencodeUnmarshal,
	"json":    jsonUnmarshal,
	"yaml":    yamlUnmarshal,
}

var aliases = map[string]string{
	"javascript": "json",
	"lisp":       "sexp",
}

// extensions are used to override Linguist's auto-detect by mapping file-type
// extensions to supported formats.
var extensions = map[string]string{
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
