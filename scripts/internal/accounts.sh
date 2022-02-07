#!/bin/bash

check_accounts_dependencies() {
  local script_dir
  script_dir=$(cd $(dirname "${BASH_SOURCE[0]}") && pwd)
  "$script_dir"/check-jq.sh
}

check_accounts_dependencies

add_account() {
  local address="$1"
  local amount="$2"
  jq ". += [{ \"address\": \"$address\", \"amount\":  \"$amount\"}]"
}