#include <stdio.h>

#include <jv.h>
#include <jq.h>

#include "_cgo_export.h"

static inline void indirectGoCallback(void *data, jv it) {
  goJQErrorHandler((GoUint64)data, it);
}

void set_jq_error_cb_default(jq_state *jq) {
  jq_set_error_cb(jq, NULL, NULL);
}

void register_jq_error_cb(jq_state *jq, GoUint64 id) {
  jq_set_error_cb(jq, indirectGoCallback, (void*)id);
}
