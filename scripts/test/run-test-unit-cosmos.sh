#!/bin/bash
set -euo pipefail

# Usage: skip_test "$testFoo" "$packageBar"
# Returns TEST_LIST containing all the tests to be executed
# To skip it, we need to:
# 1. Retrieve all the tests in the same package
# 2. Remove the test we want to skip
# 3. Format the list of remaining tests as a regex in the form 'TestName|TestName|...'
# 4. Pass the list of tests to the go test command with the -run flag
#
# There is no easier way to skip specific tests with go test currently.
# See: https://github.com/golang/go/issues/41583
skip_test() {
	TEST_LIST="$(go test -list '.*' $2 |
		grep -v $1 |
		sed '$d' |
		sed ':a; /$/N; s/\n/|/; ta')"

	if [ -z "$TEST_LIST" ]; then
		TEST_LIST="skip all"
	fi
}

PKG_LIST=$(cat $1)

for PACKAGE in $PKG_LIST; do
	TEST_LIST=""

	# skip entire package due to build error. Also there are no tests in it
	if [ "$PACKAGE" == "github.com/cosmos/cosmos-sdk/server/grpc" ]; then
		continue
	fi

	# skip entire package due to build error - missing go.sum entry for module providing package github.com/cosmos/cosmos-sdk/db
	if [ "$PACKAGE" == "github.com/cosmos/cosmos-sdk/testutil/mock/db" ]; then
		continue
	fi

	# skip entire package due to build error - undefined: rapid.Run
	if [ "$PACKAGE" == "github.com/cosmos/cosmos-sdk/x/group/internal/orm" ]; then
		continue
	fi

	# failed to catch permissions error, got: [*errors.errorString] Cancelled in prerun
	if [ "$PACKAGE" == "github.com/cosmos/cosmos-sdk/server" ]; then
		skip_test "TestInterceptConfigsWithBadPermissions" $PACKAGE
	fi

	# # # ../../../tests/fixtures/adr-024-coin-metadata_genesis.json does not exist, run `init` first
	if [ "$PACKAGE" == "github.com/cosmos/cosmos-sdk/x/genutil/types" ]; then
		skip_test "TestGenesisStateFromGenFile" $PACKAGE
	fi

	if [ "$TEST_LIST" == "skip all" ]; then
		echo "No tests in package: $PACKAGE"
	else
		go test -mod=readonly -tags='ledger test_ledger_mock' -race -timeout 30m -run=$TEST_LIST $PACKAGE
	fi
done
