FROM quay.io/jzelinskie/golang:1.12rc2-alpine-edge
RUN apk add --no-cache git jq-dev gcc libc-dev oniguruma-dev bash

RUN go get -u github.com/golang/dep/cmd/...
WORKDIR /go/src/github.com/jzelinskie/faq
COPY . .

RUN go get -u golang.org/x/lint/golint/...
RUN go get -u golang.org/x/tools/cmd/...

ENV GO111MODULE=on
RUN /go/src/github.com/jzelinskie/faq/test.sh
RUN go install -v --ldflags '-s -w -linkmode external -extldflags "-v -static"' github.com/jzelinskie/faq
