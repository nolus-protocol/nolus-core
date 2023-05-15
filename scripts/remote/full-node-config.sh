#!/bin/bash
# Configure a locally installed full node for the needs of the test or main networks
#
# arg: home directory of the full node, mandatory
# arg: external IP address, mandatory
# arg: P2P port, mandatory
# arg: RPC port, mandatory
# arg: Monitoring port, mandatory
# arg: API port, mandatory

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR"/lib/lib.sh
source "$SCRIPT_DIR"/../common/cmd.sh

home_dir="$1"
node_moniker="$2"
base_port="$3"
timeout_commit="$4"
persistent_peers="$5"

EXTERNAL_ADDRESS="0.0.0.0"
HOST="127.0.0.1"
P2P_PORT=$((base_port))
RPC_PORT=$((base_port + 1))
MONITORING_PORT=$((base_port + 2))
API_PORT=$((base_port + 3))
GRPC_PORT=$((base_port + 4))

if [[ -n "${home_dir:-}" ]]; then
    rm -rf "$home_dir"
fi

mkdir -p "$home_dir"

run_cmd "$home_dir" init "$node_moniker" >/dev/null

declare -r config_dir="$home_dir"/config

# although the API endpoint is deprecated it is still required by Keplr
# the grpc endpoint is required by the market data feeder
update_app "$config_dir" '."api"."enable"' "true"
update_app "$config_dir" '."api"."address"' '"tcp://'"$EXTERNAL_ADDRESS:$API_PORT"'"'
update_app "$config_dir" '."api"."enabled-unsafe-cors"' "true"
update_app "$config_dir" '."grpc"."enable"' "true" >/dev/null
update_app "$config_dir" '."grpc"."address"' '"0.0.0.0:'"$GRPC_PORT"'"' >/dev/null
update_app "$config_dir" '."grpc-web"."enable"' "false"
update_app "$config_dir" '."telemetry"."enabled"' "true"
update_app "$config_dir" '."telemetry"."prometheus-retention-time"' "1"

update_config "$config_dir" '."rpc"."laddr"' '"tcp://'"$EXTERNAL_ADDRESS:$RPC_PORT"'"'
update_config "$config_dir" '."rpc"."cors_allowed_origins"' '["*"]'
update_config "$config_dir" '."p2p"."laddr"' '"tcp://'"$EXTERNAL_ADDRESS:$P2P_PORT"'"'
update_config "$config_dir" '."p2p"."seed_mode"' "false"
update_config "$config_dir" '."p2p"."pex"' "true"
update_config "$config_dir" '."p2p"."persistent_peers"' '"'"$persistent_peers"'"'
update_config "$config_dir" '."p2p"."private_peer_ids"' '""'
update_config "$config_dir" '."p2p"."addr_book_strict"' "false"
update_config "$config_dir" '."p2p"."allow_duplicate_ip"' "false"
update_config "$config_dir" '."instrumentation"."prometheus"' "true"
update_config "$config_dir" '."instrumentation"."prometheus_listen_addr"' '"'":$MONITORING_PORT"'"'
update_config "$config_dir" '."log_format"' '"json"'
update_config "$config_dir" '."log_level"' '"debug"'
