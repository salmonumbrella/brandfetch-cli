SHELL := /bin/bash

.PHONY: fmt lint test build clean tools setup

setup:
	@command -v lefthook >/dev/null || (echo "Install lefthook: brew install lefthook" && exit 1)
	lefthook install

TOOLS_DIR := $(CURDIR)/.tools
GOFUMPT := $(TOOLS_DIR)/gofumpt
GOIMPORTS := $(TOOLS_DIR)/goimports
GOLANGCI_LINT := $(TOOLS_DIR)/golangci-lint

tools:
	@mkdir -p $(TOOLS_DIR)
	@GOBIN=$(TOOLS_DIR) go install mvdan.cc/gofumpt@latest
	@GOBIN=$(TOOLS_DIR) go install golang.org/x/tools/cmd/goimports@latest
	@GOBIN=$(TOOLS_DIR) go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

fmt: tools
	@$(GOIMPORTS) -w .
	@$(GOFUMPT) -w .

fmt-check: tools
	@$(GOIMPORTS) -w .
	@$(GOFUMPT) -w .
	@git diff --exit-code -- '*.go' go.mod go.sum

lint: tools
	@$(GOLANGCI_LINT) run

test:
	@go test -v ./...

build:
	@go build -o ./bin/brandfetch ./cmd/brandfetch

clean:
	@rm -rf ./bin ./.tools

ci: fmt-check lint test
