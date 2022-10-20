#!/bin/bash
set -euxo pipefail

COSMOSSDK_DIR=$(go list -m -f '{{.Dir}}' github.com/cosmos/cosmos-sdk)
echo "Cosmos SDK Path: $COSMOSSDK_DIR"

if [ -z "$COSMOSSDK_DIR" ]
then
	echo "There is no Cosmos SDK"
else
	COSMOSSDK_PACKAGES=$(go list $COSMOSSDK_DIR/... | uniq)
	echo "Cosmos SDK packages:"
	echo "$COSMOSSDK_PACKAGES"

	for PACKAGE in $COSMOSSDK_PACKAGES
	do
		echo "Running unit tests for $PACKAGE"

		# Skip the TestInterceptConfigsWithBadPermissions test, as it fails when running
		# the test as the root user.
		# To skip it, we need to:
		# 1. Retrieve all the tests in the same package
		# 2. Remove the test we want to skip
		# 3. Format the list of remaining tests as a regex in the form TestName|TestName|...
		# 4. Pass the list of tests to the go test command with the -run flag
		# There is no easier way to skip specific tests with go test currently.
		# See: https://github.com/golang/go/issues/41583
		if [ "$PACKAGE" == "github.com/cosmos/cosmos-sdk/server" ]
		then
			ARGS="-run $(go test -list '.*' github.com/cosmos/cosmos-sdk/server | \
				grep -v TestInterceptConfigsWithBadPermissions | \
				sed '$d' | \
				sed ':a; /$/N; s/\n/|/; ta')"
		else
			ARGS=
		fi

		go test -mod=readonly -tags='cgo ledger test_ledger_mock norace' $ARGS "$PACKAGE"
	done
fi
