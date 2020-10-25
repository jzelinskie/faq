FROM golang:1.15-alpine AS build
RUN apk add --no-cache git make jq-dev gcc libc-dev oniguruma-dev bash

RUN go get -u github.com/golang/dep/cmd/...
WORKDIR /go/src/github.com/jzelinskie/faq
COPY . .

RUN go get -u golang.org/x/lint/golint/...
RUN go get -u golang.org/x/tools/cmd/...

ENV GO111MODULE=on
RUN make FAQ_LINK_STATIC=true
RUN make install

FROM scratch
COPY --from=build /usr/local/bin/faq /faq
ENTRYPOINT ["/faq"]
