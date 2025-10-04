#!/usr/bin/make -f

.ONESHELL:
SHELL             := /bin/bash
SHELLFLAGS := -u nounset -ec

MAKE          := make
MAKEFILE      := $(realpath $(lastword $(MAKEFILE_LIST)))
MAKEFLAGS     += --no-print-directory
MAKEFLAGS     += --warn-undefined-variables
THIS_DIR      := $(shell dirname $(MAKEFILE))
THIS_PROJECT  := nq

# https://dustinrue.com/2021/08/parameters-in-a-makefile/
# setup arguments
RUN_ARGS          := $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))
# ...and turn them into do-nothing targets
$(eval $(RUN_ARGS):;@:)

.PHONY: build install install-user test test-unit test-integration test-coverage test-clean

build: build-env
	go mod tidy
	go build -ldflags "$(LDFLAGS)" -o gum

# Build flags for version information
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

LDFLAGS := -X 'github.com/shalomb/gum/cmd.Version=$(VERSION)' \
           -X 'github.com/shalomb/gum/cmd.GitCommit=$(GIT_COMMIT)' \
           -X 'github.com/shalomb/gum/cmd.BuildDate=$(BUILD_DATE)'

install: build
	@# Prefer user location first, fall back to system location
	@if [ -d "$$HOME/.local/bin" ]; then \
		echo "Installing to ~/.local/bin"; \
		cp gum $$HOME/.local/bin/; \
		chmod +x $$HOME/.local/bin/gum; \
		echo "gum installed to $$HOME/.local/bin/gum"; \
		echo "Make sure ~/.local/bin is in your PATH"; \
	elif [ -w "/usr/local/bin" ]; then \
		echo "Installing to /usr/local/bin"; \
		cp gum /usr/local/bin/; \
		chmod +x /usr/local/bin/gum; \
		echo "gum installed to /usr/local/bin/gum"; \
	else \
		echo "Installing to /usr/local/bin (requires sudo)"; \
		sudo cp gum /usr/local/bin/; \
		sudo chmod +x /usr/local/bin/gum; \
		echo "gum installed to /usr/local/bin/gum"; \
	fi

install-user: build
	@# Install to user location, creating directory if needed
	@mkdir -p $$HOME/.local/bin
	@echo "Installing to ~/.local/bin"
	@cp gum $$HOME/.local/bin/
	@chmod +x $$HOME/.local/bin/gum
	@echo "gum installed to $$HOME/.local/bin/gum"
	@echo "Make sure ~/.local/bin is in your PATH"
	@echo "Add this to your shell profile: export PATH=\"$$HOME/.local/bin:\$$PATH\""

init:
	go mod init github.com/shalomb/gum

build-env:
	go mod download

# Run all unit tests (excluding integration tests)
test-unit:
	go test ./internal/... ./cmd -v

# Run all tests including integration tests
test-integration:
	go test ./... -v

# Run tests with coverage reporting
test-coverage: test-unit
	go test ./internal/... ./cmd -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	go tool cover -func=coverage.out

# Clean test artifacts
test-clean:
	rm -f coverage.out coverage.html

# Run specific test package
test-pkg:
	go test ./$(PKG) -v

# Run tests with race detection
test-race:
	go test ./internal/... ./cmd -race -v

# Run benchmarks
test-bench:
	go test ./internal/... ./cmd -bench=. -benchmem

# Default test target (unit tests only)
test: test-unit
