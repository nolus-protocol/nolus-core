#!/bin/bash
# DEX setup

set -euox pipefail

SCRIPTS_DIR=$(cd $(dirname "${BASH_SOURCE[0]}") && cd .. && pwd)
source "$SCRIPTS_DIR"/common/cmd.sh
source "$SCRIPTS_DIR"/internal/accounts.sh

_open_connection() {
  declare -r hermes_binary_dir="$1"
  declare -r a_chain="$2"
  declare -r b_chain="$3"
  declare -r connection="$4"

  "$hermes_binary_dir"/hermes create connection --a-chain "$a_chain" --b-chain "$b_chain"
  "$hermes_binary_dir"/hermes create channel --a-chain "$a_chain" --a-connection "$connection" --a-port transfer --b-port transfer --order unordered
}

leaser_dex_setup() {
declare -r nolus_net_address="$1"
declare -r nolus_home_dir="$2"
declare -r wallet_with_funds_key="$3"
declare -r hermes_binary_dir="$4"
declare -r hermes_address="$5"
declare -r a_chain="$6"
declare -r b_chain="$7"

# Prepare Hermes
declare -r wallet_with_funds_addr=$(run_cmd "$nolus_home_dir" keys show "$wallet_with_funds_key" -a)
declare -r flags="--fees 1000unls --gas auto --gas-adjustment 1.3 --node $nolus_net_address"

echo 'y' | run_cmd "$nolus_home_dir" tx bank send "$wallet_with_funds_addr" "$hermes_address" 2000000unls $flags --broadcast-mode block

declare -r connection="connection-0"
_open_connection "$hermes_binary_dir" "$a_chain" "$b_chain" "$connection"
}