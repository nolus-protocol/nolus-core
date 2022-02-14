#!/bin/bash
# Configure a locally installed validator node for the needs of the dev net
# The Nolus binary should be accessible on the system path.
# TBD what part of the scripts should be available next to this script
#
# arg1: home directory of the validator node, mandatory
# arg2: node's moniker, mandatory
# arg3: base port, mandatory. Used to determine the endpoint ports.
# arg4: tls enabled, mandatory. Pass "true" to configure TLS.
# arg5: first node's identificator, optional. Empty, if this is the first node.
#
# Returns the node identificator in the form of "node-id@host:p2p-port" followed
# by the node public key in JSON.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR"/lib/lib.sh
source "$SCRIPT_DIR"/../common/cmd.sh

home_dir="$1"
node_moniker="$2"
base_port="$3"
tls_enable="$4"
if [[ $# -gt 4 ]]
then first_node_id="$5"
else first_node_id=""
fi

exit_if_not_present() {
    local file_name="$1"
    if [[ ! -f "$file_name" ]]; then
        echo "$file_name not found!"
        exit 1
    fi
}

HOST="127.0.0.1"
P2P_PORT=$((base_port))
RPC_PORT=$((base_port+1))
API_PORT=$((base_port+3))
TLS_CERT_FILE="/etc/pki/tls/certs/wildcard.nolus.io.pem"
TLS_KEY_FILE="/etc/pki/tls/private/wildcard.nolus.io.key"

rm -fr "$home_dir"
mkdir -p "$home_dir"

run_cmd "$home_dir" init "$node_moniker" >/dev/null
# although the API endpoint is deprecated it is still required by Keplr
# TBD reevaluate the necessity to remain open
update_app "$home_dir" '."api"."enable"' "true" >/dev/null
update_app "$home_dir" '."api"."address"' '"tcp://0.0.0.0:'"$API_PORT"'"' >/dev/null
update_app "$home_dir" '."api"."enabled-unsafe-cors"' "true" >/dev/null
update_app "$home_dir" '."grpc"."enable"' "false" >/dev/null
update_app "$home_dir" '."grpc-web"."enable"' "false" >/dev/null

update_config "$home_dir" '."rpc"."laddr"' '"tcp://0.0.0.0:'"$RPC_PORT"'"' >/dev/null
update_config "$home_dir" '."rpc"."cors_allowed_origins"' '["*"]' >/dev/null
if [[ "$tls_enable" == "true" ]]; then
    exit_if_not_present "$TLS_CERT_FILE"
    exit_if_not_present "$TLS_KEY_FILE"
    update_config "$home_dir" '."rpc"."tls_cert_file"' '"'"$TLS_CERT_FILE"'"' >/dev/null
    update_config "$home_dir" '."rpc"."tls_key_file"' '"'"$TLS_KEY_FILE"'"' >/dev/null
fi
update_config "$home_dir" '."p2p"."laddr"' '"tcp://'"$HOST:$P2P_PORT"'"' >/dev/null
update_config "$home_dir" '."p2p"."addr_book_strict"' 'false' >/dev/null
update_config "$home_dir" '."p2p"."allow_duplicate_ip"' 'true' >/dev/null
update_config "$home_dir" '."p2p"."persistent_peers"' '"'"$first_node_id"'"' >/dev/null

update_config "$home_dir" '."proxy_app"' '""' >/dev/null

tendermint_node_id=$(run_cmd "$home_dir" tendermint show-node-id)
validator_pub_key=$(run_cmd "$home_dir" tendermint show-validator)
echo "$tendermint_node_id@$HOST:$P2P_PORT $validator_pub_key"