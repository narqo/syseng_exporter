GO ?= go
GOFLAGS ?=

DOCKER ?= docker

PKG := ./
IMAGE_NAME := syseng-exporter

.PHONY: all
all: build test

.PHONY: build
build:
	$(GO) build $(GOFLAGS) -o syseng_exporter $(PKG)

.PHONY: build-docker
build-docker:
	$(DOCKER) build -t $(IMAGE_NAME) $(PKG)

.PHONY: test
test:
	$(GO) test $(GOFLAGS) -v $(PKG)

.PHONY: clean
clean:
	$(GO) clean $(GOFLAGS) -i $(PKG)
	$(RM) syseng_exporter
