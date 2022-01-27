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
  local val_accounts_dir="$1"
  local validators="$2"
  local chain_id="$3"
  local native_currency="$4"
  local suspend_admin="$5"
  local val_tokens="$6"
  local val_stake="$7"
  
  local accounts_file="$val_accounts_dir/accounts.json"
  local proto_genesis_file="$val_accounts_dir/penultimate-genesis.json"
  local final_genesis_file="$val_accounts_dir/genesis.json"


  init_local_sh "$val_accounts_dir" "$chain_id"
  node_id_and_val_pubkeys="$(__setup_nodes "$validators")"
  val_addrs="$(__gen_val_accounts "$node_id_and_val_pubkeys")"
  __gen_accounts_spec "$val_addrs" "$accounts_file" "$val_tokens"
  generate_proto_genesis "$chain_id" "$accounts_file" "$native_currency" "$proto_genesis_file" "$suspend_admin"
  create_validator_txs="$(__init_validators "$proto_genesis_file" "$node_id_and_val_pubkeys" "$val_stake")"
  integrate_genesis_txs "$proto_genesis_file" "$create_validator_txs" "$final_genesis_file"
  __propagate_genesis_all "$final_genesis_file" "$validators"
}

#####################
# private functions #
#####################

# Setup validator nodes and collect their ids and validator public keys
#
# The nodes are installed and configured depending on the sourced implementation script.
# The node ids and validator public keys are printed on the standard output one at a line.
__setup_nodes() {
  set -euxo pipefail
  local validators="$1"
  for i in $(seq "$validators"); do
    config "$i"
  done
}

__gen_val_accounts() {
  local node_id_and_val_pubkeys="$1"
  for node_id_and_val_pubkey in "$node_id_and_val_pubkeys"; do
    local node_id
    read -r node_id __val_pub_key <<< $node_id_and_val_pubkey
    local address
    address=$(gen_val_account "$node_id")
    echo "$address"
  done
}

__gen_accounts_spec() {
  local val_addrs="$1"
  local file="$2"
  local val_tokens="$3"

  local accounts="[]"
  for address in $val_addrs; do
    accounts=$(echo "$accounts" | add_account "$address" "$val_tokens")
  done
  echo "$accounts" > "$file"
}

__init_validators() {
  local proto_genesis_file="$1"
  local node_id_and_val_pubkeys="$2"
  local val_stake="$3"

  for node_id_and_val_pubkey in "$node_id_and_val_pubkeys"; do
    local node_id
    local val_pub_key
    read -r node_id val_pub_key <<< $node_id_and_val_pubkey
    local create_validator_tx
    create_validator_tx=$(gen_validator "$proto_genesis_file" "$node_id" "$val_pub_key" "$val_stake")
    echo "$create_validator_tx"
  done
}

__propagate_genesis_all() {
  local genesis_file="$1"
  local validators="$2"

  for i in $(seq "$validators"); do
    propagate_genesis "$i" "$genesis_file"
  done
}
