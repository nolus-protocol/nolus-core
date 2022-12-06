BUILDDIR ?= $(CURDIR)/target/release
PACKAGES=$(shell go list ./... | grep -v '/simulation')
TMVERSION := $(shell go list -m github.com/tendermint/tendermint | sed 's:.* ::')
LEDGER_ENABLED ?= true
VERSION := $(shell echo $(shell git describe --tags) | sed 's/^v//')
COMMIT := $(shell git log -1 --format='%H')
NOLUS_BINARY=nolusd
FUZZ_NUM_SEEDS ?= 2
FUZZ_NUM_RUNS_PER_SEED ?= 3
FUZZ_NUM_BLOCKS ?= 100
FUZZ_BLOCK_SIZE ?= 200
export GO111MODULE = on

# Default target executed when no arguments are given to make.
default_target: all

.PHONY: default_target

# process build tags

build_tags = netgo
ifeq ($(LEDGER_ENABLED),true)
	ifeq ($(OS),Windows_NT)
		GCCEXE = $(shell where gcc.exe 2> NUL)
		ifeq ($(GCCEXE),)
			$(error gcc.exe not installed for ledger support, please install or set LEDGER_ENABLED=false)
		else
			build_tags += ledger
		endif
	else
		UNAME_S = $(shell uname -s)
		ifeq ($(UNAME_S),OpenBSD)
			$(warning OpenBSD detected, disabling ledger support (https://github.com/cosmos/cosmos-sdk/issues/1988))
		else
			GCC = $(shell command -v gcc 2> /dev/null)
			ifeq ($(GCC),)
				$(error gcc not installed for ledger support, please install or set LEDGER_ENABLED=false)
			else
				build_tags += ledger
			endif
		endif
	endif
endif

ifeq (cleveldb,$(findstring cleveldb,$(NOLUS_BUILD_OPTIONS)))
	build_tags += gcc
endif
build_tags += $(BUILD_TAGS)
build_tags := $(strip $(build_tags))

whitespace :=
whitespace += $(whitespace)
comma := ,
build_tags_comma_sep := $(subst $(whitespace),$(comma),$(build_tags))

ldflags = -X github.com/cosmos/cosmos-sdk/version.Name=nolus \
			-X github.com/cosmos/cosmos-sdk/version.AppName=${NOLUS_BINARY} \
			-X github.com/cosmos/cosmos-sdk/version.Version=$(VERSION) \
			-X github.com/cosmos/cosmos-sdk/version.Commit=$(COMMIT) \
			-X "github.com/cosmos/cosmos-sdk/version.BuildTags=$(build_tags_comma_sep)" \
			-X github.com/tendermint/tendermint/version.TMCoreSemVer=$(TMVERSION)

# DB backend selection
ifeq (cleveldb,$(findstring cleveldb,$(COSMOS_BUILD_OPTIONS)))
	ldflags += -X github.com/cosmos/cosmos-sdk/types.DBBackend=cleveldb
endif
ifeq (badgerdb,$(findstring badgerdb,$(COSMOS_BUILD_OPTIONS)))
	ldflags += -X github.com/cosmos/cosmos-sdk/types.DBBackend=badgerdb
endif
# handle rocksdb
ifeq (rocksdb,$(findstring rocksdb,$(COSMOS_BUILD_OPTIONS)))
	CGO_ENABLED=1
	BUILD_TAGS += rocksdb
	ldflags += -X github.com/cosmos/cosmos-sdk/types.DBBackend=rocksdb
endif
# handle boltdb
ifeq (boltdb,$(findstring boltdb,$(COSMOS_BUILD_OPTIONS)))
	BUILD_TAGS += boltdb
	ldflags += -X github.com/cosmos/cosmos-sdk/types.DBBackend=boltdb
endif

ifeq (,$(findstring nostrip,$(COSMOS_BUILD_OPTIONS)))
	ldflags += -w -s
endif
ifeq ($(LINK_STATICALLY),true)
	ldflags += -linkmode=external -extldflags "-Wl,-z,muldefs -static"
endif
ldflags += $(LDFLAGS) 
ldflags := $(strip $(ldflags))

BUILD_FLAGS := -tags "$(build_tags)" -ldflags '$(ldflags)'
# check for nostrip option
ifeq (,$(findstring nostrip,$(COSMOS_BUILD_OPTIONS)))
	BUILD_FLAGS += -trimpath
endif

ifneq (${GOCACHE_DIR},)
	export GOCACHE=${GOCACHE_DIR}
endif

ifneq (${GOMODCACHE_DIR},)
	export GOMODCACHE=${GOMODCACHE_DIR}
endif

ifneq (${WASMVM_DIR},)
	export CGO_LDFLAGS=-L${WASMVM_DIR}
endif

ifeq (${GOROOT},)
	export GOROOT=$(shell go env GOROOT)
endif

#$(info $$BUILD_FLAGS is [$(BUILD_FLAGS)])

###############################################################################
###                              Documentation                              ###
###############################################################################
.PHONY: all build install go.sum fuzz

all: build install fuzz test-unit-cosmos

BUILD_TARGETS := build install

build: BUILD_ARGS=-o $(BUILDDIR)/

lint:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	golangci-lint run --verbose

test-fuzz:
	go test ./app $(BUILD_FLAGS) -mod=readonly -run TestAppStateDeterminism -Enabled=true \
		-NumBlocks=$(FUZZ_NUM_BLOCKS) -BlockSize=$(FUZZ_BLOCK_SIZE) -Commit=true -Period=0 -v \
		-NumSeeds=$(FUZZ_NUM_SEEDS) -NumTimesToRunPerSeed=$(FUZZ_NUM_RUNS_PER_SEED) -timeout 24h

test-unit-cosmos:
	./scripts/test/run-test-unit-cosmos.sh >&2

test-unit:
	go install gotest.tools/gotestsum@latest
	gotestsum --junitfile testreport.xml --format testname -- $(BUILD_FLAGS) -mod=readonly -coverprofile=cover.out -covermode=atomic ./...

test-unit-coverage: ## Generate global code coverage report
	go install github.com/boumenot/gocover-cobertura@latest
	gocover-cobertura < cover.out > coverage.xml

test-unit-coverage-report: ## Generate global code coverage report in HTML
	sh  ./scripts/test/coverage.sh html;

$(BUILD_TARGETS): go.sum $(BUILDDIR)/
	go $@ -mod=readonly $(BUILD_FLAGS) $(BUILD_ARGS) ./...

$(BUILDDIR)/:
	mkdir -p $(BUILDDIR)/

go.sum: go.mod
	@echo "--> Ensure dependencies have not been modified"
	@go mod verify

.PHONY: all build install go.sum fuzz test-unit-cosmos check-format

###############################################################################
###                                  Proto                                  ###
###############################################################################
protoVer=v0.7
protoImageName=tendermintdev/sdk-proto-gen:$(protoVer)
containerProtoGen=nolus-proto-gen-$(protoVer)

proto-gen:
	@echo "Generating Protobuf files"
	@if docker ps -a --format '{{.Names}}' | grep -Eq "^${containerProtoGen}$$"; then docker start -a $(containerProtoGen); else docker run --name $(containerProtoGen) -v $(CURDIR):/workspace --workdir /workspace $(protoImageName) \
		sh ./scripts/protocgen.sh; fi
