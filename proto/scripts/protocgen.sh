#!/usr/bin/env bash
set -eo pipefail

# move generated files to the right places
cp -r ./nolus ./api/  
rm -rf ./nolus

go mod tidy

echo "Generating gogo proto code"
cd proto
buf dep update
buf generate
# Fix buf.lock permissions after generation
if [ -f buf.lock ]; then
  echo "Fixing buf.lock permissions"
  chown $(id -u):$(id -g) buf.lock
  chmod 644 buf.lock
fi

cd ..

# move proto files to the right places
cp -r ./github.com/Nolus-Protocol/nolus-core/x/* x/  
rm -rf ./github.com
