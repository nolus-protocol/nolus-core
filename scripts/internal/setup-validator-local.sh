#!/bin/bash

# start "instance" variables
setup_validator_local_scripts_home_dir=""
setup_validator_local_root_dir=""
setup_validator_local_prev_node_id=""
# end "instance" variables
SETUP_VALIDATOR_LOCAL_BASE_PORT=26606
SETUP_VALIDATOR_LOCAL_TIMEOUT_COMMIT="1s"

init_setup_validator_local_sh() {
  setup_validator_local_scripts_home_dir="$1"
  setup_validator_local_root_dir="$2"
}

# Setup validator nodes and collect their ids and validator public keys
#
# The node ids and validator public keys are printed on the standard output one at a line.
setup_validators() {
  set -euo pipefail
  local validators_nb="$1"

  for i in $(seq "$validators_nb"); do
    __config "$i"
  done
}

propagate_genesis() {
  local genesis_file_path="$1"
  local validators_nb="$2"

  for i in $(seq "$validators_nb"); do
    cp "$genesis_file_path" "$(__home_dir "$i")/config/genesis.json"
  done
}

first_node_rpc_port() {
  local base_port
  base_port=$(__node_base_port 1)
  echo $((base_port+1))
}

__home_dir() {
  local node_index=$1
  local node_id
  node_id=$(__node_moniker "$node_index")
  echo "$setup_validator_local_root_dir/$node_id"
}

__node_moniker() {
  echo "local-validator-$1"
}

__node_base_port() {
  local node_index=$1
  echo $((SETUP_VALIDATOR_LOCAL_BASE_PORT + node_index*5))
}

__config() {
  local node_index="$1"

  local home_dir
  home_dir=$(__home_dir "$node_index")
  local node_moniker
  node_moniker=$(__node_moniker "$node_index")
  local node_base_port
  node_base_port=$(__node_base_port "$node_index")

  local node_id_val_pub_key
  node_id_val_pub_key=$("$setup_validator_local_scripts_home_dir"/remote/validator-dev.sh "$home_dir" "$node_moniker" \
                                          "$node_base_port" "$SETUP_VALIDATOR_LOCAL_TIMEOUT_COMMIT" \
                                          "$setup_validator_local_prev_node_id")
  read -r setup_validator_local_prev_node_id __val_pub_key <<< "$node_id_val_pub_key"
  echo "$node_id_val_pub_key"
}

