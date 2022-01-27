#!/bin/bash
set -euxo pipefail

SCRIPT_DIR=$(cd $(dirname "${BASH_SOURCE[0]}") && pwd)
source "$SCRIPT_DIR"/cmd.sh

# start "instance" variables
local_val_accounts_dir=""
local_chain_id=""
# end "instance" variables

init_local_sh() {
  local_val_accounts_dir="$1"
  local_chain_id="$2"

  rm -fr "$local_val_accounts_dir"
  mkdir -p "$local_val_accounts_dir"
  
  run_cmd "$local_val_accounts_dir" config chain-id "$local_chain_id"
  run_cmd "$local_val_accounts_dir" config keyring-backend test
}

gen_val_account() {
  local node_id="$1"

  run_cmd "$local_val_accounts_dir" keys add "$node_id" --output json 1>/dev/null
  run_cmd "$local_val_accounts_dir" keys show -a "$node_id"
}

# outputs the generated create validator transaction to the standard output
gen_validator() {
  local genesis_file="$1"
  local node_id="$2"
  local val_pub_key="$3"
  local stake="$4"

  local tx_out_file="$local_val_accounts_dir/config/gentx_out_$node_id.json"

  cp "$genesis_file" "$local_val_accounts_dir/config/genesis.json"
  # ip_spec=""
  # if [[ -n "${ip_address+}" ]]; then
  #   ip_spec="--ip $ip_address"
  # fi
  # $ip_spec
  run_cmd "$local_val_accounts_dir" gentx "$node_id" "$stake" --pubkey "$val_pub_key" --chain-id "$local_chain_id" \
        --moniker "$node_id" --output-document "$tx_out_file" 1>/dev/null
  cat "$tx_out_file"
}
