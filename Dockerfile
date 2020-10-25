FROM golang:1.15-alpine AS build
RUN apk add --no-cache git make jq-dev gcc libc-dev oniguruma-dev bash

WORKDIR /go/src/github.com/jzelinskie/faq
COPY . .

RUN make install FAQ_LINK_STATIC=true

FROM scratch
COPY --from=build /usr/local/bin/faq /faq
ENTRYPOINT ["/faq"]
