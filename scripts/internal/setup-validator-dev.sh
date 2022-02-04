#!/bin/bash
set -euxo pipefail

# start "instance" variables
setup_validator_dev_scripts_home_dir=""
setup_validator_dev_binary_artifact=""
setup_validator_dev_scripts_artifact=""
setup_validator_dev_root_dir=""
setup_validator_dev_prev_node_id=""
# end "instance" variables
SETUP_VALIDATOR_DEV_BASE_PORT=26606
SETUP_VALIDATOR_DEV_ARTIFACT_S3_BUCKET="nolus-artifact-bucket/dev"
SETUP_VALIDATOR_DEV_BIN_TAR="nolusd.tar"
SETUP_VALIDATOR_DEV_AWS_INSTANCE_ID="i-0307d4bb453d880f3"

init_setup_validator_dev_sh() {
  setup_validator_dev_scripts_home_dir="$1"
  setup_validator_dev_binary_artifact="$2"
  setup_validator_dev_scripts_artifact="$3"
  setup_validator_dev_root_dir="$4"
}

# Setup validator nodes and collect their ids and validator public keys
#
# The nodes are installed and configured depending on the sourced implementation script.
# The node ids and validator public keys are printed on the standard output one at a line.
setup_all() {
  set -euxo pipefail
  local validators_nb="$1"

  __deploy
  for i in $(seq "$validators_nb"); do
    config "$i"
  done
}

propagate_genesis_all() {
  local genesis_file="$1"
  local validators_nb="$2"

  for i in $(seq "$validators_nb"); do
    propagate_genesis "$i" "$genesis_file"
  done
}

#
# Return the node ids and validator public keys printed on the standard output delimited with a space.
#
config() {
  set -euxo pipefail
  local node_index="$1"

  local home_dir
  home_dir=$(__home_dir "$node_index")
  local node_moniker
  node_moniker=$(__node_moniker "$node_index")
  local node_base_port
  node_base_port=$(__node_base_port "$node_index")

  local node_id_val_pub_key
  node_id_val_pub_key=$("$setup_validator_dev_scripts_home_dir"/remote/validator-dev.sh "$home_dir" "$node_moniker" \
                                          "$node_base_port" "$setup_validator_dev_prev_node_id")
  read -r setup_validator_dev_prev_node_id __val_pub_key <<< "$node_id_val_pub_key"
  echo "$node_id_val_pub_key"
}

propagate_genesis() {
  local node_index="$1"
  local genesis_file="$2"

  cp "$genesis_file" "$(__home_dir "$node_index")/config/genesis.json"
}

#####################
# private functions #
#####################
__home_dir() {
  local node_index=$1
  local node_id
  node_id=$(__node_moniker "$node_index")
  echo "$setup_validator_dev_root_dir/$node_id"
}

__node_moniker() {
  echo "dev-validator-$1"
}

__node_base_port() {
  local node_index=$1
  echo $((SETUP_VALIDATOR_DEV_BASE_PORT + node_index*5))
}

__deploy() {
  __upload_to_s3
  __download_from_s3 "$setup_validator_dev_binary_artifact" "/usr/bin"
  __download_from_s3 "$setup_validator_dev_scripts_artifact" "/opt/deploy"
}

__upload_to_s3() {
  aws s3 cp "$setup_validator_dev_binary_artifact" s3://"$SETUP_VALIDATOR_DEV_ARTIFACT_S3_BUCKET"/ 1>&2 2>/dev/null
  aws s3 cp "$setup_validator_dev_scripts_artifact" s3://"$SETUP_VALIDATOR_DEV_ARTIFACT_S3_BUCKET"/ 1>&2 2>/dev/null
}

__download_from_s3() {
  local archive_full_path="$1"
  local target_dir="$2"
  archive_name="$(basename $archive_full_path)"
  "$setup_validator_dev_scripts_home_dir"/aws/run-shell-script.sh \
      "aws s3 cp s3://$SETUP_VALIDATOR_DEV_ARTIFACT_S3_BUCKET/$archive_name $target_dir && \
      tar -xvf $target_dir/$archive_name -C $target_dir" "$SETUP_VALIDATOR_DEV_AWS_INSTANCE_ID"
}
