GO := go
GOFLAGS :=

GIT := git
DOCKER := docker

GIT_REV := $(shell $(GIT) describe --always --tags --dirty=-dev)

OUTPUT_DIR := $(CURDIR)
IMAGE_NAME := syseng-exporter

BUILD.go = $(GO) build $(GOFLAGS)
TEST.go  = $(GO) test $(GOFLAGS)

go_packages := $(shell $(GO) list ./... | grep -v /vendor/)

.PHONY: all
all: build test

.PHONY: build
build:
	$(BUILD.go) -ldflags "-X main.revision=$(GIT_REV)" -o $(OUTPUT_DIR)/syseng_exporter .

.PHONY: build-docker
build-docker:
	$(DOCKER) build -t $(IMAGE_NAME) .

.PHONY: test
test:
	$(TEST.go) -v $(go_packages)

.PHONY: clean
clean:
	$(GO) clean $(GOFLAGS) -i .
	$(RM) $(OUTPUT_DIR)/syseng_exporter
