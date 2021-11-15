#!/bin/bash
set -euo pipefail

PROTO_GENESIS="genesis.json"
VALIDATORS=1
IP_ADDRESSES=()
CUSTOM_IPS=false
OUTPUT_DIR="./validator_setup"
MONIKER="test-moniker-"
MODE="local"
KEYRING="test"

POSITIONAL=()
while [[ $# -gt 0 ]]; do
  key="$1"

  case $key in
  -g | --genesis)
    PROTO_GENESIS="$2"
    shift # past argument
    shift # past value
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
  -ips)
    for i in $(echo "$2" | sed "s/,/ /g"); do
      IP_ADDRESSES+=("$i")
    done
    CUSTOM_IPS=true
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
  --help)
    echo "Usage: ./init-network.sh [-g|--genesis <genesis_file>] [-v|--validators <num_validators>] [-m|--mode <local|docker>] [-o|--output <output_dir>]"
    exit 0
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

init_local() {
  for i in $(seq "$VALIDATORS"); do
    # create node directory and initialize chain structure
    local NODE_DIR="$OUTPUT_DIR/node${i}"
    mkdir -p "$NODE_DIR"
    cd "$NODE_DIR"
    run_cmd "$NODE_DIR" init "$MONIKER${i}" --chain-id "$CHAINID"
    cp "$OUTPUT_DIR/proto-genesis.json" "$NODE_DIR/config/genesis.json"

    # enrich the genesis with a validator account which will later be available to the validator's node
    run_cmd "$NODE_DIR" keys add "$MONIKER${i}" --keyring-backend "$KEYRING"
    run_cmd "$NODE_DIR" add-genesis-account "$MONIKER${i}" 1000000000nomo --keyring-backend "$KEYRING"

    # replace proto-genesis with the validator's one
    cp "$NODE_DIR/config/genesis.json" "$OUTPUT_DIR/proto-genesis.json"
    cd "$OUTPUT_DIR"
  done

  # generate create validator messages for each validator
  for i in $(seq "$VALIDATORS"); do
    local NODE_DIR="$OUTPUT_DIR/node${i}"
    cd "$NODE_DIR"
    IP=""
    if [[ $CUSTOM_IPS = true ]]; then
      IP="--ip ${IP_ADDRESSES[$(("$i" - 1))]}"
    fi
    run_cmd "$NODE_DIR" gentx "$MONIKER${i}" 1000000nomo --keyring-backend "$KEYRING" --chain-id "$CHAINID" $IP
    cd "$OUTPUT_DIR"
  done

  # collect the generated messages in validator 1's node for collection and propagate the resulting genesis file
  cd "node1/config"
  cp "$OUTPUT_DIR/proto-genesis.json" "genesis.json" # replace the node1 genesis with the latest one, which may include accounts from other validators
  for i in $(seq 2 "$VALIDATORS"); do
    cp -a "$OUTPUT_DIR/node${i}/config/gentx/." "./gentx/"
  done

  run_cmd "$OUTPUT_DIR/node1" collect-gentxs
  cp "genesis.json" "$OUTPUT_DIR/genesis.json"
  cd "$OUTPUT_DIR"

  for i in $(seq 2 "$VALIDATORS"); do
    local NODE_DIR="$OUTPUT_DIR/node${i}"
    cp "genesis.json" "$NODE_DIR/config/"
  done
}

## validate dependencies are installed
command -v jq >/dev/null 2>&1 || {
  echo >&2 "jq not installed. More info: https://stedolan.github.io/jq/download/"
  exit 1
}

if [[ "$CUSTOM_IPS" = true && "${#IP_ADDRESSES[@]}" -ne "$VALIDATORS" ]]; then
  echo >&2 "non matching ip addesses"
  exit 1
fi

if [[ ! -f "$PROTO_GENESIS" ]]; then
  echo "Genesis file '$PROTO_GENESIS' not found, aborting"
  exit 1
fi
CHAINID=$(jq -r .chain_id <"$PROTO_GENESIS")

rm -rf "$OUTPUT_DIR"
mkdir -p "$OUTPUT_DIR"
cp "$PROTO_GENESIS" "$OUTPUT_DIR/proto-genesis.json"
cd "$OUTPUT_DIR"
OUTPUT_DIR=$(pwd)

init_local
