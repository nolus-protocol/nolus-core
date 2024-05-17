#!/bin/bash
# Install and Configure Hermes
#
# $1: Hermes working directory path (mandatory)
# $2: Chain ID(mandatory)
# $3: IP address (mandatory)
# $4: RPC port (mandatory)
# $5: gRPC port (mandatory)
# $6: Timeout for RPC requests (in seconds) (mandatory)
# $7: Default gas amount for transactions (mandatory)
# $8: Maximum gas amount for transactions (mandatory)
# $9: Gas price for transactions (mandatory)
# ${10}: Gas multiplier for adjusting gas prices (mandatory)
# ${11}: Maximum number of messages per transaction (mandatory)
# ${12}: Maximum transaction size (mandatory)
# ${13}: Maximum allowable clock drift (in seconds) (mandatory)
# ${14}: Maximum allowable block time (in seconds) (mandatory)
# ${15}: Trusting period (in seconds) (mandatory)
# ${16}: Numerator for the trust threshold calculation (mandatory)
# ${17}: Denominator for the trust threshold calculation (mandatory)
# ${18}: Mnemonic seed for the Hermes account (mandatory)
# ${19}: Version of Hermes to install and configure (mandatory)

set -euox pipefail

SCRIPTS_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && cd .. && pwd)"
source "$SCRIPTS_DIR"/remote/lib/lib.sh
source "$SCRIPTS_DIR"/common/cmd.sh
source "$SCRIPTS_DIR"/common/rm-dir.sh
source "$SCRIPTS_DIR"/internal/add-dex-support.sh

declare -r hermes_root="$1"
declare -r chain1_id="$2"
declare -r chain1_ip_addr="$3"
declare -r chain1_rpc_port="$4"
declare -r chain1_grpc_port="$5"
declare -r chain1_rpc_timeout_secs="$6"
declare -r chain1_default_gas="$7"
declare -r chain1_max_gas="$8"
declare -r chain1_gas_price_price="$9"
declare -r chain1_gas_multiplier="${10}"
declare -r chain1_max_msg_num="${11}"
declare -r chain1_max_tx_size="${12}"
declare -r chain1_clock_drift_secs="${13}"
declare -r chain1_max_block_time_secs="${14}"
declare -r chain1_trusting_period_secs="${15}"
declare -r chain1_trust_threshold_numerator="${16}"
declare -r chain1_trust_threshold_denumerator="${17}"
declare -r hermes_mnemonic="${18}"
declare -r hermes_version="${19}"

# Install

declare -r archive_name="hermes-binary.tar.gz"
wget -O "$archive_name" https://github.com/informalsystems/hermes/releases/download/"$hermes_version"/hermes-"$hermes_version"-x86_64-unknown-linux-gnu.tar.gz

declare -r hermes_binary_dir="$hermes_root"/hermes
mkdir -p  "$hermes_binary_dir"
tar -C "$hermes_binary_dir" -vxzf "$archive_name"
rm "$archive_name"

# Configure

declare -r chain1_key_name="hermes-nolus"

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

# Add Nolus chain configuration
update_config "$hermes_config_dir" '.chains[0]."id"' '"'"$chain1_id"'"'
update_config "$hermes_config_dir" '.chains[0]."rpc_addr"' '"http://'"$chain1_ip_addr"':'"$chain1_rpc_port"'"'
update_config "$hermes_config_dir" '.chains[0]."grpc_addr"' '"http://'"$chain1_ip_addr"':'"$chain1_grpc_port"'"'
update_config "$hermes_config_dir" '.chains[0]."rpc_timeout"' '"'"$chain1_rpc_timeout_secs"'s"'
update_config "$hermes_config_dir" '.chains[0]."account_prefix"' '"nolus"'
update_config "$hermes_config_dir" '.chains[0]."key_name"' '"'"$chain1_key_name"'"'
update_config "$hermes_config_dir" '.chains[0]."address_type"' '{ derivation : "cosmos" }'
update_config "$hermes_config_dir" '.chains[0]."store_prefix"' '"ibc"'
update_config "$hermes_config_dir" '.chains[0]."default_gas"'  "$chain1_default_gas"
update_config "$hermes_config_dir" '.chains[0]."max_gas"' "$chain1_max_gas"
update_config "$hermes_config_dir" '.chains[0]."gas_price"' '{ price : '"$chain1_gas_price_price"', denom : "unls" }'
update_config "$hermes_config_dir" '.chains[0]."gas_multiplier"' "$chain1_gas_multiplier"
update_config "$hermes_config_dir" '.chains[0]."max_msg_num"' "$chain1_max_msg_num"
update_config "$hermes_config_dir" '.chains[0]."max_tx_size"' "$chain1_max_tx_size"
update_config "$hermes_config_dir" '.chains[0]."clock_drift"' '"'"$chain1_clock_drift_secs"'s"'
update_config "$hermes_config_dir" '.chains[0]."max_block_time"' '"'"$chain1_max_block_time_secs"'s"'
update_config "$hermes_config_dir" '.chains[0]."trusting_period"'  '"'"$chain1_trusting_period_secs"'s"'
update_config "$hermes_config_dir" '.chains[0]."trust_threshold"' '{ numerator : '"$chain1_trust_threshold_numerator"', denominator : '"$chain1_trust_threshold_denumerator"' }'
update_config "$hermes_config_dir" '.chains[0]."memo_prefix"' '"''"'
update_config "$hermes_config_dir" '.chains[0]."event_source"' '{ mode : "push", url : "ws://'"$chain1_ip_addr"':'"$chain1_rpc_port"'/websocket", batch_delay : "500ms" }'

# Account setup
declare hermes_mnemonic_file="$hermes_config_dir"/hermes.seed
echo "$hermes_mnemonic" > "$hermes_mnemonic_file"

dex_account_setup "$hermes_binary_dir" "$chain1_id" "$hermes_config_dir"/hermes.seed
