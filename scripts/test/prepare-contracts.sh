#!/bin/sh

wasmd_dir=$(go list -f "{{ .Dir }}" -m github.com/CosmWasm/wasmd)
reflect=${wasmd_dir}/x/wasm/keeper/testdata/reflect.wasm
if test ! -f ${reflect}; then
    echo "Reflect contract not found in CosmWasm at ${reflect}"
    exit 1
fi

cp "${reflect}" testdata/
exit 0
