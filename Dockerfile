FROM golang:alpine
RUN apk add --no-cache git jq-dev gcc libc-dev

RUN go get -u github.com/golang/dep/cmd/...
WORKDIR /go/src/github.com/jzelinskie/faq
COPY . .

RUN dep ensure
RUN go install -v github.com/jzelinskie/faq
