#!/bin/bash
set -euxo pipefail

SCRIPT_DIR=$(cd $(dirname "${BASH_SOURCE[0]}") && pwd)

cleanup() {
  cleanup_init_network_sh
  exit
}
trap cleanup INT TERM EXIT

VAL_ROOT_DIR="networks/nolus"
VALIDATORS=1
VAL_ACCOUNTS_DIR="$VAL_ROOT_DIR/val-accounts"
POSITIONAL=()

NATIVE_CURRENCY="unolus"
VAL_TOKENS="1000000000""$NATIVE_CURRENCY"
VAL_STAKE="1000000""$NATIVE_CURRENCY"
CHAIN_ID="nolus-private"
SUSPEND_ADMIN=""


while [[ $# -gt 0 ]]; do
  key="$1"

  case $key in

  -h | --help)
    printf \
    "Usage: %s
    [--chain_id <string>]
    [--validators_dir <validators_root_dir>]
    [-v|--validators <number>]
    [--validator_accounts_dir <validator_accounts_dir>]
    [--currency <native_currency>]
    [--validator-tokens <tokens_for_val_genesis_accounts>]
    [--validator-stake <tokens_val_will_stake>]
    [-ips <ip_addrs>]
    [--suspend-admin <bech32address>]" "$0"
    exit 0
    ;;

   --chain-id)
    CHAIN_ID="$2"
    shift
    shift
    ;;

   --validators_dir)
    VAL_ROOT_DIR="$2"
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

   --validator_accounts_dir)
    VAL_ACCOUNTS_DIR="$2"
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

  --suspend-admin)
    SUSPEND_ADMIN="$2"
    shift
    shift
    ;;

  *) # unknown option
    POSITIONAL+=("$1") # save it in an array for later
    shift              # past argument
    ;;

  esac
done

if [[ -z "$SUSPEND_ADMIN" ]]; then
  echo >&2 "Suspend admin was not set"
  exit 1
fi

# TBD open a few sample private investor accounts
# TBD open admin accounts, e.g. a treasury and a suspender
#  and pass them to init_network
source "$SCRIPT_DIR"/internal/config-validator-dev.sh
init_config_validator_dev_sh "$SCRIPT_DIR" "$VAL_ROOT_DIR"

source "$SCRIPT_DIR"/internal/init-network.sh
init_network "$VAL_ACCOUNTS_DIR" "$VALIDATORS" "$CHAIN_ID" "$NATIVE_CURRENCY" "$SUSPEND_ADMIN" "$VAL_TOKENS" "$VAL_STAKE" "[]"
