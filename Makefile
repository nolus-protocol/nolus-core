BUILD_IMAGE_SRC_DIR := "/code"

LOCAL_GOCACHE_DIR := $(shell go env GOCACHE)
BUILD_IMAGE_GOCACHE_DIR := "/.cache/go-build"

LOCAL_GOMODCACHE_DIR := $(shell go env GOMODCACHE)
BUILD_IMAGE_GOMODCACHE_DIR := "/go/pkg/mod"

BUILDDIR ?= $(CURDIR)/target/release

WASMVM_DOWNLOAD_URL := https://github.com/CosmWasm/wasmvm/releases/download/v0.16.0
WASMVM_LIB := libwasmvm_muslc.a
WASMVM_CHECKSUMS := checksums.txt
WASMVM_LIB_CHECKSUM := $(shell curl -L ${WASMVM_DOWNLOAD_URL}/${WASMVM_CHECKSUMS} | grep ${WASMVM_LIB} | awk '{print $$1;}')
WASMVM_LIB_URL := ${WASMVM_DOWNLOAD_URL}/${WASMVM_LIB}
WASMVM_MOD := github.com/\!cosm\!wasm

WASMVM_LOCAL_LIB_PATH := ${LOCAL_GOMODCACHE_DIR}/${WASMVM_MOD}/${WASMVM_LIB}
WASMVM_BUILD_IMAGE_LIB_DIR := ${BUILD_IMAGE_GOMODCACHE_DIR}/${WASMVM_MOD}

UID := $(shell id -u)
GID := $(shell id -g)

# Default target executed when no arguments are given to make.
default_target: build_local

.PHONY: default_target build_local install_local build_docker build_node prep_docker_builder

all: build_local build_docker

build_local: build/Makefile
	make -f build/Makefile BUILDDIR=${BUILDDIR} build

install_local: build/Makefile
	make -f build/Makefile BUILDDIR=${BUILDDIR} install

build_docker: prep_docker_builder ${WASMVM_LOCAL_LIB_PATH}
	docker run --rm -it \
		--name cosmzone-builder \
		-u ${UID}:${GID} \
		-v ${PWD}:${BUILD_IMAGE_SRC_DIR} \
		-v ${LOCAL_GOCACHE_DIR}:${BUILD_IMAGE_GOCACHE_DIR} \
		-v ${LOCAL_GOMODCACHE_DIR}:${BUILD_IMAGE_GOMODCACHE_DIR} \
		nomo/builder build_local

build_node: build/node_spec build_docker
	docker build -t nomo/node -f build/node_spec .

# implementation targets follow
prep_docker_builder: build/builder_spec
	docker build -t nomo/builder -f build/builder_spec \
		--build-arg CODE_DIR=${BUILD_IMAGE_SRC_DIR} \
		--build-arg GOCACHE_DIR=${BUILD_IMAGE_GOCACHE_DIR} \
		--build-arg GOMODCACHE_DIR=${BUILD_IMAGE_GOMODCACHE_DIR} \
		--build-arg CGO_LDFLAGS=-L${WASMVM_BUILD_IMAGE_LIB_DIR} .

${WASMVM_LOCAL_LIB_PATH}:
	@curl -L ${WASMVM_LIB_URL} -o ${WASMVM_LOCAL_LIB_PATH}
	@sha256sum ${WASMVM_LOCAL_LIB_PATH} | grep ${WASMVM_LIB_CHECKSUM} > /dev/null
