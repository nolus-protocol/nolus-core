#!/bin/bash
set -euxo pipefail

cleanup() {
  if [[ -n "${GENESIS_HOME_DIR:-}" ]]; then
    rm -rf "$GENESIS_HOME_DIR"
  fi
  exit
}
trap cleanup INT TERM EXIT

SCRIPT_DIR=$(cd $(dirname "${BASH_SOURCE[0]}") && pwd)

VALIDATORS=1
IP_ADDRESSES=()
CUSTOM_IPS=false
POSITIONAL=()

MODE="local"
NATIVE_CURRENCY="unolus"
VAL_TOKENS="1000000000""$NATIVE_CURRENCY"
VAL_STAKE="1000000""$NATIVE_CURRENCY"
CHAIN_ID="nolus-private"
OUTPUT_DIR="dev-net"
SUSPEND_ADMIN=""
GENESIS_HOME_DIR=$(mktemp -d)

while [[ $# -gt 0 ]]; do
  key="$1"

  case $key in

  -h | --help)
    printf \
    "Usage: %s
    [--chain_id <string>]
    [-v|--validators <number>]
    [--currency <native_currency>]
    [--validator-tokens <tokens_for_val_genesis_accounts>]
    [--validator-stake <tokens_val_will_stake>]
    [-ips <ip_addrs>]
    [--suspend-admin <bech32address>]
    [-m|--mode <local|docker>]
    [-o|--output <output_dir>]" "$0"
    exit 0
    ;;

   --chain-id)
    CHAIN_ID="$2"
    shift
    shift
    ;;

   -v | --validators)
    VALIDATORS="$2"
    [ "$VALIDATORS" -gt 0 ] || {
      echo >&2 "validators must be a positive number"
      exit 1
    }
    shift
    shift
    ;;

  --currency)
    NATIVE_CURRENCY="$2"
    shift
    shift
    ;;

  --validator-tokens)
    VAL_TOKENS="$2"
    shift
    shift
    ;;

  --validator-stake)
    VAL_STAKE="$2"
    shift
    shift
    ;;

  -ips)
    for i in ${2//,/ }; do
      IP_ADDRESSES+=("$i")
    done
    CUSTOM_IPS=true
    shift
    shift
    ;;
  --suspend-admin)
    SUSPEND_ADMIN="$2"
    shift
    shift
    ;;
  -m | --mode)
    MODE="$2"
    [[ "$MODE" == "local" || "$MODE" == "docker" ]] || {
      echo >&2 "mode must be either local or docker"
      exit 1
    }
    shift
    shift
    ;;

  -o | --output)
    OUTPUT_DIR="$2"
    shift
    shift
    ;;

  *) # unknown option
    POSITIONAL+=("$1") # save it in an array for later
    shift              # past argument
    ;;

  esac
done

source "$SCRIPT_DIR"/internal/cmd.sh
source "$SCRIPT_DIR"/internal/config-validator-dev.sh
init_config_validator_dev_sh "$SCRIPT_DIR" "$OUTPUT_DIR"

source "$SCRIPT_DIR"/internal/local.sh
init_local_sh "$OUTPUT_DIR" "$CHAIN_ID"

source "$SCRIPT_DIR"/internal/accounts.sh
source "$SCRIPT_DIR"/internal/genesis.sh


# Init validator nodes, generate validator accounts and collect their addresses
#
# The nodes are placed in sub directories of $OUTPUT_DIR
# The validator addresses are printed on the standard output one at a line
init_nodes() {
  for i in $(seq "$VALIDATORS"); do
    config "$i"
    local address
    address=$(gen_account "$i")
    echo "$address"
  done
}

gen_accounts_spec() {
  local addresses="$1"
  local file="$2"

  local accounts="[]"
  for address in $addresses; do
    accounts=$(echo "$accounts" | add_account "$address" "$VAL_TOKENS")
  done
  echo "$accounts" > "$file"
}

init_validators() {
  local proto_genesis_file="$1"

  for i in $(seq "$VALIDATORS"); do
    local create_validator_tx
    create_validator_tx=$(gen_validator "$i" "$proto_genesis_file" "$VAL_STAKE")
    echo "$create_validator_tx"
  done
}

propagate_genesis_all() {
  local genesis_file="$1"

  for i in $(seq "$VALIDATORS"); do
    propagate_genesis "$i" "$genesis_file"
  done
}

## validate dependencies are installed
command -v jq >/dev/null 2>&1 || {
  echo >&2 "jq not installed. More info: https://stedolan.github.io/jq/download/"
  exit 1
}

if [[ "$CUSTOM_IPS" = true && "${#IP_ADDRESSES[@]}" -ne "$VALIDATORS" ]]; then
  echo >&2 "non matching ip addresses"
  exit 1
fi

if [[ -z "$SUSPEND_ADMIN" ]]; then
  echo >&2 "Suspend admin was not set"
  exit 1
fi

ACCOUNTS_FILE="$OUTPUT_DIR/accounts.json"
PROTO_GENESIS_FILE="$OUTPUT_DIR/penultimate-genesis.json"
FINAL_GENESIS_FILE="$OUTPUT_DIR/genesis.json"

addresses="$(init_nodes)"
gen_accounts_spec "$addresses" "$ACCOUNTS_FILE"
generate_proto_genesis "$GENESIS_HOME_DIR" "$CHAIN_ID" "$ACCOUNTS_FILE" "$NATIVE_CURRENCY" "$PROTO_GENESIS_FILE" "$SUSPEND_ADMIN"
create_validator_txs="$(init_validators "$PROTO_GENESIS_FILE")"
integrate_genesis_txs "$GENESIS_HOME_DIR" "$PROTO_GENESIS_FILE" "$create_validator_txs" "$FINAL_GENESIS_FILE"
propagate_genesis_all "$FINAL_GENESIS_FILE"
