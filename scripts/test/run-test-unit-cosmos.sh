#!/bin/sh

COSMOSSDK_DIR=$(go list -m -f '{{.Dir}}' github.com/cosmos/cosmos-sdk)
echo "Cosmos SDK Path: $COSMOSSDK_DIR"

if [ -z "$COSMOSSDK_DIR" ]
then
	echo "There is no Cosmos SDK"
else
	cd $COSMOSSDK_DIR
	COSMOSSDK_PACKAGES=$(go list ./... | uniq)
	echo "Cosmos SDK packages:"
	echo "$COSMOSSDK_PACKAGES"

	for PACKAGE in $COSMOSSDK_PACKAGES
	do
		echo "Running unit tests for $PACKAGE"
		go test $PACKAGE
	done
fi
