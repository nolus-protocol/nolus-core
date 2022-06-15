#!/bin/bash

check_accounts_dependencies() {
  local script_dir
  script_dir=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
  "$script_dir"/check-jq.sh
}

check_accounts_dependencies

add_account() {
  local address="$1"
  local amount="$2"
  jq ". += [{ \"address\": \"$address\", \"amount\":  \"$amount\"}]"
}

recover_account() {
  local -r mnemonic="$1"
  local -r name="anonymous"
  local tmp_faucet_dir
  tmp_faucet_dir="$(mktemp -d)"
  run_cmd "$tmp_faucet_dir" keys add --recover "$name" --keyring-backend test <<< "$mnemonic" 1>/dev/null
  run_cmd "$tmp_faucet_dir" keys show "$name" -a --keyring-backend test
}

generate_account() {
  local -r name="$1"
  local -r home_dir="$2"

  echo 'y' | run_cmd "$home_dir" keys add "$name" --output json --keyring-backend test 1>/dev/null
  run_cmd "$home_dir" keys show -a "$name" --keyring-backend test
}