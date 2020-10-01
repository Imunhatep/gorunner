GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)
GOFLAGS = -mod=vendor
export GOFLAGS

APP_VERSION ?= devel
IMAGE_PREFIX ?= imunhatep/
IMAGE_TAG ?= latest

.PHONY: build
build: build/gorunner

.PHONY: build/gorunner
build/gorunner:
	mkdir -p build
	CGO_ENABLED=0 go build -o $@ -ldflags "-X github.com/imunhatep/gorunner.Version=$(APP_VERSION)" ./

.PHONY: image
image:
	docker build -t $(IMAGE_PREFIX)gorunner:$(IMAGE_TAG) .
