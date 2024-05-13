#!/bin/bash

wait_nolus_gets_ready() {
  local -r nolus_home_dir="$1"

  while pidof -q "nolusd"
  do
    if local response="$(run_cmd "$nolus_home_dir" "status")"
    then
      if echo "${response}" | jq -e ".SyncInfo.latest_block_height | tonumber | . == 0" 2>"/dev/null"
      then
        echo "Block not incremented!"
      else
        return 0
      fi
    else
      echo "Failed to fetch node instance's status!"
    fi

    sleep 1
  done

  return 1
}

wait_hermes_config_gets_healthy() {
  local -r hermes_binary_dir="$1"
  local -r hermes_connection_warn="Reason: error in underlying transport when making gRPC call"
  local -r tmp_file="hermes_connection_check.txt"
  local hermes_connection_check="NOT_STARTED"

  while [ "$hermes_connection_check" == "NOT_STARTED"  ]
  do
    sleep 1
    "$hermes_binary_dir"/hermes health-check &>"$tmp_file"
    grep -q "$hermes_connection_warn" "$tmp_file" && hermes_connection_check="NOT_STARTED" || hermes_connection_check="STARTED"
  done

  rm -rf "$tmp_file"
}

wait_tx_included_in_block() {
  local -r nolus_home_dir="$1"
  local -r nolus_net="$2"
  local -r tx_hash="$3"

  local tx_check="NOT_INCLUDED"

  while [ "$tx_check" == "NOT_INCLUDED"  ]
  do
    sleep 1
    tx_check=$(run_cmd "$nolus_home_dir" q tx "$tx_hash" --node "$nolus_net") || tx_check="NOT_INCLUDED"
  done
}