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

// Package jq provides Go bindings for libjq providing a streaming filter of
// JSON documents.
//
// This package provides a thin layer on top of stedolan's libjq -- it would
// likely be helpful to read through the wiki pages about it:
//
// jv: the JSON value type https://github.com/stedolan/jq/wiki/C-API:-jv
//
// libjq: https://github.com/stedolan/jq/wiki/C-API:-libjq
//
// This package has been forked from github.com/ashb/jqrepl to include static
// linking and more idiomatic Go.
package jq

/*
#cgo LDFLAGS: -ljq -lonig

#include <jq.h>
#include <jv.h>

#include <stdlib.h>

void install_jq_error_cb(jq_state *jq, unsigned long long id);
*/
import "C"
import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"unsafe"
)

// Jq encapsulates the state needed to interface with the libjq C library
type Jq struct {
	_state       *C.struct_jq_state
	errorStoreID uint64
	running      sync.WaitGroup
}

// New initializes a new JQ object and the underlying C library.
func New() (*Jq, error) {
	jq := new(Jq)

	var err error
	jq._state, err = C.jq_init()

	if err != nil {
		return nil, err
	} else if jq == nil {
		return nil, errors.New("jq_init returned nil -- out of memory?")
	}

	return jq, nil
}

// Close the handle to libjq and free C resources.
//
// If Start() has been called this will block until the input Channel it
// returns has been closed.
func (jq *Jq) Close() {
	// If the goroutine from Start() is running we need to make sure it finished cleanly
	// Wait until we aren't running before freeing C things.
	//
	jq.running.Wait()
	if jq._state != nil {
		C.jq_teardown(&jq._state)
		jq._state = nil
	}
	if jq.errorStoreID != 0 {
		globalErrorChannels.Delete(jq.errorStoreID)
		jq.errorStoreID = 0
	}
}

// We cant pass many things over the Go/C boundary, so instead of passing the error channel we pass an opaque indentifier (a 64bit int as it turns out) and use that to look up in a global variable
type errorLookupState struct {
	sync.RWMutex
	idCounter uint64
	channels  map[uint64]chan<- error
}

func (e *errorLookupState) Add(c chan<- error) uint64 {
	newID := atomic.AddUint64(&e.idCounter, 1)
	e.RWMutex.Lock()
	defer e.RWMutex.Unlock()
	e.channels[newID] = c
	return newID
}

func (e *errorLookupState) Get(id uint64) chan<- error {
	e.RWMutex.RLock()
	defer e.RWMutex.RUnlock()
	c, ok := e.channels[id]
	if !ok {
		panic(fmt.Sprintf("Tried to get error channel #%d out of store but it wasn't there!", id))
	}
	return c
}

func (e *errorLookupState) Delete(id uint64) {
	e.RWMutex.Lock()
	defer e.RWMutex.Unlock()
	delete(e.channels, id)
}

// The global state - this also serves to keep the channel in scope by keeping
// a reference to it that the GC can see
var globalErrorChannels = errorLookupState{
	channels: make(map[uint64]chan<- error),
}

//export goLibjqErrorHandler
func goLibjqErrorHandler(id uint64, jv C.jv) {
	ch := globalErrorChannels.Get(id)

	err := _ConvertError(jv)
	ch <- err
}

