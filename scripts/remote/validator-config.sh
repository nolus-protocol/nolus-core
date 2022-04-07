#!/bin/bash
# Configure a locally installed validator node for the needs of the test or main networks
#
# arg: home directory of the validator node, mandatory
# arg: external IP address, mandatory
# arg: P2P port, mandatory
# arg: RPC port, mandatory
# arg: Monitoring port, mandatory
# arg: timeout commit, mandatory. Example: "3s".
# arg: a comma separated list of sentry node URLs, mandatory
# arg: a comma separated list of sentry node IDs, mandatory

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR"/lib/lib.sh

declare -r home_dir="$1"
declare -r ip_address="$2"
declare -r p2p_port="$3"
declare -r rpc_port="$4"
declare -r monitoring_port="$5"
declare -r timeout_commit="$6"
declare -r sentry_node_urls_str="$7"
declare -r sentry_node_ids_str="$8"

update_app "$home_dir" '."api"."enable"' "false"
update_app "$home_dir" '."grpc"."enable"' "false"
update_app "$home_dir" '."grpc-web"."enable"' "false"
update_app "$home_dir" '."minimum-gas-prices"' '"'"0.0025unolus"'"'
update_app "$home_dir" '."telemetry"."enabled"' "true"
update_app "$home_dir" '."telemetry"."prometheus-retention-time"' "1"

update_config "$home_dir" '."rpc"."laddr"' '"tcp://'"$ip_address:$rpc_port"'"'
update_config "$home_dir" '."p2p"."laddr"' '"tcp://'"$ip_address:$p2p_port"'"'
update_config "$home_dir" '."p2p"."seed_mode"' "false"
update_config "$home_dir" '."p2p"."pex"' "false"
update_config "$home_dir" '."p2p"."persistent_peers"' '"'"$sentry_node_urls_str"'"'
update_config "$home_dir" '."p2p"."unconditional_peer_ids"' '"'"$sentry_node_ids_str"'"'
update_config "$home_dir" '."p2p"."addr_book_strict"' "false"
update_config "$home_dir" '."p2p"."allow_duplicate_ip"' "false"
update_config "$home_dir" '."consensus"."double_sign_check_height"' "10"
update_config "$home_dir" '."proxy_app"' '""'
update_config "$home_dir" '."consensus"."timeout_commit"' '"'"$timeout_commit"'"'
update_config "$home_dir" '."instrumentation"."prometheus"' "true"
update_config "$home_dir" '."instrumentation"."prometheus_listen_addr"' '"'":$monitoring_port"'"'
