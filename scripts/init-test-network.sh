#!/bin/bash
set -euxo pipefail

command -v create-vesting-account.sh >/dev/null 2>&1 || {
  echo >&2 "scripts are not found in \$PATH."
  exit 1
}

ORIG_DIR=$(pwd)

cleanup() {
  if [ -n "${ORIG_DIR:-}" ]; then
    cd "$ORIG_DIR"
    exit
  fi
}

trap cleanup INT TERM EXIT

VALIDATORS=1
IP_ADDRESSES=()
CUSTOM_IPS=false
POSITIONAL=()

MONIKER="test-moniker-"
MODE="local"
KEYRING="test"
NATIVE_CURRENCY="nolus"
VAL_TOKENS="1000000000nolus"
VAL_STAKE="1000000nolus"
CHAIN_ID="nolus-private"
OUTPUT_DIR="./validator_setup"

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
  rm -rf keygenerator
  mkdir keygenerator
  ACCOUNTS_FILE="accounts.json"
  echo '[]' > "$ACCOUNTS_FILE"
  run_cmd "keygenerator" init "key-gen" --chain-id "nolus-private"
  for i in $(seq "$VALIDATORS"); do
    local out
    out=$(run_cmd "keygenerator" keys add "val_$i" --keyring-backend test --output json)
    echo "$out"| jq -r .mnemonic > "val_${i}_mnemonic"
    address=$(run_cmd "keygenerator" keys show -a "val_${i}" --keyring-backend test)
    append=$(jq ". += [{ \"address\": \"$address\", \"amount\":  \"$VAL_TOKENS\"}]" < "$ACCOUNTS_FILE")
    echo "$append" > "$ACCOUNTS_FILE"
  done

  penultimate-genesis.sh --chain-id "$CHAIN_ID" --accounts "$ACCOUNTS_FILE" --currency "$NATIVE_CURRENCY" --output "penultimate-genesis.json"

  rm "$ACCOUNTS_FILE"
}

init_local() {
  rm -rf gentxs
  mkdir gentxs
  for i in $(seq "$VALIDATORS"); do
    IP=""
    if [[ $CUSTOM_IPS = true ]]; then
      IP="--ip ${IP_ADDRESSES[$(("$i" - 1))]}"
    fi
    init-validator-node.sh -g "penultimate-genesis.json" -d "node${i}" --moniker "validator-${i}" --mnemonic "$(cat "val_${i}_mnemonic")" --stake "$VAL_STAKE" "$IP"
    cp -a "node${i}/config/gentx/." "gentxs"
  done

  collect-validator-gentxs.sh --collector "node1" --gentxs "gentxs"
  cp "node1/config/genesis.json" "genesis.json"

  # collect the generated messages in validator 1's node for collection and propagate the resulting genesis file
  for i in $(seq 2 "$VALIDATORS"); do
    cp "genesis.json" "node${1}/config/"
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
cd "$OUTPUT_DIR"

init_genesis
init_local
