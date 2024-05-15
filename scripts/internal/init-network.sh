#!/bin/bash

cleanup_init_network_sh() {
  cleanup_genesis_sh
}

SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
source "$SCRIPT_DIR"/genesis.sh

# arg1: an existing local dir where validator accounts are created, mandatory
init_network() {
  local -r val_accounts_dir="$1"
  local -r validators="$2"
  local -r minimum_gas_price="$3"
  local -r query_gas_limit="$4"
  local -r chain_id="$5"
  local -r native_currency="$6"
  local -r val_tokens="$7"
  local -r val_stake="$8"
  local genesis_accounts_spec="$9"
  local -r wasm_script_path="${10}"
  local -r wasm_code_path="${11}"
  local -r treasury_init_tokens_u128="${12}"
  local -r gov_voting_period="${13}"
  local -r gov_max_deposit_period="${14}"
  local -r staking_max_validators="${15}"
  local -r feerefunder_ack_fee_min="${16}"
  local -r feerefunder_timeout_fee_min="${17}"
  local -r dex_admin_mnemonic="${18}"
  local -r store_code_privileged_account_mnemonic="${19}"
  local -r admins_tokens="${20}"

  node_id_and_val_pubkeys="$(setup_validators "$validators" "$minimum_gas_price" "$query_gas_limit")"
  local final_genesis_file;
  final_genesis_file=$(generate_genesis "$chain_id" "$native_currency" \
                                          "$val_tokens" "$val_stake" \
                                          "$val_accounts_dir" "$genesis_accounts_spec" \
                                          "$wasm_script_path" "$wasm_code_path" \
                                          "$treasury_init_tokens_u128" \
                                          "$node_id_and_val_pubkeys" \
                                          "$gov_voting_period" "$gov_max_deposit_period" "$staking_max_validators" \
                                          "$feerefunder_ack_fee_min" "$feerefunder_timeout_fee_min" \
                                          "$dex_admin_mnemonic" "$store_code_privileged_account_mnemonic" "$admins_tokens")
  propagate_genesis "$final_genesis_file" "$validators"
}