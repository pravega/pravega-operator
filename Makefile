# Copyright (c) 2018 Dell Inc., or its subsidiaries. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0

SHELL=/bin/bash -o pipefail

REPO?=pravega/pravega-operator
TAG?=$(shell git rev-parse --short HEAD)
VERSION?=$(shell cat VERSION)
GOOS:=linux
GOARCH:=amd64
pkgs=$(shell go list ./... | grep -v /vendor/ | grep -v /test/)

.PHONY: all build image format clean test

all: format image

build: test
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build \
	-ldflags "-X github.com/pravega/pravega-operator/pkg/version.Version=$(VERSION)" \
	-o bin/pravega-operator cmd/pravega-operator/main.go

image: build
	docker build -t $(REPO):$(TAG) .
	docker build -t $(REPO):latest .

push: image
	docker push $(REPO):latest
	docker push $(REPO):$(TAG)

clean:
	rm -f bin/pravega-operator

format: go-fmt check-license

go-fmt:
	go fmt $(pkgs)

check-license:
	./scripts/check_license.sh

test:
	go test $$(go list ./... | grep -v /vendor/)
