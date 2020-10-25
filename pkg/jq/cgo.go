package jq

/*
#include <jq.h>

// This declares a layer of indirection for calling function pointers.
// See https://github.com/golang/go/wiki/cgo#function-pointer-callbacks
void errorCallback(unsigned long long, jv);
void gojq_error_cb(void *data, jv jv) {
  errorCallback((unsigned long long)data, jv);
};

// This sets the go_jq_error_cb, casting the id into a void*.
// This has to be done in C because Go will not cast C.ulonglong into
// unsafe.Pointer (the type that represents void*).
void gojq_set_error_cb(jq_state *jq, unsigned long long id) {
	jq_set_error_cb(jq, (jq_msg_cb)gojq_error_cb, (void*)id);
};

// This could be called directly from Go, but the name clarifies the intent.
void gojq_reset_error_cb(jq_state *jq) {
	jq_set_error_cb(jq, NULL, NULL);
};
*/
import "C"
