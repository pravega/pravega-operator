# Copyright (c) 2018 Dell Inc., or its subsidiaries. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0

SHELL=/bin/bash -o pipefail

REPO?=pravega/pravega-operator
# TAG?=$(shell git describe --always --tags --dirty)
VERSION?=$(shell cat VERSION)
GOOS:=linux
GOARCH:=amd64

.PHONY: all build check clean test

all: check build

build: test build-go build-image

build-go:
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build \
	-ldflags "-X github.com/pravega/pravega-operator/pkg/version.Version=$(VERSION)" \
	-o bin/pravega-operator cmd/pravega-operator/main.go

build-image: build-go
	docker build -t $(REPO):$(VERSION) .
	docker build -t $(REPO):latest .

test:
	go test $$(go list ./... | grep -v /vendor/)

login:
	@docker login -u "$(DOCKER_USER)" -p "$(DOCKER_PASS)"

push: build-image login
	docker push $(REPO):latest
	docker push $(REPO):$(VERSION)

clean:
	rm -f bin/pravega-operator

check: check-format check-license

check-format:
	./scripts/check_format.sh

check-license:
	./scripts/check_license.sh
