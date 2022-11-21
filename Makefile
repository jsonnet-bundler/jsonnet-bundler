.PHONY: all check-license crossbuild build install test generate embedmd

SHELL=/bin/bash

GITHUB_URL=github.com/jsonnet-bundler/jsonnet-bundler
VERSION := $(shell git describe --tags --dirty --always)
OUT_DIR=_output
BIN?=jb
PKGS=$(shell go list ./... | grep -v /vendor/)

all: check-license build generate test

# Binaries
LDFLAGS := '-s -w -extldflags "-static" -X main.Version=${VERSION}'
cross: clean
	CGO_ENABLED=0 gox \
	  -output="$(OUT_DIR)/jb-{{.OS}}-{{.Arch}}" \
	  -ldflags=$(LDFLAGS) \
	  -arch="amd64 arm64 arm" -os="linux" \
	  -arch="amd64 arm64" -os="darwin" \
	  -osarch="windows/amd64" \
	  ./cmd/$(BIN)

static:
	CGO_ENABLED=0 go build -ldflags=${LDFLAGS} -o $(OUT_DIR)/$(BIN) ./cmd/$(BIN)

build:
	CGO_ENABLED=0 go build -o $(OUT_DIR)/$(BIN) ./cmd/$(BIN)

install: static
	@echo ">> copying $(BIN) into $(GOPATH)/bin/$(BIN)"
	cp $(OUT_DIR)/$(BIN) $(GOPATH)/bin/$(BIN)

# Tests
test:
	@echo ">> running all unit tests"
	go test -v $(PKGS)

test-integration:
	@echo ">> running all integration tests"
	go test -v -tags=integration $(PKGS)

# Documentation
generate: embedmd
	@echo ">> generating docs"
	@./scripts/generate-help-txt.sh
	$(GOPATH)/bin/embedmd -w `find ./ -path ./vendor -prune -o -name "*.md" -print`

check-license:
	@echo ">> checking license headers"
	@./scripts/check_license.sh

embedmd:
	pushd /tmp && go install github.com/campoy/embedmd@latest && popd

# Other
clean:
	rm -rf $(OUT_DIR) $(BIN)

drone:
	drone jsonnet --format
