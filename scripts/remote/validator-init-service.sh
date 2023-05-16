#!/bin/bash
# Initialize a locally installed validator node.
# The Nolus binary should be accessible on the system path.
#
# arg1: home directory of the validator node, mandatory
# arg2: node's moniker, mandatory
#
# Returns the node identifier as provided by Tendermint followed
# by the node public key in JSON.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR"/../common/cmd.sh
source "$SCRIPT_DIR"/../common/rm-dir.sh

declare -r home_dir="$1"
declare -r node_moniker="$2"

init_systemd_service() {
    cp --force "$SCRIPT_DIR/nolusd.service" "/etc/systemd/system/$node_moniker.service"
    sed -i "s/nolus$/&\/$node_moniker/" "/etc/systemd/system/$node_moniker.service"
    systemctl daemon-reload >/dev/null
    systemctl enable $node_moniker >/dev/null
}

rm_dir "$home_dir"
mkdir -p "$home_dir"

run_cmd "$home_dir" init "$node_moniker" >/dev/null

init_systemd_service

tendermint_node_id=$(run_cmd "$home_dir" tendermint show-node-id)
validator_pub_key=$(run_cmd "$home_dir" tendermint show-validator)
echo "$tendermint_node_id $validator_pub_key"
