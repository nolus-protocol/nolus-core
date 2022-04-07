#!/bin/bash

SETUP_VALIDATOR_P2P_PORT=26656
SETUP_VALIDATOR_RPC_PORT=26657
SETUP_VALIDATOR_MONITORING_PORT=26660
SETUP_VALIDATOR_API_PORT=1317
SETUP_VALIDATOR_TIMEOUT_COMMIT="5s"
SETUP_VALIDATOR_HOME_DIR="/opt/deploy/nolus"

stop_nodes() {
  local -r scripts_home_dir="$1"
  local -r validator_aws_instance_id="$2"
  local -r -a sentry_aws_instance_ids="$3"

  __do_cmd_services "stop" "$scripts_home_dir" "$validator_aws_instance_id" \
                      "$sentry_aws_instance_ids"
} 

deploy_nodes() {
  local -r scripts_home_dir="$1"
  local -r binary_artifact_path="$2"
  local -r scripts_artifact_path="$3"
  local -r deploy_medium_s3_bucket="$4"
  local -r validator_aws_instance_id="$5"
  local -r -a sentry_aws_instance_ids="$6"

  local -r and_untar="true"
  __transfer_file "$scripts_home_dir" "$binary_artifact_path" "/usr/bin/" \
                    "$deploy_medium_s3_bucket" "$validator_aws_instance_id" \
                    "$sentry_aws_instance_ids" "$and_untar"
  __transfer_file "$scripts_home_dir" "$scripts_artifact_path" "/opt/deploy/" \
                    "$deploy_medium_s3_bucket" "$validator_aws_instance_id" \
                    "$sentry_aws_instance_ids" "$and_untar"

  __ensure_tomlq_nodes "$scripts_home_dir" "$validator_aws_instance_id" "$sentry_aws_instance_ids"
}

