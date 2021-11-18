#!/bin/bash
set -euxo pipefail

ROOT_DIR=$(pwd)
VALIDATORS=1
IP_ADDRESSES=()
CUSTOM_IPS=false
POSITIONAL=()

MONIKER="test-moniker-"
MODE="local"
KEYRING="test"
VAL_TOKENS="1000000000nomo"
SCRIPTS_DIR="$ROOT_DIR/scripts"
OUTPUT_DIR="$ROOT_DIR/validator_setup"

while [[ $# -gt 0 ]]; do
  key="$1"

  case $key in

  -h | --help)
    printf \
    "Usage: $0\n
    [-r| --root <root_directory>]\n
    [-v|--validators <num_validators>]\n
    [--validator-tokens <tokens_for_val_genesis_accounts>]\n
    [-ips <val_ip_addrs>]\n
    [-m|--mode <local|docker>]\n
    [-o|--output <output_dir>]\n"
    exit 0
    ;;

   -r | --root)
    ROOT_DIR="$2"
    [ -z "$ROOT_DIR"] || {
      echo >&2 "root directory not provided"
      exit 1
    }
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

  --validator-tokens)
    VAL_TOKENS="$2"
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
  local) cosmzoned $@ --home "$DIR" 2>&1 ;;
  docker) docker run --rm -u "$(id -u)":"$(id -u)" -v "$DIR:/tmp/.cosmzone:Z" nomo/node $@ --home /tmp/.cosmzone 2>&1 ;;
  esac
}

init_genesis() {
  rm -rf $OUTPUT_DIR/keygenerator
  mkdir $OUTPUT_DIR/keygenerator
  ACCOUNTS_FILE="$OUTPUT_DIR/accounts.json"
  echo '[]' > "$ACCOUNTS_FILE"
  run_cmd "$OUTPUT_DIR/keygenerator" init "key-gen" --chain-id "nomo-private" --home .
  for i in $(seq "$VALIDATORS"); do
    local out=$(run_cmd "$OUTPUT_DIR/keygenerator" keys add "val_$i" --keyring-backend test --home . --output json)
    echo "$out"| jq -r .mnemonic > "$OUTPUT_DIR/val_${i}_mnemonic"
    address=$(run_cmd "$OUTPUT_DIR/keygenerator" keys show -a "val_${i}" --keyring-backend test)
    append=$(jq ". += [{ \"address\": \"$address\", \"amount\":  \"$VAL_TOKENS\"}]" < "$ACCOUNTS_FILE")
    echo "$append" > "$ACCOUNTS_FILE"
  done

  $SCRIPTS_DIR/proto-genesis.sh --accounts "$ACCOUNTS_FILE" --output "$OUTPUT_DIR/proto-genesis.json"

  rm "$ACCOUNTS_FILE"
}

init_local() {
  rm -rf $OUTPUT_DIR/gentxs
  mkdir $OUTPUT_DIR/gentxs
  for i in $(seq "$VALIDATORS"); do
    IP=""
    if [[ $CUSTOM_IPS = true ]]; then
      IP="--ip ${IP_ADDRESSES[$(("$i" - 1))]}"
    fi
    $SCRIPTS_DIR/init-validator-node.sh -g "$OUTPUT_DIR/proto-genesis.json" -d "$OUTPUT_DIR/node${i}" --moniker "validator-${i}" --mnemonic "$(cat "$OUTPUT_DIR/val_${i}_mnemonic")" --stake "1000000nomo" $IP
    cp -a "$OUTPUT_DIR/node${i}/config/gentx/." "$OUTPUT_DIR/gentxs"
  done

  $SCRIPTS_DIR/collect-validator-gentxs.sh --collector "$OUTPUT_DIR/node1" --gentxs "$OUTPUT_DIR/gentxs"
  cp "$OUTPUT_DIR/node1/config/genesis.json" "$OUTPUT_DIR/genesis.json"

  # collect the generated messages in validator 1's node for collection and propagate the resulting genesis file
  for i in $(seq 2 "$VALIDATORS"); do
    cp "$OUTPUT_DIR/genesis.json" "$OUTPUT_DIR/node${1}/config/"
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

rm -rf "$OUTPUT_DIR"
mkdir -p "$OUTPUT_DIR"

init_genesis
init_local
