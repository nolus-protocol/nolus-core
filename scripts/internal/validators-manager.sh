#!/bin/bash

SCRIPT_DIR=$(cd $(dirname "${BASH_SOURCE[0]}") && pwd)
source "$SCRIPT_DIR"/../common/cmd.sh

# start "instance" variables
val_mngr_home_dir=""
val_mngr_chain_id=""
# end "instance" variables

init_val_mngr_sh() {
  val_mngr_home_dir="$1"
  val_mngr_chain_id="$2"

  rm -fr "$val_mngr_home_dir"
  mkdir -p "$val_mngr_home_dir"

  run_cmd "$val_mngr_home_dir" config chain-id "$val_mngr_chain_id"
  run_cmd "$val_mngr_home_dir" config keyring-backend test
}

gen_val_account() {
  local account_name="$1"

  run_cmd "$val_mngr_home_dir" keys add "$account_name" --output json 1>/dev/null
  run_cmd "$val_mngr_home_dir" keys show -a "$account_name"
}

# outputs the generated create validator transaction to the standard output
gen_val_txn() {
  local genesis_file="$1"
  local val_account_name="$2"
  local val_pub_key="$3"
  local stake="$4"

  local tx_out_file="$val_mngr_home_dir/config/gentx_out_$val_account_name.json"

  cp "$genesis_file" "$val_mngr_home_dir/config/genesis.json"
  # ip_spec=""
  # if [[ -n "${ip_address+}" ]]; then
  #   ip_spec="--ip $ip_address"
  # fi
  # $ip_spec
  run_cmd "$val_mngr_home_dir" gentx "$val_account_name" "$stake" --pubkey "$val_pub_key" --chain-id "$val_mngr_chain_id" \
        --moniker "$val_account_name" --output-document "$tx_out_file" 1>/dev/null
  cat "$tx_out_file"
}
