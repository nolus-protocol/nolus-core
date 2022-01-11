#!/bin/bash
set -euxo pipefail

integrate_genesis_txs() {
  local genesis_in_file="$1"
  local txs="$2"
  local genesis_out_file="$3"

  local work_dir=$(mktemp -d)
  local genesis_file="$work_dir"/config/genesis.json
  mkdir "$work_dir"/config
  cp "$genesis_in_file" "$genesis_file"

  local txs_dir="$work_dir"/txs
  {
    mkdir "$txs_dir"
    local index=0
    for tx in $txs; do
        echo "$tx" > "$txs_dir"/tx"$index".json
        index=$((index+1))
    done
  }

  run_cmd "$work_dir" collect-gentxs --gentx-dir "$txs_dir"
  cp "$genesis_file" "$genesis_out_file"
  
  rm -fr "$work_dir"
}
