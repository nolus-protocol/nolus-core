#!/bin/bash
set -euxo pipefail

PROTO_GENESIS="genesis.json"
VALIDATORS=1
IP_ADDRESSES=()
CUSTOM_IPS=false
OUTPUT_DIR="./validator_setup"
MONIKER="test-moniker-"
MODE="local"
KEYRING="test"
VAL_TOKENS="1000000000nomo"

POSITIONAL=()
while [[ $# -gt 0 ]]; do
  key="$1"

  case $key in
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
  --help)
    echo "Usage: ./init-test-network.sh [-g|--genesis <genesis_file>] [-v|--validators <num_validators>] [--validator-tokens <tokens_for_val_genesis_accounts>] [-ips <val_ip_addrs>] [-m|--mode <local|docker>] [-o|--output <output_dir>]"
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
  local) cosmzoned $@ --home "$DIR" 2>&1 ;;
  docker) docker run --rm -u "$(id -u)":"$(id -u)" -v "$DIR:/tmp/.cosmzone:Z" nomo/node $@ --home /tmp/.cosmzone 2>&1 ;;
  esac
}

init_genesis() {
  rm -rf keygenerator
  mkdir keygenerator
  ACCOUNTS_FILE='accounts.json'
  echo '[]' > "$ACCOUNTS_FILE"
  run_cmd "keygenerator" init "key-gen" --chain-id "nomo-private" --home .
  for i in $(seq "$VALIDATORS"); do
    local out=$(run_cmd "keygenerator" keys add "val_$i" --keyring-backend test --home . --output json)
     echo "$out"| jq -r .mnemonic > "val_${i}_mnemonic"
    address=$(run_cmd "keygenerator" keys show -a "val_${i}"  --keyring-backend test )
    append=$(jq ". += [{ \"address\": \"$address\", \"amount\":  \"$VAL_TOKENS\"}]" < "$ACCOUNTS_FILE")
    echo "$append" > "$ACCOUNTS_FILE"
  done
  cd ..
  ./proto-genesis.sh --accounts "$OUTPUT_DIR/$ACCOUNTS_FILE" --output "$OUTPUT_DIR/proto-genesis.json"
  cd "$OUTPUT_DIR"

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
    ../init-validator-node.sh -g 'proto-genesis.json' -d "node${i}" --moniker "validator-${i}" --mnemonic "$(cat "val_${i}_mnemonic")" --stake "1000000nomo" $IP
    cp -a "node${i}/config/gentx/." "./gentxs"
  done

  ../collect-validator-gentxs.sh --collector "node1" --gentxs "./gentxs"
  cp "node1/config/genesis.json" "genesis.json"

  # collect the generated messages in validator 1's node for collection and propagate the resulting genesis file
  for i in $(seq 2 "$VALIDATORS"); do
    local NODE_DIR="node${i}"
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


rm -rf "$OUTPUT_DIR"
mkdir -p "$OUTPUT_DIR"

cd "$OUTPUT_DIR"
OUTPUT_DIR=$(pwd)

init_genesis
init_local
