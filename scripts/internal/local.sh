#!/bin/bash
set -euxo pipefail

# start "instance" variables
local_root_dir=""
local_chain_id=""
# end "instance" variables

init_local_sh() {
  local_root_dir="$1"
  local_chain_id="$2"

  rm -fr "$local_root_dir"
  mkdir -p "$local_root_dir"
}

gen_account() {
  local node_index="$1"
  local node_dir
  node_dir=$(node_dir "$node_index")
  local node_id
  node_id=$(node_id "$node_index")

  run_cmd "$node_dir" keys add "$node_id" --keyring-backend test --output json 1>/dev/null
  run_cmd "$node_dir" keys show -a "$node_id" --keyring-backend test
}

# outputs the generated create validator transaction to the standard output
gen_validator() {
  local node_index="$1"
  local genesis_file="$2"
  local stake="$3"
  # local ip_address="$5"
  local node_dir
  node_dir=$(node_dir "$node_index")
  local node_id
  node_id=$(node_id "$node_index")

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

propagate_genesis() {
  local node_index="$1"
  local genesis_file="$2"

  cp "$genesis_file" "$(node_dir "$node_index")/config/genesis.json"
}

#####################
# private functions #
#####################
node_dir() {
  local node_index=$1
  local node_id
  node_id=$(node_id "$node_index")
  echo "$local_root_dir/$node_id"
}

node_id() {
  echo "dev-validator-$1"
}