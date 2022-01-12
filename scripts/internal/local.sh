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

deploy() {
  local node_id="$1"
  local node_dir=$(node_dir $node_id)
  rm -fr "$node_dir"
  mkdir "$node_dir"

  run_cmd "$node_dir" init "$node_id" --chain-id "$local_chain_id" 1>/dev/null
}

gen_account() {
  local node_id="$1"
  local node_dir=$(node_dir $node_id)

  local add_key_out=$(run_cmd "$node_dir" keys add "$node_id" --keyring-backend test --output json 1>/dev/null)
  #TBD keep the mnemonic if necessary
  #echo "$add_key_out"| jq -r .mnemonic > "$node_dir"/mnemonic
  echo $(run_cmd "$node_dir" keys show -a "$node_id" --keyring-backend test)
}

# outputs the generated create validator transaction to the standard output
gen_validator() {
  local node_id="$1"
  local genesis_file="$2"
  local stake="$3"
  # local ip_address="$5"

  local node_dir=$(node_dir $node_id)
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
  local node_id="$1"
  local genesis_file="$2"

  cp "$genesis_file" "$(node_dir $node_id)/config/genesis.json"
}

#####################
# private functions #
#####################
node_dir() {
  echo "$local_root_dir/$1"
}