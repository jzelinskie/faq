# syntax=docker/dockerfile:1-labs

ARG GO_VERSION=1.17
ARG OSXCROSS_VERSION=11.3
ARG GORELEASER_XX_VERSION=1.2.5
ARG JQ_VERSION="jq-1.6"

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

FROM vendored AS libjq-linux
RUN apt-get install -y autoconf automake flex libtool
ARG TARGETPLATFORM
RUN goxx-apt-get install -y binutils gcc pkg-config
WORKDIR /usr/local/src/jq
ARG JQ_VERSION
RUN <<EOT
set -e
git clone --depth 1 --recurse-submodules --shallow-submodules -b $JQ_VERSION https://github.com/stedolan/jq.git .
HOST_TRIPLE=$(. goxx-env && echo $GOXX_TRIPLE)
BUILD_TRIPLE=$(TARGETPLATFORM= . goxx-env && echo $GOXX_TRIPLE)
autoreconf -fi
CC="$HOST_TRIPLE-gcc" ./configure \
  --prefix=/usr/$HOST_TRIPLE \
  --host=$HOST_TRIPLE \
  --build=$BUILD_TRIPLE \
  --target=$BUILD_TRIPLE \
  --disable-maintainer-mode \
  --disable-docs \
  --enable-all-static \
  --with-oniguruma
make
make -j$(nproc) install DESTDIR="/out"
EOT

FROM scratch AS libjq-dummy
WORKDIR /out

FROM libjq-dummy AS libjq-windows
FROM libjq-dummy AS libjq-darwin
FROM libjq-${TARGETOS} AS libjq

FROM vendored AS build
COPY --from=libjq /out /
ARG TARGETPLATFORM
ENV CGO_ENABLED=1
ENV OSXCROSS_MP_INC=1
RUN goxx-apt-get install -y binutils gcc pkg-config
RUN goxx-macports --static install jq
RUN --mount=type=bind,target=/src,rw \
  --mount=type=cache,target=/root/.cache \
  --mount=target=/go/pkg/mod,type=cache \
  --mount=type=secret,id=GITHUB_TOKEN <<EOT
set -e
EXTLDFLAGS="-v"
if [ "$(. goxx-env && echo $GOOS)" = "linux" ]; then
  EXTLDFLAGS="$EXTLDFLAGS -static"
  export CGO_CFLAGS="-lm -g -O2"
  export CGO_LDFLAGS="-lm -g -O2"
fi
GITHUB_TOKEN=$(cat /run/secrets/GITHUB_TOKEN) goreleaser-xx --debug \
  --config=".goreleaser.yml" \
  --name="faq" \
  --dist="/out" \
  --main="./cmd/faq" \
  --ldflags="-s -w -X 'github.com/jzelinskie/faq/internal/version.Version={{.Version}}' -extldflags '$EXTLDFLAGS'" \
  --tags="netgo" \
  --files="LICENSE" \
  --files="README.md"
EOT

FROM scratch AS artifacts
COPY --from=build /out/*.tar.gz /
COPY --from=build /out/*.zip /

FROM scratch
COPY --from=build /usr/local/bin/faq /faq
ENTRYPOINT ["/faq"]
