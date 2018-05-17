// Copyright (c) 2017 Jimmy Zelinskie
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
	"testing"

	"github.com/ashb/jqrepl/jq"
	"github.com/cheekybits/is"
)

func TestJvKind(t *testing.T) {
	is := is.New(t)

	cases := []struct {
		*jq.Jv
		jq.JvKind
		string
	}{
		{jq.JvNull(), jq.JV_KIND_NULL, "null"},
		{jq.JvFromString("a"), jq.JV_KIND_STRING, "string"},
	}

	for _, c := range cases {
		defer c.Free()
		is.Equal(c.Kind(), c.JvKind)
		is.Equal(c.Kind().String(), c.string)
	}
}

func TestJvString(t *testing.T) {
	is := is.New(t)

	jv := jq.JvFromString("test")
	defer jv.Free()

	str, err := jv.String()

	is.Equal(str, "test")
	is.NoErr(err)

	i := jv.ToGoVal()

	is.Equal(i, "test")
}

func TestJvStringOnNonStringType(t *testing.T) {
	is := is.New(t)

	// Test that on a non-string value we get a go error, not a C assert
	jv := jq.JvNull()
	defer jv.Free()

	_, err := jv.String()
	is.Err(err)
}

func TestJvFromJSONString(t *testing.T) {
	is := is.New(t)

	jv, err := jq.JvFromJSONString("[]")
	is.NoErr(err)
	is.OK(jv)
	is.Equal(jv.Kind(), jq.JV_KIND_ARRAY)

	jv, err = jq.JvFromJSONString("not valid")
	is.Err(err)
	is.Nil(jv)
}

func TestJvFromFloat(t *testing.T) {
	is := is.New(t)

	jv := jq.JvFromFloat(1.23)
	is.OK(jv)
	is.Equal(jv.Kind(), jq.JV_KIND_NUMBER)
	gv := jv.ToGoVal()
	n, ok := gv.(float64)
	is.True(ok)
	is.Equal(n, float64(1.23))
}

func TestJvFromInterface(t *testing.T) {
	is := is.New(t)

	// Null
	jv, err := jq.JvFromInterface(nil)
	is.NoErr(err)
	is.OK(jv)
	is.Equal(jv.Kind(), jq.JV_KIND_NULL)

	// Boolean
	jv, err = jq.JvFromInterface(true)
	is.NoErr(err)
	is.OK(jv)
	is.Equal(jv.Kind(), jq.JV_KIND_TRUE)

	jv, err = jq.JvFromInterface(false)
	is.NoErr(err)
	is.OK(jv)
	is.Equal(jv.Kind(), jq.JV_KIND_FALSE)

	// Float
	jv, err = jq.JvFromInterface(1.23)
	is.NoErr(err)
	is.OK(jv)
	is.Equal(jv.Kind(), jq.JV_KIND_NUMBER)
	gv := jv.ToGoVal()
	n, ok := gv.(float64)
	is.True(ok)
	is.Equal(n, float64(1.23))

	// Integer
	jv, err = jq.JvFromInterface(456)
	is.NoErr(err)
	is.OK(jv)
	is.Equal(jv.Kind(), jq.JV_KIND_NUMBER)
	gv = jv.ToGoVal()
	n2, ok := gv.(int)
	is.True(ok)
	is.Equal(n2, 456)

	// String
	jv, err = jq.JvFromInterface("test")
	is.NoErr(err)
	is.OK(jv)
	is.Equal(jv.Kind(), jq.JV_KIND_STRING)
	gv = jv.ToGoVal()
	s, ok := gv.(string)
	is.True(ok)
	is.Equal(s, "test")

	jv, err = jq.JvFromInterface([]string{"test", "one", "two"})
	is.NoErr(err)
	is.OK(jv)
	is.Equal(jv.Kind(), jq.JV_KIND_ARRAY)
	gv = jv.ToGoVal()
	is.Equal(gv.([]interface{})[2], "two")

	jv, err = jq.JvFromInterface(map[string]int{"one": 1, "two": 2})
	is.NoErr(err)
	is.OK(jv)
	is.Equal(jv.Kind(), jq.JV_KIND_OBJECT)
	gv = jv.ToGoVal()
	is.Equal(gv.(map[string]interface{})["two"], 2)
}

func TestJvDump(t *testing.T) {
	is := is.New(t)

	jv := jq.JvFromString("test")
	defer jv.Free()

	dump := jv.Copy().Dump(jq.JvPrintNone)

	is.Equal(`"test"`, dump)
	dump = jv.Copy().Dump(jq.JvPrintColour)

	is.Equal([]byte("\x1b[0;32m"+`"test"`+"\x1b[0m"), []byte(dump))
}

func TestJvInvalid(t *testing.T) {
	is := is.New(t)

	jv := jq.JvInvalid()

	is.False(jv.IsValid())

	_, ok := jv.Copy().GetInvalidMessageAsString()
	is.False(ok) // "Expected no Invalid message"

	jv = jv.GetInvalidMessage()
	is.Equal(jv.Kind(), jq.JV_KIND_NULL)
}

func TestJvInvalidWithMessage_string(t *testing.T) {
	is := is.New(t)

	jv := jq.JvInvalidWithMessage(jq.JvFromString("Error message 1"))

	is.False(jv.IsValid())

	msg := jv.Copy().GetInvalidMessage()
	is.Equal(msg.Kind(), jq.JV_KIND_STRING)
	msg.Free()

	str, ok := jv.GetInvalidMessageAsString()
	is.True(ok)
	is.Equal("Error message 1", str)
}

func TestJvInvalidWithMessage_object(t *testing.T) {
	is := is.New(t)

	jv := jq.JvInvalidWithMessage(jq.JvObject())

	is.False(jv.IsValid())

	msg := jv.Copy().GetInvalidMessage()
	is.Equal(msg.Kind(), jq.JV_KIND_OBJECT)
	msg.Free()

	str, ok := jv.GetInvalidMessageAsString()
	is.True(ok)
	is.Equal("{}", str)

}
