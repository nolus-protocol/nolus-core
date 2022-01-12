#!/bin/bash
set -euxo pipefail

integrate_genesis_txs() {
  local genesis_home_dir="$1"
  local genesis_in_file="$2"
  local txs="$3"
  local genesis_out_file="$4"

  local genesis_basedir="$genesis_home_dir"/config
  local genesis_file="$genesis_basedir"/genesis.json
  mkdir "$genesis_basedir"
  cp "$genesis_in_file" "$genesis_file"

  local txs_dir="$genesis_home_dir"/txs
  {
    mkdir "$txs_dir"
    local index=0
    for tx in $txs; do
        echo "$tx" > "$txs_dir"/tx"$index".json
        index=$((index+1))
    done
  }

  run_cmd "$genesis_home_dir" collect-gentxs --gentx-dir "$txs_dir"
  cp "$genesis_file" "$genesis_out_file"
}
