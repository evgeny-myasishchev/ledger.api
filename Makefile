SERVICE_NAME = ledger-api

VERSION = 0.0.4-a

REGISTRY ?= evgenymyasishchev

SRC_DIRS := cmd pkg # directories which hold app source (not vendored)

OUTBIN := $(GOPATH)/bin/$(BIN)

GIT_HASH = $(shell git rev-parse HEAD)
GIT_REF = $(shell git branch | grep \\* | cut -d ' ' -f2)
GIT_URL = $(shell git config --get remote.origin.url)

INSTALL_ENV = CGO_ENABLED=0 GO111MODULE=on
INSTALL_FLAGS = -installsuffix "static" -ldflags "\
	-X $(shell go list -m)/pkg/version.AppName=${SERVICE_NAME}\
	-X $(shell go list -m)/pkg/version.Version=${VERSION}\
	-X $(shell go list -m)/pkg/version.GitHash=${GIT_HASH}\
	-X $(shell go list -m)/pkg/version.GitRef=${GIT_REF}\
	-X $(shell go list -m)/pkg/version.GitURL=${GIT_URL}\
	"

DEV_IMAGE = dev/${SERVICE_NAME}:latest
PROD_IMAGE = ${REGISTRY}/${SERVICE_NAME}


all: build

build:
	$(INSTALL_ENV) go install $(INSTALL_FLAGS) ./...

test: 
	@go test ./...

docker_build:
	@echo Building dev image
	docker build -f docker/Dockerfile.dev --build-arg SERVICE_NAME=$(SERVICE_NAME) . -t ${DEV_IMAGE}

docker_build_release: docker_build
	@echo Building release image
	docker build \
		--build-arg DEV_IMAGE=$(DEV_IMAGE) \
		--build-arg SERVICE_NAME=$(SERVICE_NAME) \
		-f docker/Dockerfile.prod . \
		-t ${PROD_IMAGE}:latest -t ${PROD_IMAGE}:${VERSION}

docker_push_release:
	@echo Pushing release image
	docker push ${PROD_IMAGE}:latest ${PROD_IMAGE}:${VERSION}
