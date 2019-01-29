#!/bin/bash

set -e

go mod vendor
go vet $(go list ./...)
diff <(goimports -d $(find . -type f -name '*.go' -not -path "./vendor/*")) <(printf "")
(for d in $(go list ./...); do diff <(golint $d) <(printf "") || exit 1;  done)
