#!/bin/bash
set -euo pipefail

cleanup() {
  if [[ -n "${TMPDIR:-}" ]]; then
    rm -rf "$TMPDIR"
    exit
  fi
}

trap cleanup INT TERM EXIT

CHAINID="nomo-private"
OUTPUT_FILE="genesis.json"
MODE="local"
ACCOUNTS_FILE=""
TMPDIR=$(mktemp -d)

POSITIONAL=()
while [[ $# -gt 0 ]]; do
  key="$1"

  case $key in
  -c | --chain-id)
    CHAINID="$2"
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
    echo "Usage: ./init-proto-genesis.sh [-c|--chain-id <chain_id>] [-o|--output <output_file>] [-m|--mode <local|docker>] [--accounts <accounts_file>]"
    exit 0
    ;;
  *) # unknown option
    POSITIONAL+=("$1") # save it in an array for later
    shift              # past argument
    ;;
  esac
done

update_genesis() {
  jq "$1" <"$TMPDIR/config/genesis.json" >"$TMPDIR/config/tmp_genesis.json" && mv "$TMPDIR/config/tmp_genesis.json" "$TMPDIR/config/genesis.json"
}

run_cmd() {
  local DIR="$1"
  shift
  case $MODE in
  local) cosmzoned $@ --home "$DIR" ;;
  docker) docker run --rm -u "$(id -u)":"$(id -u)" -v "$DIR:/tmp/.cosmzone:Z" nomo/node $@ --home /tmp/.cosmzone ;;
  esac
}

# validate dependencies are installed
command -v jq >/dev/null 2>&1 || {
  echo >&2 "jq not installed. More info: https://stedolan.github.io/jq/download/"
  exit 1
}
MONIKER="localtestnet"
KEYRING="test"

ORIG_DIR=$(pwd)
cd "$TMPDIR"
run_cmd "$TMPDIR" init $MONIKER --chain-id "$CHAINID" --home .
run_cmd "$TMPDIR" config keyring-backend "$KEYRING" --home .
run_cmd "$TMPDIR" config chain-id "$CHAINID" --home .

# Change parameter token denominations to nomo
update_genesis '.app_state["staking"]["params"]["bond_denom"]="nomo"'
update_genesis '.app_state["crisis"]["constant_fee"]["denom"]="nomo"'
update_genesis '.app_state["gov"]["deposit_params"]["min_deposit"][0]["denom"]="nomo"'
update_genesis '.app_state["gov"]["deposit_params"]["min_deposit"][0]["denom"]="nomo"'
update_genesis '.app_state["mint"]["params"]["mint_denom"]="nomo"'

if [[ -n "${ACCOUNTS_FILE+x}" ]]; then
  for i in $(jq '. | keys | .[]' "$ACCOUNTS_FILE"); do
    row=$(jq ".[$i]" "$ACCOUNTS_FILE")
    address=$(jq -r  '.address' <<< "$row")
    amount=$(jq -r  '.amount' <<< "$row")
    run_cmd "$TMPDIR" add-genesis-account "$address" "$amount" --home .
  done
fi

cd "$ORIG_DIR"
cp "$TMPDIR/config/genesis.json" "$OUTPUT_FILE"
