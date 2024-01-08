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
  local -r chain_id="$3"
  local -r native_currency="$4"
  local -r val_tokens="$5"
  local -r val_stake="$6"
  local genesis_accounts_spec="$7"
  local -r wasm_script_path="$8"
  local -r wasm_code_path="$9"
  local -r treasury_init_tokens_u128="${10}"
  local -r gov_voting_period="${11}"
  local -r feerefunder_ack_fee_min="${12}"
  local -r feerefunder_timeout_fee_min="${13}"
  local -r dex_admin_mnemonic="${14}"
  local -r store_code_privileged_account_mnemonic="${15}"
  local -r admins_tokens="${16}"

  node_id_and_val_pubkeys="$(setup_validators "$validators")"
  local final_genesis_file;
  final_genesis_file=$(generate_genesis "$chain_id" "$native_currency" \
                                          "$val_tokens" "$val_stake" \
                                          "$val_accounts_dir" "$genesis_accounts_spec" \
                                          "$wasm_script_path" "$wasm_code_path" \
                                          "$treasury_init_tokens_u128" \
                                          "$node_id_and_val_pubkeys" \
                                          "$gov_voting_period" "$feerefunder_ack_fee_min" "$feerefunder_timeout_fee_min" \
                                          "$dex_admin_mnemonic" "$store_code_privileged_account_mnemonic" "$admins_tokens")
  propagate_genesis "$final_genesis_file" "$validators"
}