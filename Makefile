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

.PHONY: test

build: build-env
	go mod tidy
	go build

init:
	go mod init github.com/shalomb/gum

build-env:
	go mod download

test: build
	./gum "$(RUN_ARGS)"
