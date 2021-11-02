BUILD_IMAGE_SRC_DIR := "/code"

LOCAL_GOCACHE_DIR := $(shell go env GOCACHE)
BUILD_IMAGE_GOCACHE_DIR := "/.cache/go-build"

LOCAL_GOMODCACHE_DIR := $(shell go env GOMODCACHE)
BUILD_IMAGE_GOMODCACHE_DIR := "/go/pkg/mod"

UID := $(shell id -u)
GID := $(shell id -g)

# Default target executed when no arguments are given to make.
default_target: build_local

.PHONY: default_target

all: build_local build_docker

build_local: build/Makefile
	make -f build/Makefile build

install_local: build/Makefile
	make -f build/Makefile install

build_docker: prep_docker_builder
	docker run --rm -it \
 		--name cosmzone-builder \
		-u ${UID}:${GID} \
		-v ${PWD}:${BUILD_IMAGE_SRC_DIR} \
		-v ${LOCAL_GOCACHE_DIR}:${BUILD_IMAGE_GOCACHE_DIR} \
		-v ${LOCAL_GOMODCACHE_DIR}:${BUILD_IMAGE_GOMODCACHE_DIR} \
		nomo/builder

prep_docker_builder: build/builder_spec
	docker build -t nomo/builder -f build/builder_spec \
		--build-arg CODE_DIR=${BUILD_IMAGE_SRC_DIR} \
		--build-arg GOCACHE_DIR=${BUILD_IMAGE_GOCACHE_DIR} \
		--build-arg GOMODCACHE_DIR=${BUILD_IMAGE_GOMODCACHE_DIR} .
