GO ?= go
DOCKER ?= docker

GOFLAGS :=
PKG := ./
IMAGE_NAME := syseng-exporter

GIT_REV := $(shell git describe --always --tags --dirty=-dev)

.PHONY: all
all: build test

.PHONY: build
build:
	$(GO) build $(GOFLAGS) -ldflags "-X main.revision=$(GIT_REV)" -o syseng_exporter $(PKG)

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
