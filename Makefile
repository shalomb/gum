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

.PHONY: test test-unit test-integration test-coverage test-clean

build: build-env
	go mod tidy
	go build -o gum

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
