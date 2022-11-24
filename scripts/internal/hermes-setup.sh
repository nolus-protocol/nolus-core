#!/bin/bash

set -euox pipefail

# Accounts setup
setup_accounts() {
  declare -r nolus_home_dir="$1"
  declare -r nolus_net_address="$2"
  declare -r wallet_with_funds_key="$3"
  declare -r hermes_binary_dir="$4"
  declare -r a_chain="$5"
  declare -r b_chain="$6"
  declare -r hermes_mnemonic="$7"

  declare -r hermes_key="hermes"
  declare -r hermes_mnemonic_file="hermes.seed"
  touch "$hermes_mnemonic_file"
  echo "$hermes_mnemonic" > "$hermes_mnemonic_file"

  run_cmd "$nolus_home_dir" keys add "$hermes_key" --recover < "$hermes_mnemonic_file"
  declare -r hermes_address=$(run_cmd "$nolus_home_dir" keys show "$hermes_key" -a)
  echo 'y' | run_cmd "$nolus_home_dir" tx bank send "$wallet_with_funds_key" "$hermes_address" 2000000unls --fees 500unls --node "$nolus_net_address" --broadcast-mode block

  "$hermes_binary_dir"/hermes keys add --chain "$a_chain" --mnemonic-file "$hermes_mnemonic_file"
  "$hermes_binary_dir"/hermes keys add --chain "$b_chain" --mnemonic-file "$hermes_mnemonic_file"

  rm "$hermes_mnemonic_file"
}

# Open connection
open_connection() {
  declare -r nolusd_home_dir="$1"
  declare -r nolus_net_address="$2"
  declare -r hermes_binary_dir="$3"
  declare -r a_chain="$4"
  declare -r b_chain="$5"
  declare -r connection="$6"

  "$hermes_binary_dir"/hermes create connection --a-chain "$a_chain" --b-chain "$b_chain"
  "$hermes_binary_dir"/hermes create channel --a-chain "$a_chain" --a-connection "$connection" --a-port transfer --b-port transfer --order unordered

  declare -r counterparty_channel_id=$(run_cmd "$nolusd_home_dir" q ibc channel connections "$connection" --node "$nolus_net_address" --output json | jq '.channels[0].counterparty.channel_id' | tr -d '"')
  echo "$counterparty_channel_id"
}