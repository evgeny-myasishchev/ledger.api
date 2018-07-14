# The binary to build (just the basename).
BIN := ledger-api

# This repo's root import path (under GOPATH).
PKG := ledger.api

# Where to push the docker image.
REGISTRY ?= evgenymyasishchev
#
# This version-strategy uses a manual value to set the version string
VERSION := 0.0.3-a

###
### These variables should not need tweaking.
###

SRC_DIRS := cmd pkg # directories which hold app source (not vendored)

IMAGE := $(REGISTRY)/$(BIN)

# If you want to build all binaries, see the 'all-build' rule.
# If you want to build all containers, see the 'all-container' rule.
# If you want to build AND push all containers, see the 'all-push' rule.
all: build

build: bin

bin:
	go install -ldflags "-X ${PKG}/pkg/version.VERSION=${VERSION}" ./...

docker-build: 
	@docker build . -t $(IMAGE):$(VERSION)

docker-shell:  docker-build
	@docker run --rm -it $(IMAGE):$(VERSION) sh

# docker-dev-shell:  docker-build
# 	@docker run \
# 		--rm -it \
# 		-u $$(id -u):$$(id -g) \
# 		-w /go/src/$(PKG) \
# 		$(IMAGE):$(VERSION) sh

docker-push: docker-build
	@docker push $(IMAGE):$(VERSION)
	@echo "pushed: $(IMAGE):$(VERSION)"

version:
	@echo $(VERSION)

docker-image:
	@echo ${IMAGE}:$(VERSION)

test: 
	go test ./...