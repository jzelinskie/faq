// Copyright (c) 2018 Jimmy Zelinskie
// Copyright (c) 2015 Ash Berlin
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package jq_test

import (
	"fmt"
	"testing"

	"github.com/jzelinskie/faq/jq"
)

func TestJvKind(t *testing.T) {
	table := []struct {
		testName string
		*jq.Jv
		jq.JvKind
		stringKind string
	}{
		{"Null", jq.JvNull(), jq.JvKindNull, "null"},
		{"FromString", jq.JvFromString("a"), jq.JvKindString, "string"},
	}

	for _, tt := range table {
		t.Run(tt.testName, func(t *testing.T) {
			defer tt.Free()
			if tt.Kind() != tt.JvKind {
				t.Errorf("JvKind() got: %v, want: %v", tt.Kind(), tt.JvKind)
			}

			if tt.Kind().String() != tt.stringKind {
				t.Errorf("JvKind().String() got: %s, want: %s", tt.Kind().String(), tt.stringKind)
			}
		})
	}
}

func TestJvString(t *testing.T) {
	jv := jq.JvFromString("test")
	defer jv.Free()

	str, err := jv.String()
	if err != nil {
		t.Errorf("error when converting jv into string: %s", err)
	}
	if str != "test" {
		t.Errorf(`jvFromString("test") got: %s, want: test`, str)
	}

	if jv.ToGoVal() != "test" {
		t.Errorf(`jvFromString("test").ToGoVal() got: %s, want: test`, jv.ToGoVal())
	}
}

func TestJvStringOnNonStringType(t *testing.T) {
	jv := jq.JvNull()
	defer jv.Free()

	if _, err := jv.String(); err == nil {
		t.Errorf("created string from jv null value")
	}
}

func TestJvFromJSONString(t *testing.T) {
	jv, err := jq.JvFromJSONString("[]")
	if err != nil {
		t.Errorf("error when parsing jv from JSON string: %s", err)
	}
	if jv == nil {
		t.Errorf(`nil jv when parsing from JSON string "[]"`)
	}

	if jv.Kind() != jq.JvKindArray {
		t.Errorf(`jv kind is not Array for JSON string "[]"`)
	}

	jv, err = jq.JvFromJSONString("not valid")
	if err == nil {
		t.Errorf("parsing jv succeeded when parsing invalid JSON")
	}
	if jv != nil {
		t.Errorf("jv value was not nil when parsing invalid JSON")
	}
}

func TestJvFromFloat(t *testing.T) {
	const exampleFloat = 1.23

	jv := jq.JvFromFloat(exampleFloat)
	if jv.Kind() != jq.JvKindNumber {
		t.Errorf("jv kind is not Number when initialized by a float")
	}

	gv := jv.ToGoVal()
	n, ok := gv.(float64)
	if !ok {
		t.Errorf("failed to cast jv float to Go float64")
	}
	if n != float64(exampleFloat) {
		t.Errorf("float value casted from jv is not equal to original Go value")
	}
}

func TestJvFromInterface(t *testing.T) {
	// Null
	jv, err := jq.JvFromInterface(nil)
	if err != nil {
		t.Errorf("JvFromInterface() with nil failed to parse: %s", err)
	}
	if jv == nil {
		t.Errorf("JvFromInterface() with nil suceeded, but returned nil")
	}
	if jv.Kind() != jq.JvKindNull {
		t.Errorf("JvFromInterface() with nil did not parse into a JvKindNull")
	}

	// Boolean true
	jv, err = jq.JvFromInterface(true)
	if err != nil {
		t.Errorf("JvFromInterface() with true failed to parse: %s", err)
	}
	if jv == nil {
		t.Errorf("JvFromInterface() with true suceeded, but returned nil")
	}
	if jv.Kind() != jq.JvKindTrue {
		t.Errorf("JvFromInterface() with true did not parse into a JvKindTrue")
	}

	// Boolean false
	jv, err = jq.JvFromInterface(false)
	if err != nil {
		t.Errorf("JvFromInterface() with false failed to parse: %s", err)
	}
	if jv == nil {
		t.Errorf("JvFromInterface() with false suceeded, but returned nil")
	}
	if jv.Kind() != jq.JvKindFalse {
		t.Errorf("JvFromInterface() with false did not parse into a JvKindFalse")
	}

	// Float
	jv, err = jq.JvFromInterface(1.23)
	if err != nil {
		t.Errorf("JvFromInterface() with a float failed to parse: %s", err)
	}
	if jv == nil {
		t.Errorf("JvFromInterface() with a float suceeded, but returned nil")
	}
	if jv.Kind() != jq.JvKindNumber {
		t.Errorf("JvFromInterface() with a float did not parse into a JvKindNumber")
	}
	gv := jv.ToGoVal()
	n, ok := gv.(float64)
	if !ok {
		t.Errorf("JVFromInterface() with a float failed to cast back to Go value")
	}
	if n != float64(1.23) {
		t.Errorf("JVFromInterface() with a float casted back is not equal to original Go value")
	}

	// Integer
	jv, err = jq.JvFromInterface(456)
	if err != nil {
		t.Errorf("JvFromInterface() with an integer failed to parse: %s", err)
	}
	if jv == nil {
		t.Errorf("JvFromInterface() with an integer suceeded, but returned nil")
	}
	if jv.Kind() != jq.JvKindNumber {
		t.Errorf("JvFromInterface() with an integer did not parse into a JvKindNumber")
	}
	gv = jv.ToGoVal()
	n2, ok := gv.(int)
	if !ok {
		t.Errorf("JVFromInterface() with an integer failed to cast back to Go value")
	}
	if n2 != 456 {
		t.Errorf("JVFromInterface() with an integer casted back is not equal to original Go value")
	}

	// String
	jv, err = jq.JvFromInterface("test")
	if err != nil {
		t.Errorf("JvFromInterface() with a string failed to parse: %s", err)
	}
	if jv == nil {
		t.Errorf("JvFromInterface() with a string suceeded, but returned nil")
	}
	if jv.Kind() != jq.JvKindString {
		t.Errorf("JvFromInterface() with a string did not parse into a JvKindString")
	}
	gv = jv.ToGoVal()
	s, ok := gv.(string)
	if !ok {
		t.Errorf("JVFromInterface() with a string failed to cast back to Go value")
	}
	if s != "test" {
		t.Errorf("JVFromInterface() with a string casted back is not equal to original Go value")
	}

	jv, err = jq.JvFromInterface([]string{"test", "one", "two"})
	if err != nil {
		t.Errorf("JvFromInterface() with an array failed to parse: %s", err)
	}
	if jv == nil {
		t.Errorf("JvFromInterface() with an array suceeded, but returned nil")
	}
	if jv.Kind() != jq.JvKindArray {
		t.Errorf("JvFromInterface() with an array did not parse into a JvKindArray")
	}
	gv = jv.ToGoVal()
	if gv.([]interface{})[2] != "two" {
		t.Errorf("JVFromInterface() with an array casted back is not equal to original Go value")
	}

	jv, err = jq.JvFromInterface(map[string]int{"one": 1, "two": 2})
	if err != nil {
		t.Errorf("JvFromInterface() with a map failed to parse: %s", err)
	}
	if jv == nil {
		t.Errorf("JvFromInterface() with a map suceeded, but returned nil")
	}
	if jv.Kind() != jq.JvKindObject {
		t.Errorf("JvFromInterface() with a map did not parse into a JvKindObject")
	}
	gv = jv.ToGoVal()
	if gv.(map[string]interface{})["two"] != 2 {
		t.Errorf("JVFromInterface() with a map casted back is not equal to original Go value")
	}
}

