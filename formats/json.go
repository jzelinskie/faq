package formats

type jsonEncoding struct{}

func (jsonEncoding) MarshalJSONBytes(jsonBytes []byte) ([]byte, error) {
	return jsonBytes, nil
}

func (jsonEncoding) UnmarshalJSONBytes(jsonBytes []byte) ([]byte, error) {
	return jsonBytes, nil
}

func init() {
	ByName["json"] = jsonEncoding{}
	ByName["js"] = jsonEncoding{}
	ByName["javascript"] = jsonEncoding{}
}
