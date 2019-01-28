package flagutil

import (
	"fmt"
	"strings"

	"github.com/spf13/pflag"
)

var (
	// validate these types implement the pflag.Value interface at compile time
	_ pflag.Value = &KwargStringFlag{}
	_ pflag.Value = &PositionalArgStringFlag{}
)

// KwargStringFlag implements pflag.Value that accepts key=value pairs and can be
// passed multiple times to support multiple pairs.
type KwargStringFlag struct {
	value   *map[string][]byte
	changed bool
}

// NewKwargStringFlag returns an initialized KwargStringFlag
func NewKwargStringFlag() *KwargStringFlag {
	return &KwargStringFlag{
		value: new(map[string][]byte),
	}
}

// AsMap returns a map of key value pairs passed via the flag
func (f *KwargStringFlag) AsMap() map[string][]byte {
	if f.value != nil {
		return *f.value
	}
	return nil
}

// String implements pflag.Value
func (f *KwargStringFlag) String() string {
	return fmt.Sprintf("%v", f.AsMap())
}

// Set implements pflag.Value
func (f *KwargStringFlag) Set(input string) error {
	p, err := parseKeyValuePair(input)
	if err != nil {
		return err
	}
	if !f.changed {
		// first time set is called so assign to *f.value
		*f.value = map[string][]byte{p.Key: p.Value}
	} else {
		// after first time just update existing *f.value
		(*f.value)[p.Key] = p.Value
	}
	f.changed = true
	return nil
}

// Type implements pflag.Value
func (f *KwargStringFlag) Type() string {
	return "key=value"
}

type pair struct {
	Key   string
	Value []byte
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
	value := []byte(pairSplit[1])
	return &pair{key, value}, nil
}

// PositionalArgStringFlag implements pflag.Value that accepts a single string value
// and stores the value in a list of values. The flag can be specified multiple
// times to add more items the list of values.
type PositionalArgStringFlag struct {
	value [][]byte
}

// NewPositionalArgStringFlag returns an initialized PositionalArgStringFlag
func NewPositionalArgStringFlag() *PositionalArgStringFlag {
	return &PositionalArgStringFlag{}
}

// AsSlice returns a slice of values passed via the flag
func (f *PositionalArgStringFlag) AsSlice() [][]byte {
	return f.value
}

// String implements pflag.Value
func (f *PositionalArgStringFlag) String() string {
	var values []string
	for _, value := range f.value {
		values = append(values, string(value))
	}
	return fmt.Sprintf("%q", values)
}

// Set implements pflag.Value
func (f *PositionalArgStringFlag) Set(input string) error {
	f.value = append(f.value, []byte(input))
	return nil
}

// Type implements pflag.Value
func (f *PositionalArgStringFlag) Type() string {
	return "positionalArg"
}