func TestJvDump(t *testing.T) {
	table := []struct {
		input  string
		flags  jq.JvPrintFlags
		output string
	}{
		{"test", jq.JvPrintNone, `"test"`},
		{"test", jq.JvPrintColour, "\x1b[0;32m" + `"test"` + "\x1b[0m"},
	}
	for _, tt := range table {
		t.Run(fmt.Sprintf("%s-%b", tt.input, tt.flags), func(t *testing.T) {
			jv := jq.JvFromString(tt.input)
			defer jv.Free()

			dump := jv.Copy().Dump(tt.flags)
			if dump != tt.output {
				t.Errorf("dump not equal to expected got: %#v want: %#v", dump, tt.output)
			}
		})
	}
}

func TestJvInvalid(t *testing.T) {
	jv := jq.JvInvalid()
	if jv.IsValid() == true {
		t.Errorf("IsValid() returned true for JvInvalid()")
	}

	if _, ok := jv.Copy().GetInvalidMessageAsString(); ok {
		t.Errorf("GetInvalidMessageAsString() returned string for no value")

	}

	if jv.GetInvalidMessage().Kind() != jq.JvKindNull {
		t.Errorf("GetInvalidMessage().Kind() returned a kind other than JvKindNull")
	}
}

func TestJvInvalidWithMessage_string(t *testing.T) {
	errMsg := "Error message 1"
	jv := jq.JvInvalidWithMessage(jq.JvFromString(errMsg))
	if jv.IsValid() == true {
		t.Errorf("IsValid() returned true for JvInvalidWithMessage()")
	}

	msg := jv.Copy().GetInvalidMessage()
	if msg.Kind() != jq.JvKindString {
		t.Errorf("JvInvalidWithMessage().GetInvalidMessage().Kind() returned a kind other than JvKindString")
	}
	msg.Free()

	str, ok := jv.GetInvalidMessageAsString()
	if !ok {
		t.Errorf("JvInvalidWithMessage().JvGetInvalidMessageAsString() is not ok")
	}
	if str != errMsg {
		t.Errorf("JvInvalidWithMessage().JvGetInvalidMessageAsString() did not return original error message")
	}
}

func TestJvInvalidWithMessage_object(t *testing.T) {
	jv := jq.JvInvalidWithMessage(jq.JvObject())
	if jv.IsValid() == true {
		t.Errorf("IsValid() returned true for JvInvalidWithMessage()")
	}

	msg := jv.Copy().GetInvalidMessage()
	if msg.Kind() != jq.JvKindObject {
		t.Errorf("JvInvalidWithMessage().GetInvalidMessage().Kind() returned a kind other than JvKindObject")
	}
	msg.Free()

	str, ok := jv.GetInvalidMessageAsString()
	if !ok {
		t.Errorf("JvInvalidWithMessage().JvGetInvalidMessageAsString() is not ok")
	}
	if str != "{}" {
		t.Errorf(`JvInvalidWithMessage().JvGetInvalidMessageAsString() did not return "{}"`)
	}
}
