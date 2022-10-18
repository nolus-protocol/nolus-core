#!/bin/bash
set -euox pipefail

SCRIPT_GENESIS_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_GENESIS_DIR"/internal/genesis.sh
source "$SCRIPT_GENESIS_DIR"/internal/verify.sh

cleanup() {
  cleanup_genesis_sh
  exit
}
trap cleanup INT TERM EXIT

__print_usage() {
    printf \
    "Usage: %s
    <$COMMAND_FULL_GEN>
    [-c|--chain_id <string>]
    [--currency <native_currency>]
    [--accounts <accounts_spec_json>]
    [--wasm-script-path <wasm_script_path>]
    [--wasm-code-path <wasm_code_path>]
    [--wasm-admin-addr <wasm_admin_address>]
    [--treasury-nls-u128 <init treasury amount of uNLS>
    [--validator-node-urls-pubkeys <validator_node_urls_and_validator_pubkeys>]
    [--validator-accounts-dir <validator_accounts_dir>]
    [--validator-tokens <validators_initial_tokens>]
    [--validator-stake <tokens_validator_stakes>]
    [--lpp-native <lpp_native>]
    [-o|--output <genesis_file_path>]" \
     "$1"
}

COMMAND_FULL_GEN="full-gen"
CHAIN_ID=""
NATIVE_CURRENCY="unls"
ACCOUNTS_SPEC=""
WASM_SCRIPT_PATH=""
WASM_CODE_PATH=""
WASM_ADMIN_ADDR=""
TREASURY_INIT_TOKENS_U128=""
VAL_NODE_URLS_AND_VAL_PUBKEYS=""
VAL_ACCOUNTS_DIR="val-accounts"
VAL_TOKENS="1000000000""$NATIVE_CURRENCY"
VAL_STAKE="1000000""$NATIVE_CURRENCY"
OUTPUT_FILE=""
LPP_NATIVE=""
CONTRACTS_INFO_FILE="contracts-info.json"

if [[ $# -lt 1 ]]; then
  echo "Missing command!"
  __print_usage "$0"
  exit 1
fi
COMMAND="$1"
shift

while [[ $# -gt 0 ]]; do
  key="$1"

  case $key in

  -h | --help)
    __print_usage "$0"
    exit 0
    ;;

  -c | --chain-id)
    CHAIN_ID="$2"
    shift
    shift
    ;;

  --currency)
    NATIVE_CURRENCY="$2"
    shift
    shift
    ;;

  --accounts)
    ACCOUNTS_SPEC="$2"
    shift
    shift
    ;;

  --wasm-script-path)
    WASM_SCRIPT_PATH="$2"
    shift
    shift
    ;;

  --wasm-code-path)
    WASM_CODE_PATH="$2"
    shift
    shift
    ;;

  --wasm-admin-addr)
    WASM_ADMIN_ADDR="$2"
    shift
    shift
    ;;

  --treasury-nls-u128)
    TREASURY_INIT_TOKENS_U128="$2"
    shift
    shift
    ;;

  --validator-node-urls-pubkeys)
    VAL_NODE_URLS_AND_VAL_PUBKEYS="$2"
    shift
    shift
    ;;

  --validator-accounts-dir)
    VAL_ACCOUNTS_DIR="$2"
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

  --lpp-native)
    LPP_NATIVE="$2"
    shift
    shift
    ;;

  -o | --output)
    OUTPUT_FILE="$2"
    shift
    shift
    ;;

  *)
    echo "unknown option '$key'"
    exit 1
    ;;

  esac
done

if [[ "$COMMAND" == "$COMMAND_FULL_GEN" ]]; then
  verify_mandatory "$CHAIN_ID" "Nolus chain identifier"
  verify_mandatory "$ACCOUNTS_SPEC" "Nolus genesis accounts spec"
  verify_mandatory "$WASM_SCRIPT_PATH" "Wasm script path"
  verify_mandatory "$WASM_CODE_PATH" "Wasm code path"
  verify_mandatory "$WASM_ADMIN_ADDR" "Wasm admin address"
  verify_mandatory "$TREASURY_INIT_TOKENS_U128" "Treasury init tokens"
  verify_mandatory "$VAL_NODE_URLS_AND_VAL_PUBKEYS" "Validator URLs and validator public keys spec"
  verify_mandatory "$LPP_NATIVE" "Lpp native currency symbol"
  verify_mandatory "$OUTPUT_FILE" "Genesis output file"

  genesis_file=$(generate_genesis "$CHAIN_ID" "$NATIVE_CURRENCY" "$VAL_TOKENS" "$VAL_STAKE" \
                                  "$VAL_ACCOUNTS_DIR" "$ACCOUNTS_SPEC" "$WASM_SCRIPT_PATH" \
                                  "$WASM_CODE_PATH" "$WASM_ADMIN_ADDR" "$TREASURY_INIT_TOKENS_U128" \
                                  "$VAL_NODE_URLS_AND_VAL_PUBKEYS" "$LPP_NATIVE" \
                                  "$CONTRACTS_INFO_FILE")
  mv "$genesis_file" "$OUTPUT_FILE"
# elif [[ "$COMMAND" == "$COMMAND_SETUP" ]]; then
#
else
  echo "Unknown command!"
  exit 1
fi
