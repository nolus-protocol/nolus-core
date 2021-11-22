BUILDDIR ?= $(CURDIR)/target/release
PACKAGES=$(shell go list ./... | grep -v '/simulation')
TMVERSION := $(shell go list -m github.com/tendermint/tendermint | sed 's:.* ::')
LEDGER_ENABLED ?= true
VERSION := $(shell echo $(shell git describe --tags) | sed 's/^v//')
COMMIT := $(shell git log -1 --format='%H')
NOMO_BINARY=cosmozoned
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

ifeq (cleveldb,$(findstring cleveldb,$(NOMO_BUILD_OPTIONS)))
  build_tags += gcc
endif
build_tags += $(BUILD_TAGS)
build_tags := $(strip $(build_tags))

whitespace :=
whitespace += $(whitespace)
comma := ,
build_tags_comma_sep := $(subst $(whitespace),$(comma),$(build_tags))

ldflags = -X github.com/cosmos/cosmos-sdk/version.Name=nomo \
		  -X github.com/cosmos/cosmos-sdk/version.AppName=${NOMO_BINARY} \
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

#$(info $$BUILD_FLAGS is [$(BUILD_FLAGS)])

###############################################################################
###                              Documentation                              ###
###############################################################################

all: build install fuzz test-unit-cosmos

BUILD_TARGETS := build install

build: BUILD_ARGS=-o $(BUILDDIR)/

fuzz:
	go test $(BUILD_FLAGS) -mod=readonly ./app -run TestAppStateDeterminism -Enabled=true -NumBlocks=$(FUZZ_NUM_BLOCKS) -BlockSize=$(FUZZ_BLOCK_SIZE) -Commit=true -Period=0 -v -timeout 24h -NumSeeds=$(FUZZ_NUM_SEEDS) -NumTimesToRunPerSeed=$(FUZZ_NUM_RUNS_PER_SEED)

test-unit-cosmos:
  $(shell ls -al ./scripts/test && ./scripts/test/run-test-unit-cosmos.sh >&2)

$(BUILD_TARGETS): go.sum $(BUILDDIR)/
	go $@ -mod=readonly $(BUILD_FLAGS) $(BUILD_ARGS) ./...

$(BUILDDIR)/:
	mkdir -p $(BUILDDIR)/

go.sum: go.mod
	@echo "--> Ensure dependencies have not been modified"
	@go mod verify

.PHONY: all build install go.sum fuzz