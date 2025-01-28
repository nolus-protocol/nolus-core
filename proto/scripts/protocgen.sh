#!/usr/bin/env bash
set -eo pipefail

go mod tidy

echo "Generating gogo proto code"
cd proto
buf dep update
buf generate
cd ..

# move proto files to the right places
cp -r ./github.com/Nolus-Protocol/nolus-core/x/* x/  
rm -rf ./github.com
