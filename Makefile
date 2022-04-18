# Copyright (c) 2018 Dell Inc., or its subsidiaries. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0

SHELL=/bin/bash -o pipefail
# This flag works only with controller-gen 0.6.2
CRD_OPTIONS ?= "crd:trivialVersions=true"
PROJECT_NAME=pravega-operator
REPO=testpravegaop/$(PROJECT_NAME)
BASE_VERSION=0.5.7
ID=$(shell git rev-list HEAD --count)
GIT_SHA=$(shell git rev-parse --short HEAD)
VERSION=$(BASE_VERSION)-$(ID)-$(GIT_SHA)
TEST_IMAGE=$(REPO)-testimages:$(VERSION)
GOOS=linux
GOARCH=amd64
DOCKER_TEST_PASS=testpravegaop
DOCKER_TEST_USER=testpravegaop

.PHONY: all  build check clean test
# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

all: check build test

build: build-go build-image

# Generate manifests e.g. CRD, RBAC etc.
manifests: controller-gen
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

deploy: manifests kustomize
	cd config/manager && $(KUSTOMIZE) edit set image pravega/pravega-operator=$(TEST_IMAGE)
	$(KUSTOMIZE) build config/default | kubectl apply -f -

build-go:
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build \
	-ldflags "-X github.com/$(REPO)/pkg/version.Version=$(VERSION) -X github.com/$(REPO)/pkg/version.GitSHA=$(GIT_SHA)" \
	-o bin/$(PROJECT_NAME) main.go

# find or download controller-gen
# download controller-gen if necessary
controller-gen:
ifeq (, $(shell which controller-gen))
	@{ \
	set -e ;\
	CONTROLLER_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$CONTROLLER_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.6.2 ;\
	rm -rf $$CONTROLLER_GEN_TMP_DIR ;\
	}
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif

kustomize:
ifeq (, $(shell which kustomize))
	@{ \
	set -e ;\
	KUSTOMIZE_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$KUSTOMIZE_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go install sigs.k8s.io/kustomize/kustomize/v4@latest ;\
	rm -rf $$KUSTOMIZE_GEN_TMP_DIR ;\
	}
KUSTOMIZE=$(GOBIN)/kustomize
else
KUSTOMIZE=$(shell which kustomize)
endif
# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests kustomize
	cd config/manager && $(KUSTOMIZE) edit set image pravega/bookkeeper-operator=$(TEST_IMAGE)
	$(KUSTOMIZE) build config/default | kubectl apply -f -


# Undeploy controller in the configured Kubernetes cluster in ~/.kube/config
undeploy:
	$(KUSTOMIZE) build config/default | kubectl delete -f -

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy-test: manifests kustomize
	cd config/test
	$(KUSTOMIZE) build config/test | kubectl apply -f -

# Undeploy controller in the configured Kubernetes cluster in ~/.kube/config
undeploy-test: manifests kustomize
	cd config/test
	$(KUSTOMIZE) build config/test | kubectl apply -f -

build-image:
	docker build --build-arg DOCKER_REGISTRY=$(DOCKER_REGISTRY) --build-arg VERSION=$(VERSION) --build-arg GIT_SHA=$(GIT_SHA) -t $(REPO):$(VERSION) .
	docker tag $(REPO):$(VERSION) $(REPO):latest

test: test-unit test-e2e

test-unit:
	 go test $$(go list ./... | grep -v /vendor/ | grep -v /test/e2e ) -race -coverprofile=coverage.txt -covermode=atomic

test-e2e: test-e2e-remote

test-e2e-remote:
	 make login
	 docker build . -t $(TEST_IMAGE)
	 docker push $(TEST_IMAGE)
	 make deploy
	 RUN_LOCAL=false go test -v -timeout 2h ./test/e2e...
	 make undeploy

test-e2e-local:
	make deploy-test
	RUN_LOCAL=true go test -v -timeout 2h ./test/e2e... -args -ginkgo.v
	make undeploy-test

run-local:
	go run ./main.go

login:
	echo "$(DOCKER_TEST_PASS)" | docker login -u "$(DOCKER_TEST_USER)" --password-stdin

push: build login
	docker push $(REPO):$(VERSION)
	if [[ ${TRAVIS_TAG} =~ ^([0-9]+\.[0-9]+\.[0-9]+)$$ ]]; then docker push $(REPO):latest; fi;

clean:
	rm -f bin/$(PROJECT_NAME)

check: check-format check-license

check-format:
	./scripts/check_format.sh

check-license:
	./scripts/check_license.sh
