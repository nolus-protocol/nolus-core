#!/bin/bash
set -euo pipefail

SCRIPT_DIR=$(cd $(dirname "${BASH_SOURCE[0]}") && pwd)

source "$SCRIPT_DIR"/create-vesting-account.sh
source "$SCRIPT_DIR"/internal/cmd.sh
source "$SCRIPT_DIR"/internal/genesis.sh

cleanup() {
  if [[ -n "${TMPDIR:-}" ]]; then
    rm -rf "$TMPDIR"
  fi
  exit
}

trap cleanup INT TERM EXIT

CHAIN_ID="nolus-private"
OUTPUT_FILE="genesis.json"
ACCOUNTS_FILE=""
SUSPEND_ADMIN=""
TMPDIR=$(mktemp -d)
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
  --suspend-admin)
    SUSPEND_ADMIN="$2"
    shift
    shift
    ;;
  --help)
    echo "Usage: penultimate-genesis.sh [-c|--chain-id <chain_id>] [-o|--output <output_file>] [--accounts <accounts_file>] [--currency <native_currency>] [--suspend-admin <bech32address>]"
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

generate_proto_genesis "$TMPDIR" "$CHAIN_ID" "$ACCOUNTS_FILE" "$NATIVE_CURRENCY" "$OUTPUT_FILE" "$SUSPEND_ADMIN"
