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
	"strings"
	"testing"

	"github.com/ashb/jqrepl/jq"
)

func TestJqNewClose(t *testing.T) {
	jq, err := jq.New()

	if err != nil {
		t.Errorf("Error initializing jq_state: %v", err)
	}

	jq.Close()

	// We should be able to safely close multiple times.
	jq.Close()

}

func TestJqCloseRace(t *testing.T) {
	state, err := jq.New()

	if err != nil {
		t.Errorf("Error initializing jq_state: %v", err)
	}

	cIn, _, _ := state.Start(".", jq.JvArray())
	go state.Close()
	go close(cIn)
}

func feedJq(val *jq.Jv, in chan<- *jq.Jv, out <-chan *jq.Jv, errs <-chan error) ([]*jq.Jv, []error) {
	if val == nil {
		close(in)
		in = nil
	}
	outputs := make([]*jq.Jv, 0)
	errors := make([]error, 0)
	for errs != nil && out != nil {
		select {
		case e, ok := <-errs:
			if !ok {
				errs = nil
			} else {
				errors = append(errors, e)
			}
		case o, ok := <-out:
			if !ok {
				out = nil
			} else {
				outputs = append(outputs, o)
			}
		case in <- val:
			// We've sent our input, close the channel to tell Jq we're done
			close(in)
			in = nil
		}
	}
	return outputs, errors
}

func TestStartCompileError(t *testing.T) {
	state, err := jq.New()

	if err != nil {
		t.Errorf("Error initializing jq_state: %v", err)
	}
	defer state.Close()

	const program = "a b"
	cIn, cOut, cErr := state.Start(program, jq.JvArray())
	_, errors := feedJq(nil, cIn, cOut, cErr)

	// JQ might (and currently does) report multiple errors. One of them will
	// contain our input program. Check for that but don't be overly-specific
	// about the string or order of errors

	gotErrors := false
	for _, err := range errors {
		gotErrors = true
		if strings.Contains(err.Error(), program) {
			// t.Pass("Found the error we expected: %#v\n",
			return
		}
	}

	if !gotErrors {
		t.Fatal("Errors were expected but none seen")
	}
	t.Fatal("No error containing the program source found")
}

func TestCompileError(t *testing.T) {
	state, err := jq.New()

	if err != nil {
		t.Errorf("Error initializing jq_state: %v", err)
	}
	defer state.Close()

	const program = "a b"
	errors := state.Compile(program, jq.JvArray())

	// JQ might (and currently does) report multiple errors. One of them will
	// contain our input program. Check for that but don't be overly-specific
	// about the string or order of errors

	gotErrors := false
	for _, err := range errors {
		gotErrors = true
		if strings.Contains(err.Error(), program) {
			// t.Pass("Found the error we expected: %#v\n",
			return
		}
	}

	if !gotErrors {
		t.Fatal("Errors were expected but none seen")
	}
	t.Fatal("No error containing the program source found")
}

func TestCompileGood(t *testing.T) {
	state, err := jq.New()

	if err != nil {
		t.Errorf("Error initializing jq_state: %v", err)
	}
	defer state.Close()

	const program = "."
	errors := state.Compile(program, jq.JvArray())

	// JQ might (and currently does) report multiple errors. One of them will
	// contain our input program. Check for that but don't be overly-specific
	// about the string or order of errors

	if len(errors) != 0 {
		t.Fatal("Expected no errors, got", errors)
	}
}

func TestJqSimpleProgram(t *testing.T) {
	state, err := jq.New()

	if err != nil {
		t.Errorf("Error initializing state_state: %v", err)
	}
	defer state.Close()

	input, err := jq.JvFromJSONString("{\"a\": 123}")
	if err != nil {
		t.Error(err)
	}

	cIn, cOut, cErrs := state.Start(".a", jq.JvArray())
	outputs, errs := feedJq(input, cIn, cOut, cErrs)

	if len(errs) > 0 {
		t.Errorf("Expected no errors, but got %#v", errs)
	}

	if l := len(outputs); l != 1 {
		t.Errorf("Got %d outputs (%#v), expected %d", l, outputs, 1)
	} else if val := outputs[0].ToGoVal(); val != 123 {
		t.Errorf("Got %#v, expected %#v", val, 123)
	}
}

func TestJqNonChannelInterface(t *testing.T) {
	state, err := jq.New()

	if err != nil {
		t.Errorf("Error initializing state_state: %v", err)
	}
	defer state.Close()

	input, err := jq.JvFromJSONString("{\"a\": 123}")
	if err != nil {
		t.Error(err)
	}

	errs := state.Compile(".a", jq.JvArray())
	if errs != nil {
		t.Errorf("Expected no errors, but got %#v", errs)
	}

	outputs, err := state.Execute(input.Copy())
	if err != nil {
		t.Errorf("Expected no error, but got %#v", err)
	}

	if l := len(outputs); l != 1 {
		t.Errorf("Got %d outputs (%#v), expected %d", l, outputs, 1)
	} else if val := outputs[0].ToGoVal(); val != 123 {
		t.Errorf("Got %#v, expected %#v", val, 123)
	}
}

func TestJqRuntimeError(t *testing.T) {
	state, err := jq.New()

	if err != nil {
		t.Errorf("Error initializing state_state: %v", err)
	}
	defer state.Close()

	input, err := jq.JvFromJSONString(`{"a": 123}`)
	if err != nil {
		t.Error(err)
	}

	cIn, cOut, cErrs := state.Start(".[0]", jq.JvArray())
	_, errors := feedJq(input, cIn, cOut, cErrs)

	if l := len(errors); l != 1 {
		t.Errorf("Got %d errors (%#v), expected %d", l, errors, 1)
	}
}
