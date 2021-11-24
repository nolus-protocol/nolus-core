#!/bin/bash
set -euo pipefail

command -v create-vesting-account.sh >/dev/null 2>&1 || {
  echo >&2 "scripts are not found in $PATH."
  exit 1
}
source "create-vesting-account.sh"

cleanup() {
  if [[ -n "${TMPDIR:-}" ]]; then
    rm -rf "$TMPDIR"
  fi
  if [[ -n "${ORIG_DIR:-}" ]]; then
    cd "$ORIG_DIR"
  fi
  exit
}

trap cleanup INT TERM EXIT

CHAIN_ID="nolus-private"
OUTPUT_FILE="genesis.json"
MODE="local"
ACCOUNTS_FILE=""
TMPDIR=$(mktemp -d)
MONIKER="localtestnet"
KEYRING="test"
NATIVE_CURRENCY="nolus"

POSITIONAL=()
while [[ $# -gt 0 ]]; do
  key="$1"

  case $key in
  -c | --chain-id)
    CHAIN_ID="$2"
    shift # past argument
    shift # past value
    ;;
  -o | --output)
    OUTPUT_FILE="$2"
    shift
    shift
    ;;
  --accounts)
    ACCOUNTS_FILE=$(realpath "$2")
    shift
    shift
    ;;
  --currency)
    NATIVE_CURRENCY="$2"
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
  --help)
    echo "Usage: init-penultimate-genesis.sh [-c|--chain-id <chain_id>] [-o|--output <output_file>] [--accounts <accounts_file>] [--currency <native_currency>] [-m|--mode <local|docker>]"
    exit 0
    ;;
  *) # unknown option
    POSITIONAL+=("$1") # save it in an array for later
    shift              # past argument
    ;;
  esac
done

update_genesis() {
  jq $1 <"$TMPDIR/config/genesis.json" >"$TMPDIR/config/tmp_genesis.json" && mv "$TMPDIR/config/tmp_genesis.json" "$TMPDIR/config/genesis.json"
}

run_cmd() {
  local DIR="$1"
  shift
  case $MODE in
  local) cosmzoned $@ --home "$DIR" ;;
  docker) docker run --rm -u "$(id -u)":"$(id -u)" -v "$DIR:/tmp/.cosmzone:Z" nolus/node $@ --home /tmp/.cosmzone ;;
  esac
}

# validate dependencies are installed
command -v jq >/dev/null 2>&1 || {
  echo >&2 "jq not installed. More info: https://stedolan.github.io/jq/download/"
  exit 1
}

ORIG_DIR=$(pwd)
cd "$TMPDIR"
run_cmd "." init $MONIKER --chain-id "$CHAIN_ID"
run_cmd "." config keyring-backend "$KEYRING"
run_cmd "." config chain-id "$CHAIN_ID"

# Change parameter token denominations to NATIVE_CURRENCY
update_genesis '.app_state["staking"]["params"]["bond_denom"]="'"$NATIVE_CURRENCY"'"'
update_genesis '.app_state["crisis"]["constant_fee"]["denom"]="'"$NATIVE_CURRENCY"'"'
update_genesis '.app_state["gov"]["deposit_params"]["min_deposit"][0]["denom"]="'"$NATIVE_CURRENCY"'"'
update_genesis '.app_state["gov"]["deposit_params"]["min_deposit"][0]["denom"]="'"$NATIVE_CURRENCY"'"'
update_genesis '.app_state["mint"]["params"]["mint_denom"]="'"$NATIVE_CURRENCY"'"'

if [[ -n "${ACCOUNTS_FILE+x}" ]]; then
  for i in $(jq '. | keys | .[]' "$ACCOUNTS_FILE"); do
    row=$(jq ".[$i]" "$ACCOUNTS_FILE")
    address=$(jq -r '.address' <<<"$row")
    amount=$(jq -r '.amount' <<<"$row")
    if [[ "$(jq -r '.vesting' <<<"$row")" != 'null' ]]; then
      add_vesting_account "$row" "$TMPDIR"
    else
      run_cmd "." add-genesis-account "$address" "$amount"
    fi
  done
fi

cd "$ORIG_DIR"
cp "$TMPDIR/config/genesis.json" "$OUTPUT_FILE"
