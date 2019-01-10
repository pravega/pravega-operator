# Copyright (c) 2018 Dell Inc., or its subsidiaries. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0

SHELL=/bin/bash -o pipefail

PROJECT_NAME=pravega-operator
REPO=pravega/$(PROJECT_NAME)
VERSION=$(shell git describe --always --tags --dirty | sed "s/\(.*\)-g`git rev-parse --short HEAD`/\1/")
GIT_SHA=$(shell git rev-parse --short HEAD)
GOOS=linux
GOARCH=amd64
DEPLOY_IMAGE=pravega/pravega-operator:latest
TEST_IMAGE=tristan1900/test:v0.0.1

.PHONY: all dep build check clean test

all: check test build

dep:
	dep ensure -v

build: build-go build-image

build-go:
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build \
	-ldflags "-X github.com/$(REPO)/pkg/version.Version=$(VERSION) -X github.com/$(REPO)/pkg/version.GitSHA=$(GIT_SHA)" \
	-o bin/$(PROJECT_NAME) cmd/manager/main.go

build-image:
	docker build --build-arg VERSION=$(VERSION) --build-arg GIT_SHA=$(GIT_SHA) -t $(REPO):$(VERSION) .
	docker tag $(REPO):$(VERSION) $(REPO):latest

test: test-unit test-e2e

test-unit:
	go test $$(go list ./... | grep -v /vendor/ | grep -v /test/e2e )

test-e2e:
	sed "s@DEPLOY_IMAGE@$(TEST_IMAGE)@g" -i deploy/operator.yaml
	operator-sdk build $(TEST_IMAGE) --enable-tests
	docker login -u "$(DOCKER_TEST_USER)" -p "$(DOCKER_TEST_PASS)"
	docker push $(TEST_IMAGE)
	kubectl create -f deploy/crd.yaml
	kubectl create -f deploy/service_account.yaml
	kubectl create -f deploy/role.yaml
	kubectl create -f deploy/role_binding.yaml
	operator-sdk test cluster $(TEST_IMAGE) --namespace default --service-account pravega-operator

login:
	@docker login -u "$(DOCKER_USER)" -p "$(DOCKER_PASS)"

push: build-image login
	docker push $(REPO):$(VERSION)
	docker push $(REPO):latest

clean:
	rm -f bin/$(PROJECT_NAME)

check: check-format check-license

check-format:
	./scripts/check_format.sh

check-license:
	./scripts/check_license.sh
