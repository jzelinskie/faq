package pflagutil

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/pflag"
)

// Enforce that these types implement the pflag.Value interface at compile time.
var (
	_ pflag.Value = &KwargStringFlag{}
	_ pflag.Value = &KwargJSONFlag{}
	_ pflag.Value = &PositionalArgStringFlag{}
	_ pflag.Value = &PositionalArgJSONFlag{}
)

// KwargJSONFlag implements pflag.Value that accepts key=value pairs and can be
// passed multiple times to support multiple pairs.
type KwargJSONFlag struct {
	value   *map[string]interface{}
	changed bool
}

// NewKwargJSONFlag returns an initialized KwargJSONFlag
func NewKwargJSONFlag(p *map[string]interface{}) *KwargJSONFlag {
	return &KwargJSONFlag{
		value: p,
	}
}

// String implements pflag.Value
func (f *KwargJSONFlag) String() string {
	return fmt.Sprintf("%v", *f.value)
}

// Set implements pflag.Value
func (f *KwargJSONFlag) Set(input string) error {
	p, err := parseKeyValuePair(input)
	if err != nil {
		return err
	}
	var value interface{}
	err = json.Unmarshal([]byte(p.Value), &value)
	if err != nil {
		return fmt.Errorf("unable to decode kwarg %s=%s as JSON arg: %v", p.Key, p.Value, err)
	}
	if !f.changed {
		// first time set is called so assign to *f.value
		*f.value = map[string]interface{}{p.Key: value}
	} else {
		// after first time just update existing *f.value
		(*f.value)[p.Key] = value
	}
	f.changed = true
	return nil
}

// Type implements pflag.Value
func (f *KwargJSONFlag) Type() string {
	return "key=<jsonValue>"
}

// KwargStringFlag implements pflag.Value that accepts key=value pairs and can be
// passed multiple times to support multiple pairs.
type KwargStringFlag struct {
	value   *map[string]string
	changed bool
}

// NewKwargStringFlag returns an initialized KwargStringFlag
func NewKwargStringFlag(p *map[string]string) *KwargStringFlag {
	return &KwargStringFlag{
		value: p,
	}
}

// String implements pflag.Value
func (f *KwargStringFlag) String() string {
	return fmt.Sprintf("%v", *f.value)
}

// Set implements pflag.Value
func (f *KwargStringFlag) Set(input string) error {
	p, err := parseKeyValuePair(input)
	if err != nil {
		return err
	}
	if !f.changed {
		// first time set is called so assign to *f.value
		*f.value = map[string]string{p.Key: p.Value}
	} else {
		// after first time just update existing *f.value
		(*f.value)[p.Key] = p.Value
	}
	f.changed = true
	return nil
}

// Type implements pflag.Value
func (f *KwargStringFlag) Type() string {
	return "key=string"
}

type pair struct {
	Key   string
	Value string
}

func parseKeyValuePair(input string) (*pair, error) {
	if input == "" {
		return nil, nil
	}
	pairSplit := strings.SplitN(input, "=", 2)
	if len(pairSplit) == 1 {
		return nil, fmt.Errorf("did not find any key=value pairs in %s)", input)
	}
	key := pairSplit[0]
	value := pairSplit[1]
	return &pair{key, value}, nil
}

// PositionalArgJSONFlag implements pflag.Value that accepts a single string value
// and stores the value in a list of values. The flag can be specified multiple
// times to add more items the list of values.
type PositionalArgJSONFlag struct {
	value *[]interface{}
}

// NewPositionalArgJSONFlag returns an initialized
// PositionalArgJSONFlag
func NewPositionalArgJSONFlag(p *[]interface{}) *PositionalArgJSONFlag {
	return &PositionalArgJSONFlag{
		value: p,
	}
}

// String implements pflag.Value
func (f *PositionalArgJSONFlag) String() string {
	b, err := json.Marshal(*f.value)
	if err != nil {
		return err.Error()
	}
	return string(b)
}

// Set implements pflag.Value
func (f *PositionalArgJSONFlag) Set(input string) error {
	var value interface{}
	err := json.Unmarshal([]byte(input), &value)
	if err != nil {
		return fmt.Errorf("unable to decode arg %s as JSON arg: %v", input, err)
	}
	*f.value = append(*f.value, value)
	return nil
}

// Type implements pflag.Value
func (f *PositionalArgJSONFlag) Type() string {
	return "<jsonValue>"
}

// PositionalArgStringFlag implements pflag.Value that accepts a single string value
// and stores the value in a list of values. The flag can be specified multiple
// times to add more items the list of values.
type PositionalArgStringFlag struct {
	value *[]string
}

// NewPositionalArgStringFlag returns an initialized PositionalArgStringFlag
func NewPositionalArgStringFlag(p *[]string) *PositionalArgStringFlag {
	return &PositionalArgStringFlag{
		value: p,
	}
}

// String implements pflag.Value
func (f *PositionalArgStringFlag) String() string {
	var values []string
	for _, value := range *f.value {
		values = append(values, value)
	}
	return fmt.Sprintf("%q", values)
}

// Set implements pflag.Value
func (f *PositionalArgStringFlag) Set(input string) error {
	*f.value = append(*f.value, input)
	return nil
}

// Type implements pflag.Value
func (f *PositionalArgStringFlag) Type() string {
	return "string"
}