// Start will compile `program` and return a three channels: input, output and
// error. Sending a jq.Jv* to input cause the program to be run to it and
// one-or-more results returned as jq.Jv* on the output channel, or one or more
// error values sent to the error channel. When you are done sending values
// close the input channel.
//
// args is a list of key/value pairs to bind as variables into the program, and
// must be an array type even if empty. Each element of the array should be an
// object with a "name" and "value" properties. Name should exclude the "$"
// sign. For example this is `[ {"name": "n", "value": 1 } ]` would then be
// `$n` in the programm.
//
// This function is not reentereant -- in that you cannot and should not call
// Start again until you have closed the previous input channel.
//
// If there is a problem compiling the JQ program then the errors will be
// reported on error channel before any input is read so makle sure you account
// for this case.
//
// Any jq.Jv* values passed to the input channel will be owned by the channel.
// If you want to keep them afterwards ensure you Copy() them before passing to
// the channel
func (jq *Jq) Start(program string, args *Jv) (in chan<- *Jv, out <-chan *Jv, errs <-chan error) {
	// Create out two way copy of the channels. We need to be able to recv from
	// input, so need to store the original channel
	cIn := make(chan *Jv)
	cOut := make(chan *Jv)
	cErr := make(chan error)

	// And assign the read/write only versions to the output fars
	in = cIn
	out = cOut
	errs = cErr

	// Before setting up any of the global error handling state, lets check that
	// args is of the right type!
	if args.Kind() != JvKindArray {
		go func() {
			// Take ownership of the inputs
			for jv := range cIn {
				jv.Free()
			}
			cErr <- fmt.Errorf("`args` parameter is of type %s not array", args.Kind().String())
			args.Free()
			close(cOut)
			close(cErr)
		}()
		return
	}

	if jq.errorStoreID != 0 {
		// We might have called Compile
		globalErrorChannels.Delete(jq.errorStoreID)
	}
	jq.errorStoreID = globalErrorChannels.Add(cErr)

	// Because we can't pass a function pointer to an exported Go func we have to
	// call a C function which uses the exported fund for us.
	// https://github.com/golang/go/wiki/cgo#function-variables
	C.install_jq_error_cb(jq._state, C.ulonglong(jq.errorStoreID))

	jq.running.Add(1)
	go func() {

		if jq._Compile(program, args) == false {
			// Even if compile failed follow the contract. Read any inputs and take
			// ownership of them (aka free them)
			//
			// Errors from compile will be sent to the error channel
			for jv := range cIn {
				jv.Free()
			}
		} else {
			for jv := range cIn {
				results, err := jq.Execute(jv)
				for _, result := range results {
					cOut <- result
				}
				if err != nil {
					cErr <- err
				}
			}
		}
		// Once we've read all the inputs close the output to signal to caller that
		// we are done.
		close(cOut)
		close(cErr)
		C.install_jq_error_cb(jq._state, 0)
		jq.running.Done()
	}()

	return
}

// Execute will run the Compiled() program against a single input and return
// the results.
//
// Using this interface directly is not thread-safe -- it is up to the caller to
// ensure that this is not called from two goroutines concurrently.
func (jq *Jq) Execute(input *Jv) (results []*Jv, err error) {
	flags := C.int(0)
	results = make([]*Jv, 0)

	C.jq_start(jq._state, input.jv, flags)
	result := &Jv{C.jq_next(jq._state)}
	for result.IsValid() {
		results = append(results, result)
		result = &Jv{C.jq_next(jq._state)}
	}
	msg, ok := result.GetInvalidMessageAsString()
	if ok {
		// Uncaught jq exception
		// TODO: get file:line position in input somehow.
		err = errors.New(msg)
	}

	return
}

// Compile the program and make it ready to Execute()
//
// Only a single program can be compiled on a Jq object at once. Calling this
// again a second time will replace the current program.
//
// args is a list of key/value pairs to bind as variables into the program, and
// must be an array type even if empty. Each element of the array should be an
// object with a "name" and "value" properties. Name should exclude the "$"
// sign. For example this is `[ {"name": "n", "value": 1 } ]` would then be
// `$n` in the program.
func (jq *Jq) Compile(prog string, args *Jv) (errs []error) {

	// Before setting up any of the global error handling state, lets check that
	// args is of the right type!
	if args.Kind() != JvKindArray {
		args.Free()
		return []error{fmt.Errorf("`args` parameter is of type %s not array", args.Kind().String())}
	}

	cErr := make(chan error)

	if jq.errorStoreID != 0 {
		// We might have called Compile
		globalErrorChannels.Delete(jq.errorStoreID)
	}
	jq.errorStoreID = globalErrorChannels.Add(cErr)

	C.install_jq_error_cb(jq._state, C.ulonglong(jq.errorStoreID))
	defer C.install_jq_error_cb(jq._state, 0)
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		for err := range cErr {
			if err == nil {
				break
			}
			errs = append(errs, err)
		}
		wg.Done()
	}()

	compiled := jq._Compile(prog, args)
	cErr <- nil // Sentinel to break the loop above

	wg.Wait()
	globalErrorChannels.Delete(jq.errorStoreID)
	jq.errorStoreID = 0

	if !compiled && len(errs) == 0 {
		return []error{fmt.Errorf("jq_compile returned error, but no errors were reported. Oops")}
	}
	return errs
}

func (jq *Jq) _Compile(prog string, args *Jv) bool {
	cs := C.CString(prog)
	defer C.free(unsafe.Pointer(cs))

	// If there was an error it will have been sent to errorChannel via the
	// installed error handler
	return C.jq_compile_args(jq._state, cs, args.jv) != 0
}
