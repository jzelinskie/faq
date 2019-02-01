GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)
FAQ_BIN = faq-$(GOOS)-$(GOARCH)

ifeq ($(GOOS), linux)
FAQ_LINK_STATIC=true
INSTALL=install
else
FAQ_LINK_STATIC=false
INSTALL=ginstall
endif

GO_EXT_LD_FLAGS=-v
ifeq ($(FAQ_LINK_STATIC), true)
GO_EXT_LD_FLAGS+= -static
endif

FAQ_VERSION=$(shell git describe --always --abbrev=40 --dirty)
GO_LD_FLAGS=-s -w -X main.version=$(FAQ_VERSION) -linkmode external -extldflags "$(GO_EXT_LD_FLAGS)"

GO=go
GO_BUILD_ARGS=-v -ldflags '$(GO_LD_FLAGS)'
GO_FILES:=$(shell find . -name '*.go' -type f)

IMAGE_TAG = latest
IMAGE_REPO = quay.io/jzelinskie/faq

DEFAULT_TARGETS=test validate $(FAQ_BIN)

prefix = /usr/local
exec_prefix = $(prefix)
bindir = $(exec_prefix)/bin

all:
	$(MAKE) test
ifneq ($(SKIP_VALIDATE),true)
	$(MAKE) validate
endif
	$(MAKE) build

install:
	mkdir -p $(DESTDIR)/usr/bin
	$(INSTALL) -m 0755 $(FAQ_BIN) $(DESTDIR)$(bindir)/faq

$(FAQ_BIN): $(GO_FILES)
	CGO_ENABLED=1 GOOS=$(GOOS) GOARCH=$(GOARCH) $(GO) build -o $(FAQ_BIN) $(GO_BUILD_ARGS) github.com/jzelinskie/faq

PHONY: build
build: $(FAQ_BIN)

PHONY: docker-build
docker-build:
	docker build -t $(IMAGE_REPO):$(IMAGE_TAG) .

PHONY: test
test:
	scripts/test.sh

PHONY: validate
validate:
	scripts/validate.sh

clean:
	rm $(FAQ_BIN)
