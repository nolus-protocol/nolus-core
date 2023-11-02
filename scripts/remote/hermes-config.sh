#!/bin/bash
# Install and Configure Hermes
#
# arg: Hermes working directory path, mandatory
# arg: chain1 id, mandatory
# arg: chain1 IP address, mandatory
# arg: chain1 RPC port, mandatory
# arg: chain1 GRPC port, mandatory
# arg: chain2 id, mandatory
# arg: chain2 IP address - rpc , mandatory
# arg: chain2 IP address - grpc, mandatory
# arg: hermes account seed, mandatory

set -euox pipefail

SCRIPTS_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && cd .. && pwd)"
source "$SCRIPTS_DIR"/remote/lib/lib.sh
source "$SCRIPTS_DIR"/common/cmd.sh
source "$SCRIPTS_DIR"/common/rm-dir.sh

declare -r hermes_root="$1"
declare -r chain1id="$2"
declare -r chain1IpAddr="$3"
declare -r chain1rpcPort="$4"
declare -r chain1grpcPort="$5"
declare -r chain2id="$6"
declare -r chain2IpAddr_RPC="$7"
declare -r chain2IpAddr_gRPC="$8"
declare -r chain2accountPrefix="$9"
declare -r chain2priceDenom=${10}
declare -r chain2trustingPeriod=${11}
declare -r hermes_mnemonic=${12}

# Install

declare -r archive_name="hermes-binary.tar.gz"
wget -O "$archive_name" https://github.com/informalsystems/hermes/releases/download/v1.6.0/hermes-v1.6.0-x86_64-unknown-linux-gnu.tar.gz

declare -r hermes_binary_dir="$hermes_root"/hermes
mkdir -p  "$hermes_binary_dir"
tar -C "$hermes_binary_dir" -vxzf "$archive_name"
rm "$archive_name"

# Configure

declare -r chain1keyName="hermes-nolus"
declare -r chain2keyName="hermes-$chain2id"

declare -r hermes_config_dir="$HOME"/.hermes
rm_dir "$hermes_config_dir"
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
update_config "$hermes_config_dir" '.chains[0]."rpc_timeout"' '"10s"'
update_config "$hermes_config_dir" '.chains[0]."account_prefix"' '"nolus"'
update_config "$hermes_config_dir" '.chains[0]."key_name"' '"'"$chain1keyName"'"'
update_config "$hermes_config_dir" '.chains[0]."address_type"' '{ derivation : "cosmos" }'
update_config "$hermes_config_dir" '.chains[0]."store_prefix"' '"ibc"'
update_config "$hermes_config_dir" '.chains[0]."default_gas"' 1000000
update_config "$hermes_config_dir" '.chains[0]."max_gas"' 4000000
update_config "$hermes_config_dir" '.chains[0]."gas_price"' '{ price : 0.0025, denom : "unls" }'
update_config "$hermes_config_dir" '.chains[0]."gas_multiplier"' 1.1
update_config "$hermes_config_dir" '.chains[0]."max_msg_num"' 30
update_config "$hermes_config_dir" '.chains[0]."max_tx_size"' 2097152
update_config "$hermes_config_dir" '.chains[0]."clock_drift"' '"5s"'
update_config "$hermes_config_dir" '.chains[0]."max_block_time"' '"30s"'
update_config "$hermes_config_dir" '.chains[0]."trusting_period"' '"14days"'
update_config "$hermes_config_dir" '.chains[0]."trust_threshold"' '{ numerator : "1", denominator : "3" }'
update_config "$hermes_config_dir" '.chains[0]."memo_prefix"' '"''"'
update_config "$hermes_config_dir" '.chains[0]."event_source"' '{ mode : "push", url : "ws://127.0.0.1:'"$chain1rpcPort"'/websocket", batch_delay : "500ms" }'

update_config "$hermes_config_dir" '.chains[1]."id"' '"'"$chain2id"'"'
update_config "$hermes_config_dir" '.chains[1]."rpc_addr"' '"https://'"$chain2IpAddr_RPC"'"'
update_config "$hermes_config_dir" '.chains[1]."grpc_addr"' '"https://'"$chain2IpAddr_gRPC"'"'
update_config "$hermes_config_dir" '.chains[1]."rpc_timeout"' '"10s"'
update_config "$hermes_config_dir" '.chains[1]."account_prefix"' '"'"$chain2accountPrefix"'"'
update_config "$hermes_config_dir" '.chains[1]."key_name"' '"'"$chain2keyName"'"'
update_config "$hermes_config_dir" '.chains[1]."address_type"' '{ derivation : "cosmos" }'
update_config "$hermes_config_dir" '.chains[1]."store_prefix"' '"ibc"'
update_config "$hermes_config_dir" '.chains[1]."default_gas"' 5000000
update_config "$hermes_config_dir" '.chains[1]."max_gas"' 15000000
update_config "$hermes_config_dir" '.chains[1]."gas_price"' '{ price : 0.0026, denom : "'"$chain2priceDenom"'" }'
update_config "$hermes_config_dir" '.chains[1]."gas_multiplier"' 1.1
update_config "$hermes_config_dir" '.chains[1]."max_msg_num"' 20
update_config "$hermes_config_dir" '.chains[1]."max_tx_size"' 209715
update_config "$hermes_config_dir" '.chains[1]."clock_drift"' '"20s"'
update_config "$hermes_config_dir" '.chains[1]."max_block_time"' '"10s"'
update_config "$hermes_config_dir" '.chains[1]."trusting_period"' '"'"$chain2trustingPeriod"'s"'
update_config "$hermes_config_dir" '.chains[1]."trust_threshold"' '{ numerator : "1", denominator : "3" }'
update_config "$hermes_config_dir" '.chains[1]."event_source"' '{ mode : "push", url : "wss://'"$chain2IpAddr_RPC"'/websocket", batch_delay : "500ms" }'

# Accounts setup
declare -r hermes_mnemonic_file="$hermes_config_dir"/hermes.seed
echo "$hermes_mnemonic" > "$hermes_mnemonic_file"

"$hermes_binary_dir"/hermes keys add --chain "$chain1id" --mnemonic-file "$hermes_mnemonic_file"
"$hermes_binary_dir"/hermes keys add --chain "$chain2id" --mnemonic-file "$hermes_mnemonic_file"
