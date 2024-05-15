#!/bin/bash

# start "instance" variables
setup_validator_scripts_home_dir=""
setup_validator_binary_artifact=""
setup_validator_scripts_artifact=""
setup_validator_prev_node_id=""
setup_validator_moniker_base=""
setup_validator_server_user=""
setup_validator_server_ip=""
setup_validator_ssh_key=""

# end "instance" variables
SETUP_VALIDATOR_BASE_PORT=26606
SETUP_VALIDATOR_HOME_DIR="/opt/deploy/nolus"
SETUP_VALIDATOR_TIMEOUT_COMMIT="5s"
SETUP_VALIDATOR_BINARY_TARGET_DIR="/usr/bin"
SETUP_VALIDATOR_SCRIPTS_TARGET_DIR="/opt/deploy"

init_setup_validator() {
  setup_validator_scripts_home_dir="$1"
  setup_validator_binary_artifact="$2"
  setup_validator_scripts_artifact="$3"
  setup_validator_moniker_base="$4"
  setup_validator_server_user="$5"
  setup_validator_server_ip="$6"
  setup_validator_ssh_key="$7"
}

deploy_binary() {
  local target_dir=${1:-$SETUP_VALIDATOR_BINARY_TARGET_DIR}
  __upload_tar "$setup_validator_binary_artifact" "$target_dir"
  __untar "$setup_validator_binary_artifact" "$target_dir"
}

deploy_scripts() {
  __upload_tar "$setup_validator_scripts_artifact" "$SETUP_VALIDATOR_SCRIPTS_TARGET_DIR"
  __untar "$setup_validator_scripts_artifact" "$SETUP_VALIDATOR_SCRIPTS_TARGET_DIR"
}

# Setup validator nodes and collect their ids and validator public keys
#
# The node ids and validator public keys are printed on the standard output one at a line.
setup_validators() {
  set -euo pipefail
  local validators_nb="$1"
  local -r minimum_gas_price="$2"
  local -r query_gas_limit="$3"

  for i in $(seq "$validators_nb"); do
    config "$i" "$minimum_gas_price" "$query_gas_limit"
  done
}

setup_services() {
  local validators_nb="$1"

  for i in $(seq "$validators_nb"); do
    local node_moniker
    node_moniker=$(__node_moniker "$i")

    "$setup_validator_scripts_home_dir"/server/run-shell-script.sh \
      "/opt/deploy/scripts/remote/validator-init-service.sh \
      $SETUP_VALIDATOR_HOME_DIR $node_moniker" \
      $setup_validator_server_user \
      $setup_validator_server_ip \
      $setup_validator_ssh_key
  done
}

propagate_genesis() {
  local genesis_file_path="$1"
  local validators_nb="$2"

  for i in $(seq "$validators_nb"); do
    __upload_genesis "$i" "$genesis_file_path"
  done
}

start_validators() {
  local validators_nb="$1"

  __do_cmd_services "$validators_nb" "start"
}

stop_validators() {
  local validators_nb="$1"

  __do_cmd_services "$validators_nb" "stop"
}

#
# Return the node ids and validator public keys printed on the standard output delimited with a space.
#
config() {
  local node_index="$1"
  local -r minimum_gas_price="$2"
  local -r query_gas_limit="$3"

  local home_dir
  home_dir=$(__home_dir "$node_index")
  local node_moniker
  node_moniker=$(__node_moniker "$node_index")
  local node_base_port
  node_base_port=$(__node_base_port "$node_index")

  local node_id_val_pub_key
  node_id_val_pub_key=$("$setup_validator_scripts_home_dir"/server/run-shell-script.sh \
    "/opt/deploy/scripts/remote/validator-config.sh \
                              $home_dir $node_moniker $node_base_port \
                              $SETUP_VALIDATOR_TIMEOUT_COMMIT \
                              $minimum_gas_price $query_gas_limit \
                              $setup_validator_prev_node_id" \
    $setup_validator_server_user \
    $setup_validator_server_ip \
    $setup_validator_ssh_key)
  read -r setup_validator_prev_node_id __val_pub_key <<<"$node_id_val_pub_key"
  echo "$node_id_val_pub_key"
}

#####################
# private functions #
#####################
__home_dir() {
  local node_index=$1
  local node_id
  node_id=$(__node_moniker "$node_index")
  echo "$SETUP_VALIDATOR_HOME_DIR/$node_id"
}

__node_moniker() {
  echo "$setup_validator_moniker_base-$1"
}

__node_base_port() {
  local node_index=$1
  echo $((SETUP_VALIDATOR_BASE_PORT + node_index * 5))
}

__do_cmd_services() {
  local validators_nb="$1"
  local cmd="$2"

  for i in $(seq "$validators_nb"); do
    local node_moniker
    node_moniker=$(__node_moniker "$i")
    $setup_validator_scripts_home_dir/server/run-shell-script.sh \
      "systemctl $cmd $node_moniker" \
      "$setup_validator_server_user" \
      "$setup_validator_server_ip" \
      "$setup_validator_ssh_key"
  done
}

__upload_tar() {
  local archive_full_path="$1"
  local target_dir="$2"
  local archive_name="$(basename $archive_full_path)"

  "$setup_validator_scripts_home_dir"/server/run-shell-script.sh \
    "mkdir -p $target_dir" \
    $setup_validator_server_user \
    $setup_validator_server_ip \
    $setup_validator_ssh_key

  "$setup_validator_scripts_home_dir"/server/copy-file.sh \
    $archive_full_path \
    $target_dir/$archive_name \
    $setup_validator_server_user \
    $setup_validator_server_ip
}

__untar() {
  local archive_full_path="$1"
  local target_dir="$2"
  local archive_name="$(basename $archive_full_path)"

  "$setup_validator_scripts_home_dir"/server/run-shell-script.sh \
    "tar -xvf $target_dir/$archive_name -C $target_dir" \
    $setup_validator_server_user \
    $setup_validator_server_ip \
    $setup_validator_ssh_key
}

__upload_genesis() {
  local node_index="$1"
  local genesis_file_path="$2"

  local genesis_name
  genesis_name="$(basename "$genesis_file_path")"

  local home_dir
  home_dir=$(__home_dir "$node_index")

  "$setup_validator_scripts_home_dir"/server/copy-file.sh \
    $genesis_file_path \
    "$home_dir/config" \
    $setup_validator_server_user \
    $setup_validator_server_ip
}
