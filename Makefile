GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)
FAQ_BIN = faq-$(GOOS)-$(GOARCH)

ifeq ($(GOOS), linux)
INSTALL=install
else
INSTALL=ginstall
endif

FAQ_LINK_STATIC=false
GO_EXT_LD_FLAGS=-v
ifeq ($(FAQ_LINK_STATIC), true)
GO_EXT_LD_FLAGS+= -static
endif

FAQ_VERSION=$(shell git describe --always --abbrev=40 --dirty)
GO_LD_FLAGS=-s -w -X github.com/jzelinskie/faq/pkg/version.Version=$(FAQ_VERSION) -extldflags "$(GO_EXT_LD_FLAGS)"

GO=go
GO_BUILD_ARGS=-v -ldflags '$(GO_LD_FLAGS)' -tags netgo
GO_FILES:=$(shell find . -name '*.go' -type f)

prefix = /usr/local
exec_prefix = $(prefix)
bindir = $(exec_prefix)/bin

install: $(FAQ_BIN)
	mkdir -p $(DESTDIR)$(bindir)
	$(INSTALL) -m 0755 $(FAQ_BIN) $(DESTDIR)$(bindir)/faq

$(FAQ_BIN): $(GO_FILES)
	CGO_ENABLED=1 GOOS=$(GOOS) GOARCH=$(GOARCH) $(GO) build -o $(FAQ_BIN) $(GO_BUILD_ARGS) github.com/jzelinskie/faq/cmd/faq

PHONY: build
build: $(FAQ_BIN)

PHONY: test
test:
	go test ./...

clean:
	rm $(FAQ_BIN)

PHONY: lint
lint:
	golint ./...

all: lint test build
