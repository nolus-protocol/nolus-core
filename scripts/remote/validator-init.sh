#!/bin/bash
# Initialize a locally installed validator node for the needs of the test or main networks
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

declare -r home_dir="$1"
declare -r node_moniker="$2"

rm -fr "$home_dir"
mkdir -p "$home_dir"

run_cmd "$home_dir" init "$node_moniker" >/dev/null

tendermint_node_id=$(run_cmd "$home_dir" tendermint show-node-id)
validator_pub_key=$(run_cmd "$home_dir" tendermint show-validator)
echo "$tendermint_node_id $validator_pub_key"