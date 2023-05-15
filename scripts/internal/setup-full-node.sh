#!/bin/bash

# start "instance" variables
setup_full_node_scripts_home_dir=""
setup_full_node_binary_artifact=""
setup_full_node_scripts_artifact=""
setup_full_node_moniker_base=""
setup_full_node_server_user=""
setup_full_node_server_ip=""
setup_full_node_persistent_peers=""

# end "instance" variables
SETUP_FULL_NODE_BASE_PORT=26606
SETUP_FULL_NODE_HOME_DIR="/opt/deploy/nolus"
SETUP_FULL_NODE_TIMEOUT_COMMIT="5s"

init_setup_full_node() {
  setup_full_node_scripts_home_dir="$1"
  setup_full_node_binary_artifact="$2"
  setup_full_node_scripts_artifact="$3"
  setup_full_node_moniker_base="$4"
  setup_full_node_persistent_peers="$5"
  setup_full_node_server_user="$6"
  setup_full_node_server_ip="$7"
}

deploy_binary() {
  __upload_tar "$setup_full_node_binary_artifact" "/usr/bin"
  __untar "$setup_full_node_binary_artifact" "/usr/bin"
}

deploy_scripts() {
  __upload_tar "$setup_full_node_scripts_artifact" "/opt/deploy"
  __untar "$setup_full_node_scripts_artifact" "/opt/deploy"
}

setup_services() {
  local COUNT="$1"

  for i in $(seq "$COUNT"); do
    local node_moniker
    node_moniker=$(__full_node_moniker "$i")

    "$setup_full_node_scripts_home_dir"/server/run-shell-script.sh \
      "/opt/deploy/scripts/remote/validator-init-service.sh \
      $SETUP_FULL_NODE_HOME_DIR $node_moniker" \
      $setup_full_node_server_user \
      $setup_full_node_server_ip
  done
}

setup_full_node() {
  set -euo pipefail
  local COUNT="$1"

  for i in $(seq "$COUNT"); do
    __config "$i"
  done
}

#####################
# private functions #
#####################
__config() {
  local node_index="$1"

  local home_dir
  home_dir=$(__home_dir "$node_index")
  local node_moniker
  node_moniker=$(__full_node_moniker "$node_index")
  local node_base_port
  node_base_port=$(__node_base_port "$node_index")

  local -r scripts_remote_dir="/opt/deploy/scripts/remote"
  local -r config_full_node="$scripts_remote_dir/full-node-config.sh"
  "$setup_full_node_scripts_home_dir"/server/run-shell-script.sh \
    "$config_full_node \
            $home_dir \
            $node_moniker \
            $node_base_port \
            $SETUP_FULL_NODE_TIMEOUT_COMMIT
            $setup_full_node_persistent_peers" \
    $setup_full_node_server_user \
    $setup_full_node_server_ip
}

__home_dir() {
  local node_index=$1
  local node_id
  node_id=$(__full_node_moniker "$node_index")
  echo "$SETUP_FULL_NODE_HOME_DIR/$node_id"
}

__full_node_moniker() {
  echo "$setup_full_node_moniker_base-$1"
}

__node_base_port() {
  local node_index=$1
  echo $((SETUP_FULL_NODE_BASE_PORT + node_index * 5))
}

__upload_tar() {
  local archive_full_path="$1"
  local target_dir="$2"
  local archive_name="$(basename $archive_full_path)"

  "$setup_full_node_scripts_home_dir"/server/run-shell-script.sh \
    "mkdir -p $target_dir" \
    $setup_full_node_server_user \
    $setup_full_node_server_ip

  "$setup_full_node_scripts_home_dir"/server/copy-file.sh \
    $archive_full_path \
    $target_dir/$archive_name \
    $setup_full_node_server_user \
    $setup_full_node_server_ip
}

__untar() {
  local archive_full_path="$1"
  local target_dir="$2"
  local archive_name="$(basename $archive_full_path)"

  "$setup_full_node_scripts_home_dir"/server/run-shell-script.sh \
    "tar -xvf $target_dir/$archive_name -C $target_dir" \
    $setup_full_node_server_user \
    $setup_full_node_server_ip
}
