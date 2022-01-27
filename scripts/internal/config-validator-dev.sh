#!/bin/bash
set -euxo pipefail

# start "instance" variables
config_validator_dev_scripts_home_dir=""
config_validator_dev_root_dir=""
config_validator_dev_prev_node_id=""
# end "instance" variables
CONFIG_VALIDATOR_DEV_BASE_PORT=26606

init_config_validator_dev_sh() {
  config_validator_dev_scripts_home_dir="$1"
  config_validator_dev_root_dir="$2"
}

#
# Return the node ids and validator public keys printed on the standard output delimited with a space.
#
config() {
  set -euxo pipefail
  local node_index="$1"

  local home_dir
  home_dir=$(home_dir "$node_index")
  local node_moniker
  node_moniker=$(node_moniker "$node_index")
  local node_base_port
  node_base_port=$(node_base_port "$node_index")

  local node_id_val_pub_key
  node_id_val_pub_key=$("$config_validator_dev_scripts_home_dir"/config/validator-dev.sh "$home_dir" "$node_moniker" \
                                          "$node_base_port" "$config_validator_dev_prev_node_id")
  read -r config_validator_dev_prev_node_id __val_pub_key <<< "$node_id_val_pub_key"
  echo "$node_id_val_pub_key"
}

propagate_genesis() {
  local node_index="$1"
  local genesis_file="$2"

  cp "$genesis_file" "$(home_dir "$node_index")/config/genesis.json"
}

#####################
# private functions #
#####################
home_dir() {
  local node_index=$1
  local node_id
  node_id=$(node_moniker "$node_index")
  echo "$config_validator_dev_root_dir/$node_id"
}

node_moniker() {
  echo "dev-validator-$1"
}

node_base_port() {
  local node_index=$1
  echo $((CONFIG_VALIDATOR_DEV_BASE_PORT + node_index*5))
}
