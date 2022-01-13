#!/bin/bash
set -euxo pipefail

# start "instance" variables
local_root_dir=""
local_chain_id=""
# end "instance" variables
LOCAL_BASE_RPC_PORT=26654
LOCAL_BASE_P2P_PORT=26653
LOCAL_BASE_PROXY_PORT=26655

init_local_sh() {
  local_root_dir="$1"
  local_chain_id="$2"

  rm -fr "$local_root_dir"
  mkdir -p "$local_root_dir"
}

deploy() {
  local node_id="$1"
  local node_index="$2"

  local node_dir
  node_dir=$(node_dir "$node_id")
  rm -fr "$node_dir"
  mkdir "$node_dir"

  run_cmd "$node_dir" init "$node_id" --chain-id "$local_chain_id" 1>/dev/null
  update_app "$node_dir" '."api"."enable"' "false"
  update_app "$node_dir" '."grpc"."enable"' "false"
  update_app "$node_dir" '."grpc-web"."enable"' "false"
  update_config "$node_dir" '."rpc"."laddr"' '"tcp://0.0.0.0:'"$(rpc_port "$node_index")"'"'
  update_config "$node_dir" '."p2p"."laddr"' '"tcp://0.0.0.0:'"$(p2p_port "$node_index")"'"'
  update_config "$node_dir" '."proxy_app"' '"tcp://127.0.0.1:'"$(proxy_port "$node_index")"'"'
}

gen_account() {
  local node_id="$1"
  local node_dir
  node_dir=$(node_dir "$node_id")

  run_cmd "$node_dir" keys add "$node_id" --keyring-backend test --output json 1>/dev/null
  run_cmd "$node_dir" keys show -a "$node_id" --keyring-backend test
}

# outputs the generated create validator transaction to the standard output
gen_validator() {
  local node_id="$1"
  local genesis_file="$2"
  local stake="$3"
  # local ip_address="$5"

  local node_dir
  node_dir=$(node_dir "$node_id")
  local tx_out_file="$node_dir/config/gentx_out.json"

  cp "$genesis_file" "$node_dir/config/genesis.json"
  # ip_spec=""
  # if [[ -n "${ip_address+}" ]]; then
  #   ip_spec="--ip $ip_address"
  # fi
  # $ip_spec
  run_cmd "$node_dir" gentx "$node_id" "$stake" --keyring-backend test --chain-id "$local_chain_id" --output-document "$tx_out_file" 1>/dev/null
  cat "$tx_out_file"
}

propagate_genesis() {
  local node_id="$1"
  local genesis_file="$2"

  cp "$genesis_file" "$(node_dir "$node_id")/config/genesis.json"
}

#####################
# private functions #
#####################
node_dir() {
  echo "$local_root_dir/$1"
}

rpc_port() {
  port $LOCAL_BASE_RPC_PORT "$@"
}

p2p_port() {
  port $LOCAL_BASE_P2P_PORT "$@"
}

proxy_port() {
  port $LOCAL_BASE_PROXY_PORT "$@"
}

port() {
  local base_port=$1
  local index=$2
  echo $((base_port + index*3))
}