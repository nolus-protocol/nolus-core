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

	# failed to initialize database: open /tmp/Test_runMigrateCmd3518174278/001/keys/keys.db/LOCK: permission denied
	if [ "$PACKAGE" == "github.com/cosmos/cosmos-sdk/client/keys" ]; then
		skip_test "Test_runMigrateCmd" $PACKAGE
	fi

	# failed to initialize database: open /tmp/TestLegacyKeybase2255353028/001/keys/keys.db/LOCK: permission denied
	if [ "$PACKAGE" == "github.com/cosmos/cosmos-sdk/crypto/keyring" ]; then
		skip_test "TestLegacyKeybase" $PACKAGE
	fi

	# Expected nil, but got: &fs.PathError{Op:"open", Path:".touch", Err:0xd}
	if [ "$PACKAGE" == "github.com/cosmos/cosmos-sdk/store/streaming" ]; then
		skip_test "TestStreamingServiceConstructor" $PACKAGE
	fi

	if [ "$TEST_LIST" == "skip all" ]; then
		echo "No tests in package: $PACKAGE"
	else
		go test -mod=readonly -tags='ledger test_ledger_mock' -race -timeout 30m -run=$TEST_LIST $PACKAGE
	fi
done
