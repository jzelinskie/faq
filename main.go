package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/ashb/jqrepl/jq"
	"github.com/zeebo/bencode"
)

func main() {
	prog, err := jq.New()
	if err != nil {
		panic(err)
	}
	defer prog.Close()

	if len(os.Args) != 2 {
		panic("args != 2")
	}

	fileBytes, err := ioutil.ReadFile(os.ExpandEnv(os.Args[1]))
	if err != nil {
		panic(err)
	}

	var bencodeObject interface{}
	err = bencode.DecodeBytes(fileBytes, &bencodeObject)
	if err != nil {
		panic(err)
	}

	jsonBytes, err := json.Marshal(bencodeObject)
	if err != nil {
		panic(err)
	}

	fileJv, err := jq.JvFromJSONBytes(jsonBytes)
	if err != nil {
		panic(err)
	}

	emptyArgs := jq.JvArray()
	errs := prog.Compile("", emptyArgs)
	for _, err := range errs {
		if err != nil {
			panic(err)
		}
	}

	resultJvs, err := prog.Execute(fileJv)
	if err != nil {
		panic(err)
	}

	for _, resultJv := range resultJvs {
		fmt.Println(resultJv.Dump(jq.JvPrintPretty | jq.JvPrintSpace1 | jq.JvPrintColour))
	}
}
