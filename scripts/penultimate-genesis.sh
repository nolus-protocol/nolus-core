#!/bin/bash
set -euo pipefail

SCRIPT_DIR=$(cd $(dirname "${BASH_SOURCE[0]}") && pwd)

source "$SCRIPT_DIR"/common-util.sh
source "$SCRIPT_DIR"/create-vesting-account.sh

cleanup() {
  if [[ -n "${TMPDIR:-}" ]]; then
    rm -rf "$TMPDIR"
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
NATIVE_CURRENCY="unolus"

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
    echo "Usage: penultimate-genesis.sh [-c|--chain-id <chain_id>] [-o|--output <output_file>] [--accounts <accounts_file>] [--currency <native_currency>] [-m|--mode <local|docker>]"
    exit 0
    ;;
  *) # unknown option
    POSITIONAL+=("$1") # save it in an array for later
    shift              # past argument
    ;;
  esac
done

# validate dependencies are installed
command -v jq >/dev/null 2>&1 || {
  echo >&2 "jq not installed. More info: https://stedolan.github.io/jq/download/"
  exit 1
}

GENESIS_FILE="$TMPDIR/config/genesis.json"
GENESIS_TMP_FILE="$TMPDIR/config/genesis-tmp.json"

run_cmd "$MODE" "$TMPDIR" init $MONIKER --chain-id "$CHAIN_ID"
run_cmd "$MODE" "$TMPDIR" config keyring-backend "$KEYRING"
run_cmd "$MODE" "$TMPDIR" config chain-id "$CHAIN_ID"

# Change parameter token denominations to NATIVE_CURRENCY
cat "$GENESIS_FILE" \
  | jq '.app_state["staking"]["params"]["bond_denom"]="'"$NATIVE_CURRENCY"'"' \
  | jq '.app_state["crisis"]["constant_fee"]["denom"]="'"$NATIVE_CURRENCY"'"' \
  | jq '.app_state["gov"]["deposit_params"]["min_deposit"][0]["denom"]="'"$NATIVE_CURRENCY"'"' \
  | jq '.app_state["gov"]["deposit_params"]["min_deposit"][0]["denom"]="'"$NATIVE_CURRENCY"'"' \
  | jq '.app_state["mint"]["params"]["mint_denom"]="'"$NATIVE_CURRENCY"'"' > "$GENESIS_TMP_FILE"
mv "$GENESIS_TMP_FILE" "$GENESIS_FILE"

if [[ -n "${ACCOUNTS_FILE+x}" ]]; then
  for i in $(jq '. | keys | .[]' "$ACCOUNTS_FILE"); do
    row=$(jq ".[$i]" "$ACCOUNTS_FILE")
    address=$(jq -r '.address' <<<"$row")
    amount=$(jq -r '.amount' <<<"$row")
    if [[ "$(jq -r '.vesting' <<<"$row")" != 'null' ]]; then
      add_vesting_account "$row" "$TMPDIR"
    else
      run_cmd "$MODE" "$TMPDIR" add-genesis-account "$address" "$amount"
    fi
  done
fi

cp "$GENESIS_FILE" "$OUTPUT_FILE"