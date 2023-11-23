#!/bin/bash

SCRIPTS_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && cd .. && pwd)"
source "$SCRIPTS_DIR"/remote/lib/lib.sh
source "$SCRIPTS_DIR"/common/cmd.sh

add_new_chain_hermes() {
  declare -r hermes_config_file_path="$1"
  declare -r chain_id="$2"
  declare -r chain_ip_addr_RPC="$3"
  declare -r chain_ip_addr_gRPC="$4"
  declare -r chain_account_prefix="$5"
  declare -r chain_price_denom="$6"
  declare -r chain_trusting_period="$7"
  declare if_interchain_security="$8"

  declare -r chains_count=$(grep -c "\[\[chains\]\]" "$hermes_config_file_path/config.toml")

  declare -r chain_key_name="hermes-$chain_id"

  update_config "$hermes_config_file_path" '.chains['"$chains_count"']."id"' '"'"$chain_id"'"'
  update_config "$hermes_config_file_path" '.chains['"$chains_count"']."rpc_addr"' '"https://'"$chain_ip_addr_RPC"'"'
  update_config "$hermes_config_file_path" '.chains['"$chains_count"']."grpc_addr"' '"https://'"$chain_ip_addr_gRPC"'"'
  update_config "$hermes_config_file_path" '.chains['"$chains_count"']."rpc_timeout"' '"10s"'
  update_config "$hermes_config_file_path" '.chains['"$chains_count"']."account_prefix"' '"'"$chain_account_prefix"'"'
  update_config "$hermes_config_file_path" '.chains['"$chains_count"']."key_name"' '"'"$chain_key_name"'"'
  update_config "$hermes_config_file_path" '.chains['"$chains_count"']."address_type"' '{ derivation : "cosmos" }'
  update_config "$hermes_config_file_path" '.chains['"$chains_count"']."store_prefix"' '"ibc"'
  update_config "$hermes_config_file_path" '.chains['"$chains_count"']."default_gas"' 5000000
  update_config "$hermes_config_file_path" '.chains['"$chains_count"']."max_gas"' 15000000
  update_config "$hermes_config_file_path" '.chains['"$chains_count"']."gas_price"' '{ price : 0.0026, denom : "'"$chain_price_denom"'" }'
  update_config "$hermes_config_file_path" '.chains['"$chains_count"']."gas_multiplier"' 1.1
  update_config "$hermes_config_file_path" '.chains['"$chains_count"']."max_msg_num"' 20
  update_config "$hermes_config_file_path" '.chains['"$chains_count"']."max_tx_size"' 209715
  update_config "$hermes_config_file_path" '.chains['"$chains_count"']."clock_drift"' '"20s"'
  update_config "$hermes_config_file_path" '.chains['"$chains_count"']."max_block_time"' '"10s"'
  update_config "$hermes_config_file_path" '.chains['"$chains_count"']."trusting_period"' '"'"$chain_trusting_period"'s"'
  update_config "$hermes_config_file_path" '.chains['"$chains_count"']."trust_threshold"' '{ numerator : "1", denominator : "3" }'
  update_config "$hermes_config_file_path" '.chains['"$chains_count"']."event_source"' '{ mode : "push", url : "wss://'"$chain_ip_addr_RPC"'/websocket", batch_delay : "500ms" }'

  if [ "$if_interchain_security" == "true" ]
  then
    update_config "$hermes_config_file_path" '.chains['"$chains_count"']."ccv_consumer_chain"' 'true'
  fi
}

dex_account_setup() {
  declare -r hermes_binary_dir_path="$1"
  declare -r chain_id="$2"
  declare hermes_mnemonic_file="$3"

  "$hermes_binary_dir_path"/hermes keys add --chain "$chain_id" --mnemonic-file "$hermes_mnemonic_file"
}


open_connection() {
  declare -r nolus_net_address="$1"
  declare -r nolus_home_dir="$2"
  declare -r account_key_to_feed_hermes_address="$3"
  declare -r hermes_binary_dir_path="$4"
  declare -r hermes_address="$5"
  declare -r nolus_chain="$6"
  declare -r b_chain="$7"

  declare -r account_addr_to_feed_hermes_address=$(run_cmd "$nolus_home_dir" keys show "$account_key_to_feed_hermes_address" -a)
  declare -r flags="--fees 1000unls --gas auto --gas-adjustment 1.3 --node $nolus_net_address"

  echo 'y' | run_cmd "$nolus_home_dir" tx bank send "$account_addr_to_feed_hermes_address" "$hermes_address" 2000000unls $flags --broadcast-mode block

  connection_data_file=$(mktemp)
  "$hermes_binary_dir_path"/hermes create connection --a-chain "$nolus_chain" --b-chain "$b_chain" &>"$connection_data_file"

  declare connection_id
  connection_id=$(grep 'SUCCESS Connection' -A 15000 "$connection_data_file" | grep "$nolus_chain" -A 10 | grep 'ConnectionId' -A 2 | grep 'connection-')
  connection_id=${connection_id//[, ]/}
  connection_id=${connection_id//\"/}

  "$hermes_binary_dir_path"/hermes create channel --a-chain "$nolus_chain" --a-connection "$connection_id" --a-port transfer --b-port transfer --order unordered

  rm "$connection_data_file"
}

get_hermes_address() {
  local -r hermes_binary_dir="$1"
  local -r chain_id="$2"

  local address
  address="$("$hermes_binary_dir"/hermes keys list --chain "$chain_id")"
  # Look for 'Parameter Expansion' at https://linux.die.net/man/1/bash
  # First strip off the prefix
  address="${address#*nolus (}"
  # Then the suffix
  echo "${address%)}"
}
