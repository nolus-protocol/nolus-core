#!/bin/bash
set -euxo pipefail

SCRIPT_DIR=$(cd $(dirname "${BASH_SOURCE[0]}") && pwd)

source $SCRIPT_DIR/internal/local.sh
source $SCRIPT_DIR/internal/accounts.sh

VALIDATORS=1
IP_ADDRESSES=()
CUSTOM_IPS=false
POSITIONAL=()

MONIKER="test-moniker-"
MODE="local"
KEYRING="test"
NATIVE_CURRENCY="unolus"
VAL_TOKENS="1000000000unolus"
VAL_STAKE="1000000unolus"
CHAIN_ID="nolus-private"
OUTPUT_DIR="dev-net"

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

ACCOUNTS_FILE="$OUTPUT_DIR/accounts.json"
PROTO_GENESIS_FILE="$OUTPUT_DIR/penultimate-genesis.json"

# Init validator nodes, generate validator accounts and collect their addresses
#
# The nodes are placed in sub directories of $OUTPUT_DIR
# The validator addresses are printed on the standard output one at a line
init_nodes() {
  rm -fr "$OUTPUT_DIR"
  mkdir "$OUTPUT_DIR"

  for i in $(seq "$VALIDATORS"); do
    local node_id="dev-validator-$i"

    deploy "$OUTPUT_DIR" "$node_id" "$CHAIN_ID"
    local address=$(gen_account "$OUTPUT_DIR" "$node_id")
    echo "$address"
  done
}

gen_pre_genesis() {
  echo ""
}

init_local() {
  rm -rf gentxs
  mkdir gentxs
  for i in $(seq "$VALIDATORS"); do
    IP=""
    if [[ $CUSTOM_IPS = true ]]; then
      IP="--ip ${IP_ADDRESSES[$(("$i" - 1))]}"
    fi
    init-validator-node.sh -g "$PROTO_GENESIS_FILE" -d "node${i}" --moniker "validator-${i}" --mnemonic "$(cat "val_${i}_mnemonic")" --stake "$VAL_STAKE" "$IP" --mode "$MODE"
    cp -a "node${i}/config/gentx/." "gentxs"
  done

  collect-validator-gentxs.sh --collector "node1" --gentxs "gentxs" --mode "$MODE"
  cp "node1/config/genesis.json" "genesis.json"

  # collect the generated messages in validator 1's node for collection and propagate the resulting genesis file
  for i in $(seq 2 "$VALIDATORS"); do
    cp "genesis.json" "node${i}/config/"
  done
}

gen_accounts_spec() {
  local addresses="$1"
  local file="$2"

  local accounts="[]"
  for address in $addresses; do
    accounts=$(echo "$accounts" | add_account "$address" "10$NATIVE_CURRENCY")
  done
  echo "$accounts" > "$file"
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

addresses="$(init_nodes)"
gen_accounts_spec "$addresses" "$ACCOUNTS_FILE"
"$SCRIPT_DIR"/penultimate-genesis.sh --chain-id "$CHAIN_ID" --accounts "$ACCOUNTS_FILE" --currency "$NATIVE_CURRENCY" \
  --output "$PROTO_GENESIS_FILE"

#gen_pre_genesis
#init_genesis
#init_local
