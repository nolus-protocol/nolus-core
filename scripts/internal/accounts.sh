#!/bin/bash
set -euxo pipefail

SCRIPT_DIR=$(cd $(dirname "${BASH_SOURCE[0]}") && pwd)
"$SCRIPT_DIR"/check-jq.sh

add_account() {
  local address="$1"
  local amount="$2"
  jq ". += [{ \"address\": \"$address\", \"amount\":  \"$amount\"}]"
}