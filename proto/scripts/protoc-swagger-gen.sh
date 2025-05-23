#!/usr/bin/env bash

set -euxo pipefail

mkdir -p tmp_deps

#copy some deps to use their proto files to generate swagger
declare -a deps=(
    "github.com/cosmos/cosmos-sdk"
    "github.com/CosmWasm/wasmd"
    "github.com/cosmos/ibc-go/v10"
    "github.com/cometbft/cometbft"
)

for dep in "${deps[@]}"
do
    go mod download $dep
    path=$(go list -f '{{ .Dir }}' -m $dep); \
    cp -r $path tmp_deps; \
done

proto_dirs=$(find ./proto ./tmp_deps -path -prune -o -name '*.proto' -print0 | xargs -0 -n1 dirname | sort | uniq)
for dir in $proto_dirs; do

  # generate swagger files (filter query files)
  query_file=$(find "${dir}" -maxdepth 1 \( -name 'query.proto' -o -name 'service.proto' \))
  if [[ ! -z "$query_file" ]]; then
    buf generate --template proto/buf.gen.swagger.yaml $query_file
  fi
done

# Fix circular definition in cosmos b just removing them
# jq 'del(.definitions["cosmos.tx.v1beta1.ModeInfo.Multi"].properties.mode_infos.items["$ref"])' ./tmp-swagger-gen/cosmos/tx/v1beta1/service.swagger.json > ./tmp-swagger-gen/cosmos/tx/v1beta1/fixed_service.swagger.json
# jq 'del(.definitions["cosmos.autocli.v1.ServiceCommandDescriptor"].properties.sub_commands)' ./tmp-swagger-gen/cosmos/autocli/v1/query.swagger.json > ./tmp-swagger-gen/cosmos/autocli/v1/fixed_query.swagger.json

# rm -rf ./tmp-swagger-gen/cosmos/tx/v1beta1/service.swagger.json
# rm -rf ./tmp-swagger-gen/cosmos/autocli/v1/query.swagger.json

# Convert all *.swagger.json files into a single folder _all
files=$(find ./tmp-swagger-gen -name '*.swagger.json' -print0 | xargs -0)
mkdir -p ./tmp-swagger-gen/_all
counter=0
for f in $files; do
  echo "[+] $f"

  # check gaia first before cosmos
  if [[ "$f" =~ "router" ]]; then
    cp $f ./tmp-swagger-gen/_all/pfm-$counter.json
  elif [[ "$f" =~ "cosmwasm" ]]; then
    cp $f ./tmp-swagger-gen/_all/cosmwasm-$counter.json
  elif [[ "$f" =~ "osmosis" ]]; then
    cp $f ./tmp-swagger-gen/_all/osmosis-$counter.json
  elif [[ "$f" =~ "juno" ]]; then
    cp $f ./tmp-swagger-gen/_all/juno-$counter.json
  elif [[ "$f" =~ "cosmos" ]]; then
    cp $f ./tmp-swagger-gen/_all/cosmos-$counter.json
  else
    cp $f ./tmp-swagger-gen/_all/other-$counter.json
  fi

  counter=$(expr $counter + 1)
done

# merges all the above into FINAL.json
python3 ./scripts/swagger_merger.py

# Makes a swagger temp file with reference pointers
swagger-combine ./tmp-swagger-gen/FINAL.json -o ./tmp-swagger-gen/tmp_swagger.yaml -f yaml --continueOnConflictingPaths true --includeDefinitions true

# extends out the *ref instances to their full value
swagger-merger --input ./tmp-swagger-gen/tmp_swagger.yaml -o ./docs/swagger/swagger.yaml

rm -rf tmp-swagger-gen
rm -rf tmp_deps
