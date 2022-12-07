#!/bin/bash

wait_nolus_gets_ready() {
  local -r nolus_home_dir="$1"
  local nolus_node_status=""
  local latest_block_height=0

  while [ "$latest_block_height" -le 0 ]
  do
    sleep 1
    nolus_node_status=$(run_cmd "$nolus_home_dir" status) && nolus_node_status="STARTED"

    if [ "$nolus_node_status" == "STARTED" ]
      then
          latest_block_height=$(run_cmd "$nolus_home_dir" status | jq .SyncInfo.latest_block_height | tr -d '"')
      fi
  done
}

wait_hermes_config_gets_healthy() {
  local -r hermes_binary_dir="$1"
  local -r hermes_connection_warn="Reason: error in underlying transport when making gRPC call"
  local -r tmp_file="hermes_connection_check.txt"
  local hermes_connection_check=""

  "$hermes_binary_dir"/hermes health-check &>"$tmp_file"
  grep -q "$hermes_connection_warn" "$tmp_file" && hermes_connection_check="NOT_STARTED"

  while [ "$hermes_connection_check" == "NOT_STARTED"  ]
  do
    sleep 1
    "$hermes_binary_dir"/hermes health-check &>"$tmp_file"
    grep -q "$hermes_connection_warn" "$tmp_file" && hermes_connection_check="NOT_STARTED" || hermes_connection_check="STARTED"
  done

  rm -rf "$tmp_file"
}