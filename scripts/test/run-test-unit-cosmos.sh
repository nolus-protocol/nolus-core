#!/bin/bash
set -euo pipefail

COSMOSSDK_DIR=$(go list -m -f '{{.Dir}}' github.com/cosmos/cosmos-sdk)
echo "Cosmos SDK Path: $COSMOSSDK_DIR"

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
}

if [ -z "$COSMOSSDK_DIR" ]; then
	echo "There is no Cosmos SDK"
else
	COSMOSSDK_PACKAGES=$(go list $COSMOSSDK_DIR/... | uniq)
	# COSMOSSDK_PACKAGES="github.com/cosmos/cosmos-sdk/server github.com/cosmos/cosmos-sdk/x/upgrade/types"
	for PACKAGE in $COSMOSSDK_PACKAGES; do
		TEST_LIST=""

		# Skip the TestInterceptConfigsWithBadPermissions test, as it fails when running the test as the root user
		if [ "$PACKAGE" == "github.com/cosmos/cosmos-sdk/server" ]; then
			skip_test "TestInterceptConfigsWithBadPermissions" $PACKAGE
		fi
		
		# Skip because TestIntegrationTestSuite/TestBroadcastTx_GRPCGateway/valid_request fails
		if [ "$PACKAGE" == "github.com/cosmos/cosmos-sdk/x/auth/tx" ]; then
			skip_test "TestIntegrationTestSuite" $PACKAGE
		fi

		# skip entire package due to build error
		if [ "$PACKAGE" == "github.com/cosmos/cosmos-sdk/server/grpc" ]; then
			continue
		fi

		echo "go test -mod=readonly -tags='cgo ledger test_ledger_mock norace' -run=$TEST_LIST $PACKAGE"
		go test -mod=readonly -tags='cgo ledger test_ledger_mock norace' -run=$TEST_LIST $PACKAGE
	done
fi
