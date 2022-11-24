#!/bin/bash
# Install and Configure Hermes
#
# arg: Hermes working directory path, mandatory
# arg: chain1 id, mandatory
# arg: chain1 IP address, mandatory
# arg: chain1 RPC port, mandatory
# arg: chain1 GRPC port, mandatory
# arg: chain2 id, mandatory
# arg: chain2 IP address, mandatory
# arg: chain2 RPC port, mandatory
# arg: chain2 GRPC port, mandatory

set -euox pipefail

SCRIPTS_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && cd .. && pwd)"
source "$SCRIPTS_DIR"/remote/lib/lib.sh
source "$SCRIPTS_DIR"/common/cmd.sh

declare -r hermes_root="$1"
declare -r chain1id="$2"
declare -r chain1IpAddr="$3"
declare -r chain1rpcPort="$4"
declare -r chain1grpcPort="$5"
declare -r chain2id="$6"
declare -r chain2IpAddr="$7"
declare -r chain2rpcPort="$8"
declare -r chain2grpcPort="$9"

# Install

ARCHIVE_NAME="hermes-binary.tar.gz"
wget -O "$ARCHIVE_NAME" https://github.com/informalsystems/hermes/releases/download/v1.1.0/hermes-v1.1.0-x86_64-unknown-linux-gnu.tar.gz

declare -r hermes_binary_dir="$hermes_root"/hermes
mkdir -p  "$hermes_binary_dir"
tar -C "$hermes_binary_dir" -vxzf "$ARCHIVE_NAME"
rm "$ARCHIVE_NAME"

# Configure

declare -r hermes_config_dir="$HOME"/.hermes
mkdir -p "$hermes_config_dir"
config_file="$hermes_config_dir"/config.toml
touch "$config_file"

update_config "$hermes_config_dir" '.mode.clients."enabled"' "true"
update_config "$hermes_config_dir" '.mode.clients."refresh"' "true"
update_config "$hermes_config_dir" '.mode.clients."misbehaviour"' "true"
update_config "$hermes_config_dir" '.mode.connections."enabled"' "true"
update_config "$hermes_config_dir" '.mode.channels."enabled"' "true"
update_config "$hermes_config_dir" '.mode.packets."enabled"' "true"

update_config "$hermes_config_dir" '.chains[0]."id"' '"'"$chain1id"'"'
update_config "$hermes_config_dir" '.chains[0]."rpc_addr"' '"http://'"$chain1IpAddr"':'"$chain1rpcPort"'"'
update_config "$hermes_config_dir" '.chains[0]."grpc_addr"' '"http://'"$chain1IpAddr"':'"$chain1grpcPort"'"'
update_config "$hermes_config_dir" '.chains[0]."websocket_addr"' '"ws://127.0.0.1:'"$chain1rpcPort"'/websocket"'
update_config "$hermes_config_dir" '.chains[0]."rpc_timeout"' '"10s"'
update_config "$hermes_config_dir" '.chains[0]."account_prefix"' '"nolus"'
update_config "$hermes_config_dir" '.chains[0]."key_name"' '"hermes-nolus"'
update_config "$hermes_config_dir" '.chains[0]."address_type"' '{ derivation : "cosmos" }'
update_config "$hermes_config_dir" '.chains[0]."store_prefix"' '"ibc"'
update_config "$hermes_config_dir" '.chains[0]."default_gas"' 100000
update_config "$hermes_config_dir" '.chains[0]."max_gas"' 400000
update_config "$hermes_config_dir" '.chains[0]."gas_price"' '{ price : 0.0025, denom : "unls" }'
update_config "$hermes_config_dir" '.chains[0]."gas_multiplier"' 1.1
update_config "$hermes_config_dir" '.chains[0]."max_msg_num"' 30
update_config "$hermes_config_dir" '.chains[0]."max_tx_size"' 2097152
update_config "$hermes_config_dir" '.chains[0]."clock_drift"' '"5s"'
update_config "$hermes_config_dir" '.chains[0]."max_block_time"' '"30s"'
update_config "$hermes_config_dir" '.chains[0]."trusting_period"' '"14days"'
update_config "$hermes_config_dir" '.chains[0]."trust_threshold"' '{ numerator : "1", denominator : "3" }'
update_config "$hermes_config_dir" '.chains[0]."memo_prefix"' '"''"'

update_config "$hermes_config_dir" '.chains[1]."id"' '"'"$chain2id"'"'
update_config "$hermes_config_dir" '.chains[1]."rpc_addr"' '"http://'"$chain2IpAddr"':'"$chain2rpcPort"'"'
update_config "$hermes_config_dir" '.chains[1]."grpc_addr"' '"http://'"$chain2IpAddr"':'"$chain2grpcPort"'"'
update_config "$hermes_config_dir" '.chains[1]."websocket_addr"' '"ws://'"$chain2IpAddr"':'"$chain2rpcPort"'/websocket"'
update_config "$hermes_config_dir" '.chains[1]."rpc_timeout"' '"10s"'
update_config "$hermes_config_dir" '.chains[1]."account_prefix"' '"osmo"'
update_config "$hermes_config_dir" '.chains[1]."key_name"' '"hermes-osmosis"'
update_config "$hermes_config_dir" '.chains[1]."address_type"' '{ derivation : "cosmos" }'
update_config "$hermes_config_dir" '.chains[1]."store_prefix"' '"ibc"'
update_config "$hermes_config_dir" '.chains[1]."default_gas"' 5000000
update_config "$hermes_config_dir" '.chains[1]."max_gas"' 15000000
update_config "$hermes_config_dir" '.chains[1]."gas_price"' '{ price : 0.0026, denom : "uosmo" }'
update_config "$hermes_config_dir" '.chains[1]."gas_multiplier"' 1.1
update_config "$hermes_config_dir" '.chains[1]."max_msg_num"' 20
update_config "$hermes_config_dir" '.chains[1]."max_tx_size"' 209715
update_config "$hermes_config_dir" '.chains[1]."clock_drift"' '"20s"'
update_config "$hermes_config_dir" '.chains[1]."max_block_time"' '"10s"'
update_config "$hermes_config_dir" '.chains[1]."trusting_period"' '"10days"'
update_config "$hermes_config_dir" '.chains[1]."trust_threshold"' '{ numerator : "1", denominator : "3" }'