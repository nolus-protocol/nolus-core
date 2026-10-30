#!/bin/bash
set -euo pipefail

# Runs the unit tests of the cosmos-sdk version this module depends on.
#
# Skipped tests, accepted by Lockwarr on 2026-10-30:
# - TestCoinTestSuite/TestQuoIntCoins: compares big.Int internals; Go 1.27
#   represents zero with an empty slice instead of nil, the value is equal.
# - TestCLI/--output=json_flag: version.Info marshals build_deps as strings but
#   cannot unmarshal them; the list is only non-empty outside the SDK module.
# - TestParseChainIDFromGenesis/invalid_type: asserts the exact encoding/json
#   error text, which changed in Go 1.27; the input is still rejected.
# - TestGenesisStateFromGenFile: needs a fixture that only exists in the
#   cosmos-sdk repository.
SKIP='^TestCoinTestSuite$/^TestQuoIntCoins$'
SKIP+='|^TestCLI$/^--output=json_flag$'
SKIP+='|^TestParseChainIDFromGenesis$/^invalid_type$'
SKIP+='|^TestGenesisStateFromGenFile$'

go test -mod=readonly -tags='ledger test_ledger_mock' -race -timeout 30m \
	-skip "$SKIP" github.com/cosmos/cosmos-sdk/...
