#!/bin/bash
set -euxo pipefail

CMD="cosmzoned"

deploy() {
  local root_dir="$1"
  local node_id="$2"
  local chain_id="$3"

  local node_dir="$root_dir/$node_id"
  rm -fr "$node_dir"
  mkdir "$node_dir"

  run_cmd "$node_dir" init "$node_id" --chain-id "$chain_id"
}

gen_account() {
  local root_dir="$1"
  local node_id="$2"

  local node_dir="$root_dir/$node_id"

  local add_key_out=$(run_cmd "$node_dir" keys add "$node_id" --keyring-backend test --output json)
  #TBD keep the mnemonic if necessary
  #echo "$add_key_out"| jq -r .mnemonic > "$node_dir"/mnemonic
  echo $(run_cmd "$node_dir" keys show -a "$node_id" --keyring-backend test)
}

run_cmd() {
  local home="$1"
  shift

  "$CMD" $@ --home "$home" 2>&1
}

command -v "$CMD" >/dev/null 2>&1 || {
  echo >&2 "$CMD is not found in \$PATH."
  exit 1
}