# Setup а validator node and adjacent sentry nodes.
#
# Due to limitations in the key values of bash associative arrays we use two distinct indexed arrays
# with corresponding elements at the same indexes.
#
# The node urls, node_url="node_id@ip:port", and validator public keys are printed on the standard output one at a line.
# The first line contains validator's info. Each of the next lines contains sentry's info.
setup_nodes() {
  local -r scripts_home_dir="$1"
  local -r moniker_base="$2"
  local -r validator_aws_instance_id="$3"
  local -r validator_ip="$4"
  local -r -a sentry_aws_instance_ids="$5[@]"
  local -r -a sentry_aws_instance_ids_arr=("${!sentry_aws_instance_ids}")
  local -r -a sentry_aws_public_ips="$6[@]"
  local -r -a sentry_aws_public_ips_arr=("${!sentry_aws_public_ips}")
  local -r -a sentry_aws_private_ips="$7[@]"
  local -r -a sentry_aws_private_ips_arr=("${!sentry_aws_private_ips}")
  local -r others_sentry_node_urls_str="$8"

  #making sure all arrays are equal in length
  [[ ${#sentry_aws_instance_ids_arr[@]} -eq ${#sentry_aws_public_ips_arr[@]} ]]
  [[ ${#sentry_aws_instance_ids_arr[@]} -eq ${#sentry_aws_private_ips_arr[@]} ]]


  local validator_node_moniker
  validator_node_moniker=$(__validator_node_moniker "$moniker_base")

  local validator_node_id_pub_key
  validator_node_id_pub_key=$("$scripts_home_dir"/aws/run-shell-script.sh \
                          "export HOME=/home/ssm-user && /opt/deploy/scripts/remote/validator-init.sh \
                                  $SETUP_VALIDATOR_HOME_DIR $validator_node_moniker" \
                                  "$validator_aws_instance_id")
  local validator_node_id validator_pub_key validator_node_url
  read -r validator_node_id validator_pub_key <<< "$validator_node_id_pub_key"
  validator_node_url=$(__node_id_to_url "$validator_node_id" "$validator_ip" "$SETUP_VALIDATOR_P2P_PORT")
  local -r validator_node_url_pub_key="$validator_node_url $validator_pub_key"

  local sentry_node_public_url
  local sentry_node_private_url
  local -a sentry_node_ids
  local -a sentry_node_public_urls
  local -a sentry_node_private_urls
  local -a sentry_node_public_url_pub_keys
  for i in "${!sentry_aws_instance_ids_arr[@]}"; do
    local sentry_aws_instance_id="${sentry_aws_instance_ids_arr[$i]}"
    local sentry_node_moniker
    sentry_node_moniker=$(__sentry_node_moniker "$moniker_base" "$sentry_aws_instance_id")
  
    local sentry_node_id_pub_key
    sentry_node_id_pub_key=$("$scripts_home_dir"/aws/run-shell-script.sh \
                            "export HOME=/home/ssm-user && /opt/deploy/scripts/remote/validator-init.sh \
                                    $SETUP_VALIDATOR_HOME_DIR $sentry_node_moniker" \
                                    "$sentry_aws_instance_id")
    local sentry_node_id sentry_pub_key
    read -r sentry_node_id sentry_pub_key <<< "$sentry_node_id_pub_key"
    sentry_node_ids+=("$sentry_node_id")
    sentry_node_public_url=$(__node_id_to_url "$sentry_node_id" "${sentry_aws_public_ips_arr[$i]}" "$SETUP_VALIDATOR_P2P_PORT")
    sentry_node_public_urls+=("$sentry_node_public_url")
    sentry_node_private_url=$(__node_id_to_url "$sentry_node_id" "${sentry_aws_private_ips_arr[$i]}" "$SETUP_VALIDATOR_P2P_PORT")
    sentry_node_private_urls+=("$sentry_node_private_url")
    sentry_node_public_url_pub_keys+=("$sentry_node_public_url $sentry_pub_key")
  done

  local -r sentry_node_private_urls_str=$(__comma_join "${sentry_node_private_urls[@]}")
  local -r sentry_node_ids_str=$(__comma_join "${sentry_node_ids[@]}")
  "$scripts_home_dir"/aws/run-shell-script.sh \
      "/opt/deploy/scripts/remote/validator-config.sh \
            $SETUP_VALIDATOR_HOME_DIR $validator_ip $SETUP_VALIDATOR_P2P_PORT \
            $SETUP_VALIDATOR_RPC_PORT $SETUP_VALIDATOR_MONITORING_PORT \
            $SETUP_VALIDATOR_TIMEOUT_COMMIT $sentry_node_private_urls_str $sentry_node_ids_str" \
            "$validator_aws_instance_id"

  for sentry_aws_index in "${!sentry_aws_instance_ids_arr[@]}"; do
    "$scripts_home_dir"/aws/run-shell-script.sh \
        "/opt/deploy/scripts/remote/sentry-config.sh \
              $SETUP_VALIDATOR_HOME_DIR '0.0.0.0' $SETUP_VALIDATOR_P2P_PORT \
              $SETUP_VALIDATOR_RPC_PORT $SETUP_VALIDATOR_MONITORING_PORT $SETUP_VALIDATOR_API_PORT \
              $validator_node_url $validator_node_id $sentry_node_private_urls_str $sentry_node_ids_str \
              $others_sentry_node_urls_str" \
              "${sentry_aws_instance_ids_arr[$sentry_aws_index]}"
  done

  # dump the result out
  echo "$validator_node_url_pub_key"
  for sentry_node_public_url_pub_key in "${sentry_node_public_url_pub_keys[@]}"; do
    echo "$sentry_node_public_url_pub_key"
  done
}

propagate_genesis() {
  local -r scripts_home_dir="$1"
  local -r genesis_file_src_path="$2"
  local -r deploy_medium_s3_bucket="$3"
  local -r validator_aws_instance_id="$4"
  local -r -a sentry_aws_instance_ids="$5"

  local -r genesis_file_dest_dir="$SETUP_VALIDATOR_HOME_DIR/config/"
  local -r and_untar="false"
  __transfer_file "$scripts_home_dir" "$genesis_file_src_path" "$genesis_file_dest_dir" \
                    "$deploy_medium_s3_bucket" "$validator_aws_instance_id" \
                    "$sentry_aws_instance_ids" "$and_untar"
}

start_nodes() {
  local -r scripts_home_dir="$1"
  local -r validator_aws_instance_id="$2"
  local -r -a sentry_aws_instance_ids="$3"

  __do_cmd_services "start" "$scripts_home_dir" "$validator_aws_instance_id" \
                      "$sentry_aws_instance_ids"
} 

#####################
# private functions #
#####################
__validator_node_moniker() {
  echo "$1-validator"
}

__sentry_node_moniker() {
  local -r moniker_base="$1"
  local -r sentry_id="$2"
  echo "$moniker_base-sentry-$sentry_id"
}

__comma_join() {
  local IFS=","
  echo "$*"
}

__node_id_to_url() {
  echo "$1@$2:$3"
}

__do_cmd_service() {
  local -r cmd="$1"
  local -r scripts_home_dir="$2"
  local -r aws_instance_id="$3"

  "$scripts_home_dir"/aws/run-shell-script.sh \
      "systemctl $cmd nolusd.service" "$aws_instance_id"
}

__do_cmd_services() {
  local -r cmd="$1"
  local -r scripts_home_dir="$2"
  local -r aws_instance_id="$3"
  local -r -a sentry_aws_instance_ids="$4[@]"

  __do_cmd_service "$cmd" "$scripts_home_dir" "$aws_instance_id" 
  for sentry_aws_instance_id in "${!sentry_aws_instance_ids}"; do
    __do_cmd_service "$cmd" "$scripts_home_dir" "$sentry_aws_instance_id"
  done
}

__upload_to_s3() {
  aws s3 cp "$1" s3://"$2"/ >/dev/null
}

__download_from_s3() {
  local -r scripts_home_dir="$1"
  local -r file_full_path="$2"
  local -r target_dir="$3"
  local -r deploy_medium_s3_bucket="$4"
  local -r aws_instance_id="$5"
  local -r and_untar="$6"

  local file_name
  file_name="$(basename "$file_full_path")"
  local cmd="aws s3 cp s3://$deploy_medium_s3_bucket/$file_name $target_dir"
  if [[ "$and_untar" == "true" ]]; then
    cmd="$cmd && tar -xvf $target_dir/$file_name -C $target_dir"
  fi

  "$scripts_home_dir"/aws/run-shell-script.sh "$cmd" "$aws_instance_id"
}

__transfer_file() {
  local -r scripts_home_dir="$1"
  local -r file_src_path="$2"
  local -r file_dest_dir="$3"
  local -r deploy_medium_s3_bucket="$4"
  local -r validator_aws_instance_id="$5"
  local -r -a sentry_aws_instance_ids="$6[@]"
  local -r and_untar="$7"

  __upload_to_s3 "$file_src_path" "$deploy_medium_s3_bucket"
  __download_from_s3 "$scripts_home_dir" "$file_src_path" "$file_dest_dir" \
                      "$deploy_medium_s3_bucket" "$validator_aws_instance_id" \
                      "$and_untar"
  for sentry_aws_instance_id in "${!sentry_aws_instance_ids}"; do
    __download_from_s3 "$scripts_home_dir" "$file_src_path" "$file_dest_dir" \
                      "$deploy_medium_s3_bucket" "$sentry_aws_instance_id" \
                      "$and_untar"
  done
}

__ensure_tomlq() {
  local -r scripts_home_dir="$1"
  local -r aws_instance_id="$2"

  # tomlq requires jq
  "$scripts_home_dir"/aws/run-shell-script.sh \
      "sudo yum -y install jq" "$aws_instance_id"
  "$scripts_home_dir"/aws/run-shell-script.sh \
      "python3 -m ensurepip --upgrade --user && \
      pip3 install tomlq --user" "$aws_instance_id"
}

__ensure_tomlq_nodes() {
  local -r scripts_home_dir="$1"
  local -r validator_aws_instance_id="$2"
  local -r -a sentry_aws_instance_ids="$3[@]"

  __ensure_tomlq "$scripts_home_dir" "$validator_aws_instance_id"
  for sentry_aws_instance_id in "${!sentry_aws_instance_ids}"; do
    __ensure_tomlq "$scripts_home_dir" "$sentry_aws_instance_id"
  done
}