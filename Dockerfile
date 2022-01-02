# syntax=docker/dockerfile:1

ARG GO_VERSION=1.17
ARG OSXCROSS_VERSION=11.3
ARG GORELEASER_XX_VERSION=1.2.5

FROM --platform=$BUILDPLATFORM crazymax/osxcross:${OSXCROSS_VERSION} AS osxcross
FROM --platform=$BUILDPLATFORM crazymax/goreleaser-xx:${GORELEASER_XX_VERSION} AS goreleaser-xx
FROM --platform=$BUILDPLATFORM crazymax/goxx:${GO_VERSION} AS base
COPY --from=osxcross /osxcross /osxcross
COPY --from=goreleaser-xx / /
RUN apt-get update && apt-get install --no-install-recommends -y git
WORKDIR /src

FROM base AS vendored
RUN --mount=type=bind,target=.,rw \
  --mount=type=cache,target=/go/pkg/mod \
  go mod tidy && go mod download

FROM base AS lint
RUN apt-get install --no-install-recommends -y gcc libc6-dev libjq-dev libonig-dev
RUN go install golang.org/x/lint/golint@latest
RUN --mount=type=bind,target=. \
  --mount=type=cache,target=/root/.cache \
  golint ./...

FROM vendored AS test
RUN apt-get install --no-install-recommends -y gcc libc6-dev libjq-dev libonig-dev
RUN --mount=type=bind,target=. \
  --mount=type=cache,target=/go/pkg/mod \
  --mount=type=cache,target=/root/.cache \
  go test -v -coverprofile=/tmp/coverage.txt -covermode=atomic -race ./...

FROM scratch AS test-coverage
COPY --from=test /tmp/coverage.txt /coverage.txt

FROM vendored AS build
ARG TARGETPLATFORM
ENV CGO_ENABLED=1
ENV OSXCROSS_MP_INC=1
RUN --mount=type=cache,sharing=private,target=/var/cache/apt \
  --mount=type=cache,sharing=private,target=/var/lib/apt/lists \
  goxx-apt-get install -y binutils gcc g++ pkg-config libjq-dev libonig-dev
RUN goxx-macports --static install jq
RUN --mount=type=bind,target=/src,rw \
  --mount=type=cache,target=/root/.cache \
  --mount=target=/go/pkg/mod,type=cache \
  --mount=type=secret,id=GITHUB_TOKEN \
  GITHUB_TOKEN=$(cat /run/secrets/GITHUB_TOKEN) goreleaser-xx --debug \
    --config=".goreleaser.yml" \
    --name="faq" \
    --dist="/out" \
    --main="./cmd/faq" \
    --ldflags="-s -w -X 'github.com/jzelinskie/faq/pkg/version.Version={{.Version}}'" \
    --tags="netgo" \
    --files="LICENSE" \
    --files="README.md"

FROM scratch AS artifacts
COPY --from=build /out/*.tar.gz /
COPY --from=build /out/*.zip /

FROM scratch
COPY --from=build /usr/local/bin/faq /faq
ENTRYPOINT ["/faq"]
