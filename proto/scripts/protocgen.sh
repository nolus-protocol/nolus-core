#!/usr/bin/env bash

set -eo pipefail

echo "Generating gogo proto code"
cd proto
buf mod update
buf generate
cd ..

# move proto files to the right places
cp -r ./github.com/Nolus-Protocol/nolus-core/x/* x/  
rm -rf ./github.com


go mod tidy # -go=1.19
