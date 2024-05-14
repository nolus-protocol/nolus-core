#!/bin/bash

SCRIPTS_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && cd .. && pwd)"
source "$SCRIPTS_DIR"/remote/lib/lib.sh
source "$SCRIPTS_DIR"/common/cmd.sh
source "$SCRIPTS_DIR"/internal/wait_services.sh

add_new_chain_hermes() {
  declare -r hermes_config_file_path="$1"
  declare -r chain_id="$2"
  declare -r chain_ip_addr_RPC="$3"
  declare -r chain_ip_addr_gRPC="$4"
  declare -r chain_rpc_timeout_secs="$5"
  declare -r chain_account_prefix="$6"
  declare -r chain_default_gas="$7"
  declare -r chain_max_gas="$8"
  declare -r chain_gas_price_price="$9"
  declare -r chain_gas_price_denom="${10}"
  declare -r chain_gas_multiplier="${11}"
  declare -r chain_max_msg_num="${12}"
  declare -r chain_max_tx_size="${13}"
  declare -r chain_clock_drift_secs="${14}"
  declare -r chain_max_block_time_secs="${15}"
  declare -r chain_trusting_period_secs="${16}"
  declare -r chain_trust_threshold_numerator="${17}"
  declare -r chain_trust_threshold_denumerator="${18}"
  declare if_interchain_security="${19}"

  declare -r chains_count=$(grep -c "\[\[chains\]\]" "$hermes_config_file_path/config.toml")

  declare -r chain_key_name="hermes-$chain_id"

  update_config "$hermes_config_file_path" '.chains['"$chains_count"']."id"' '"'"$chain_id"'"'
  update_config "$hermes_config_file_path" '.chains['"$chains_count"']."rpc_addr"' '"https://'"$chain_ip_addr_RPC"'"'
  update_config "$hermes_config_file_path" '.chains['"$chains_count"']."grpc_addr"' '"https://'"$chain_ip_addr_gRPC"'"'
  update_config "$hermes_config_file_path" '.chains['"$chains_count"']."rpc_timeout"' '"'"$chain_rpc_timeout_secs"'s"'
  update_config "$hermes_config_file_path" '.chains['"$chains_count"']."account_prefix"' '"'"$chain_account_prefix"'"'
  update_config "$hermes_config_file_path" '.chains['"$chains_count"']."key_name"' '"'"$chain_key_name"'"'
  update_config "$hermes_config_file_path" '.chains['"$chains_count"']."address_type"' '{ derivation : "cosmos" }'
  update_config "$hermes_config_file_path" '.chains['"$chains_count"']."store_prefix"' '"ibc"'
  update_config "$hermes_config_file_path" '.chains['"$chains_count"']."default_gas"' "$chain_default_gas"
  update_config "$hermes_config_file_path" '.chains['"$chains_count"']."max_gas"' "$chain_max_gas"
  update_config "$hermes_config_file_path" '.chains['"$chains_count"']."gas_price"' '{ price : '"$chain_gas_price_price"', denom : "'"$chain_gas_price_denom"'" }'
  update_config "$hermes_config_file_path" '.chains['"$chains_count"']."gas_multiplier"' "$chain_gas_multiplier"
  update_config "$hermes_config_file_path" '.chains['"$chains_count"']."max_msg_num"' "$chain_max_msg_num"
  update_config "$hermes_config_file_path" '.chains['"$chains_count"']."max_tx_size"' "$chain_max_tx_size"
  update_config "$hermes_config_file_path" '.chains['"$chains_count"']."clock_drift"' '"'"$chain_clock_drift_secs"'s"'
  update_config "$hermes_config_file_path" '.chains['"$chains_count"']."max_block_time"' '"'"$chain_max_block_time_secs"'s"'
  update_config "$hermes_config_file_path" '.chains['"$chains_count"']."trusting_period"' '"'"$chain_trusting_period_secs"'s"'
  update_config "$hermes_config_file_path" '.chains['"$chains_count"']."trust_threshold"' '{ numerator : '"$chain_trust_threshold_numerator"', denominator : '"$chain_trust_threshold_denumerator"' }'
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

  declare tx_result
  tx_result=$(echo 'y' | run_cmd "$nolus_home_dir" tx bank send "$account_addr_to_feed_hermes_address" "$hermes_address" 2000000unls $flags --output json)
  tx_result=$(echo "$tx_result" | awk 'NR > 1')
  tx_result=$(echo "$tx_result" | jq -c '.')
  local tx_hash
  tx_hash=$(echo "$tx_result" | jq -r '.txhash')
  tx_hash=$(echo "$tx_hash" | sed '/^null$/d')

  wait_tx_included_in_block "$nolus_home_dir" "$nolus_net_address" "$tx_hash"

  connection_data_file=$(mktemp)
  "$hermes_binary_dir_path"/hermes create connection --a-chain "$nolus_chain" --b-chain "$b_chain" > "$connection_data_file"

  export CONNECTION_ID
  CONNECTION_ID=$(grep 'SUCCESS Connection' -A 15000 "$connection_data_file" | grep "$nolus_chain" -A 10 | grep 'ConnectionId' -A 2 | grep 'connection-')
  CONNECTION_ID=${CONNECTION_ID//[, ]/}
  CONNECTION_ID=${CONNECTION_ID//\"/}
  rm "$connection_data_file"

  open_channel "$nolus_chain" "$CONNECTION_ID"
}

open_channel() {
  local -r nolus_chain="$1"
  local -r connection_id="$2"

  "$hermes_binary_dir_path"/hermes create channel --a-chain "$nolus_chain" --a-connection "$connection_id" --a-port transfer --b-port transfer --order unordered
}

get_connection_info() {
  local -r nolus_home_dir="$1"
  local -r connection_id="$2"

  local -r connection_info=$(run_cmd "$nolus_home_dir" q ibc channel connections "$connection_id" --output json)

  echo "$connection_info"
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
