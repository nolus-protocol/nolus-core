#!/bin/bash
set -euxo pipefail

GENESIS="genesis.json"
IP_ADDRESS=""
MNEMONIC=""
NODE_DIR=""
MONIKER="test-moniker"
MODE="local"
STAKE="1000000nolus"
KEYRING="test"

POSITIONAL=()
while [[ $# -gt 0 ]]; do
  key="$1"

  case $key in
  --help)
    printf \
    "Usage: %s
    [-g|--genesis <genesis_file>]
    [-ip <validator_ip_addresses>]
    [-d|--directory <full_node_directory>]
    [--mnemonic <mnemonic>]
    [--moniker <moniker>]
    [--stake <validator_stake>]
    [-m|--mode <local|docker>]" "$0"
    exit 0
    ;;
  -g | --genesis)
    GENESIS=$(realpath "$2")
    shift # past argument
    shift # past value
    ;;
  --ip)
    IP_ADDRESS="$2"
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
  -d | --directory)
    NODE_DIR="$2"
    shift
    shift
    ;;
  --mnemonic)
    MNEMONIC="$2"
    shift
    shift
    ;;
  --moniker)
    MONIKER="$2"
    shift
    shift
    ;;
  --stake)
    STAKE="$2"
    shift
    shift
    ;;
  *) # unknown option
    POSITIONAL+=("$1") # save it in an array for later
    shift              # past argument
    ;;
  esac
done

run_cmd() {
  local DIR="$1"
  shift
  case $MODE in
  local) cosmzoned $@ --home "$DIR" ;;
  docker) docker run --rm -u "$(id -u)":"$(id -u)" -v "$DIR:/tmp/.cosmzone:Z" nomo/node $@ --home /tmp/.cosmzone ;;
  esac
}

## validate dependencies are installed
command -v jq >/dev/null 2>&1 || {
  echo >&2 "jq not installed. More info: https://stedolan.github.io/jq/download/"
  exit 1
}


if [[ -z "$NODE_DIR" ]]; then
  echo "NODE_DIR is unset"
   exit 1
fi


if [[ -z "$MNEMONIC" ]]; then
  echo "MNEMONIC is unset"
   exit 1
fi

if [[ ! -f "$GENESIS" ]]; then
  echo "Genesis file '$GENESIS' not found"
  exit 1
fi
CHAINID=$(jq -r .chain_id <"$GENESIS")

WORKING_DIR=$(pwd)
rm -rf "$NODE_DIR"
mkdir -p "$NODE_DIR"

run_cmd "$NODE_DIR" init "$MONIKER" --chain-id "$CHAINID"
cp "$GENESIS" "$NODE_DIR/config/genesis.json"

run_cmd "$NODE_DIR" keys add --recover "validator-key" --keyring-backend "$KEYRING" <<< "$MNEMONIC"
IP=""
if [[ -n "${IP_ADDRESS+}" ]]; then
  IP="--ip $IP_ADDRESS"
fi
run_cmd "$NODE_DIR" gentx "validator-key" "$STAKE" --keyring-backend "$KEYRING" --chain-id "$CHAINID" $IP
cd "$WORKING_DIR"