#!/bin/bash
set -euxo pipefail

# start "instance" variables
local_chain_id=""
# end "instance" variables

init_vars() {
  local_chain_id="$1"
}

deploy() {
  local root_dir="$1"
  local node_id="$2"
  local chain_id="$3"

  local node_dir="$root_dir/$node_id"
  rm -fr "$node_dir"
  mkdir "$node_dir"

  run_cmd "$node_dir" init "$node_id" --chain-id "$chain_id" 1>/dev/null
}

gen_account() {
  local root_dir="$1"
  local node_id="$2"

  local node_dir="$root_dir/$node_id"

  local add_key_out=$(run_cmd "$node_dir" keys add "$node_id" --keyring-backend test --output json 1>/dev/null)
  #TBD keep the mnemonic if necessary
  #echo "$add_key_out"| jq -r .mnemonic > "$node_dir"/mnemonic
  echo $(run_cmd "$node_dir" keys show -a "$node_id" --keyring-backend test)
}

# outputs the generated create validator transaction to the standard output
gen_validator() {
  local root_dir="$1"
  local node_id="$2"
  local genesis_file="$3"
  local stake="$4"
  # local ip_address="$5"

  local node_dir="$root_dir/$node_id"
  local tx_out_file="$node_dir/config/gentx_out.json"

  cp "$genesis_file" "$node_dir/config/genesis.json"
  # ip_spec=""
  # if [[ -n "${ip_address+}" ]]; then
  #   ip_spec="--ip $ip_address"
  # fi
  # $ip_spec
  run_cmd "$node_dir" gentx "$node_id" "$stake" --keyring-backend test --chain-id "$local_chain_id" --output-document "$tx_out_file" 1>/dev/null
  cat "$tx_out_file"
}
