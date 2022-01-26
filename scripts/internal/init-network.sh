#!/bin/bash
set -euxo pipefail

cleanup_init_network_sh() {
  cleanup_genesis_sh
}

SCRIPT_DIR=$(cd $(dirname "${BASH_SOURCE[0]}") && pwd)

source "$SCRIPT_DIR"/local.sh
source "$SCRIPT_DIR"/accounts.sh
source "$SCRIPT_DIR"/genesis.sh

init_network() {
  local output_dir="$1"
  local validators="$2"
  local chain_id="$3"
  local native_currency="$4"
  local suspend_admin="$5"
  local val_tokens="$6"
  local val_stake="$7"
  
  local accounts_file="$output_dir/accounts.json"
  local proto_genesis_file="$output_dir/penultimate-genesis.json"
  local final_genesis_file="$output_dir/genesis.json"


  init_local_sh "$output_dir" "$chain_id"
  addresses="$(init_nodes "$validators")"
  gen_accounts_spec "$addresses" "$accounts_file" "$val_tokens"
  generate_proto_genesis "$chain_id" "$accounts_file" "$native_currency" "$proto_genesis_file" "$suspend_admin"
  create_validator_txs="$(init_validators "$proto_genesis_file" "$validators" "$val_stake")"
  integrate_genesis_txs "$proto_genesis_file" "$create_validator_txs" "$final_genesis_file"
  propagate_genesis_all "$final_genesis_file" "$validators"
}

#####################
# private functions #
#####################

# Init validator nodes, generate validator accounts and collect their addresses
#
# The nodes are placed in sub directories of $OUTPUT_DIR
# The validator addresses are printed on the standard output one at a line
init_nodes() {
  local validators="$1"
  for i in $(seq "$validators"); do
    config "$i"
    local address
    address=$(gen_account "$i")
    echo "$address"
  done
}

gen_accounts_spec() {
  local addresses="$1"
  local file="$2"
  local val_tokens="$3"

  local accounts="[]"
  for address in $addresses; do
    accounts=$(echo "$accounts" | add_account "$address" "$val_tokens")
  done
  echo "$accounts" > "$file"
}

init_validators() {
  local proto_genesis_file="$1"
  local validators="$2"
  local val_stake="$3"

  for i in $(seq "$validators"); do
    local create_validator_tx
    create_validator_tx=$(gen_validator "$i" "$proto_genesis_file" "$val_stake")
    echo "$create_validator_tx"
  done
}

propagate_genesis_all() {
  local genesis_file="$1"
  local validators="$2"

  for i in $(seq "$validators"); do
    propagate_genesis "$i" "$genesis_file"
  done
}

## validate dependencies are installed
command -v jq >/dev/null 2>&1 || {
  echo >&2 "jq not installed. More info: https://stedolan.github.io/jq/download/"
  exit 1
}
