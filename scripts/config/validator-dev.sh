#!/bin/bash
# Configure a locally installed validator node for the needs of the dev net
# The Nolus binary should be accessible on the system path.
# TBD what part of the scripts should be available next to this script
#
# arg1: home directory of the validator node, mandatory
# arg2: node's moniker, mandatory
# arg3: base port, mandatory. Used to determine the endpoint ports.
# arg4: first node's identificator, optional. Empty, if this is the first node.
#
# Returns the node identificator in the form of "node-id@host:p2p-port" followed
# by the node public key in JSON.
set -euxo pipefail

SCRIPT_DIR=$(cd $(dirname "${BASH_SOURCE[0]}") && pwd)
source "$SCRIPT_DIR"/lib/lib.sh
source "$SCRIPT_DIR"/../internal/cmd.sh

home_dir="$1"
node_moniker="$2"
base_port="$3"
first_node_id="$4"

HOST="127.0.0.1"
P2P_PORT=$((base_port))
RPC_PORT=$((base_port+1))
PROXY_PORT=$((base_port+2))

rm -fr "$home_dir"
mkdir -p "$home_dir"

run_cmd "$home_dir" init "$node_moniker" 1>/dev/null
update_app "$home_dir" '."api"."enable"' "false"
update_app "$home_dir" '."grpc"."enable"' "false"
update_app "$home_dir" '."grpc-web"."enable"' "false"

update_config "$home_dir" '."rpc"."laddr"' '"tcp://0.0.0.0:'"$RPC_PORT"'"'
update_config "$home_dir" '."p2p"."laddr"' '"tcp://'"$HOST:$P2P_PORT"'"'
update_config "$home_dir" '."p2p"."addr_book_strict"' 'false'
update_config "$home_dir" '."p2p"."allow_duplicate_ip"' 'true'
update_config "$home_dir" '."p2p"."persistent_peers"' '"'"$first_node_id"'"'
update_config "$home_dir" '."proxy_app"' '"tcp://'"$HOST:$PROXY_PORT"'"'

tendermint_node_id=$(run_cmd "$home_dir" tendermint show-node-id)
validator_pub_key=$(run_cmd "$home_dir" tendermint show-validator)
echo "$tendermint_node_id@$HOST:$P2P_PORT $validator_pub_key"