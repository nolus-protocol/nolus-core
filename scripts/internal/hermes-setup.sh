#!/bin/bash

set -euox pipefail

INTERNAL_SCRIPTS_DIR=$(cd $(dirname "${BASH_SOURCE[0]}") && pwd)
source "$INTERNAL_SCRIPTS_DIR"/accounts.sh

# Open connection
open_connection() {
  declare -r hermes_binary_dir="$1"
  declare -r a_chain="$2"
  declare -r b_chain="$3"
  declare -r connection="$4"

  "$hermes_binary_dir"/hermes create connection --a-chain "$a_chain" --b-chain "$b_chain"
  "$hermes_binary_dir"/hermes create channel --a-chain "$a_chain" --a-connection "$connection" --a-port transfer --b-port transfer --order unordered
}