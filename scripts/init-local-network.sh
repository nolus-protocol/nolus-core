#!/bin/bash
set -euxo pipefail

SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
source "$SCRIPT_DIR"/common/cmd.sh
source "$SCRIPT_DIR"/internal/accounts.sh
source "$SCRIPT_DIR"/internal/verify.sh
source "$SCRIPT_DIR"/internal/leaser-dex-setup.sh

cleanup() {
  cleanup_init_network_sh
  exit
}
trap cleanup INT TERM EXIT

NOLUS_NET_RPC="http://localhost:26612/"
VALIDATORS=1
VALIDATORS_ROOT_DIR="networks/nolus"
VAL_ACCOUNTS_DIR="$VALIDATORS_ROOT_DIR/val-accounts"
USER_DIR="$HOME/.nolus"
HERMES_BINARY_DIR="$HOME/hermes"

NATIVE_CURRENCY="unls"

VAL_TOKENS="1000000000""$NATIVE_CURRENCY"
VAL_STAKE="1000000""$NATIVE_CURRENCY"
WASM_SCRIPT_PATH="$SCRIPT_DIR/../../nolus-money-market/scripts"
WASM_CODE_PATH="$SCRIPT_DIR/../../nolus-money-market/artifacts"
CHAIN_ID="nolus-local"
TREASURY_NLS_U128="1000000000000"
RESERVE_NAME="reserve"
RESERVE_TOKENS="1000000000""$NATIVE_CURRENCY"
LPP_NATIVE_TICKER="USDC"
CONTRACTS_INFO_FILE="contracts-info.json"

HERMES_BINARY_DIR="$HOME/hermes"
HERMES_ADDRESS=""
A_CHAIN=""
B_CHAIN=""

while [[ $# -gt 0 ]]; do
  key="$1"

  case $key in

  -h | --help)
    printf \
    "Usage: %s
    [--chain-id <string>]
    [--nolus-net-rpc <string>]
    [--currency <native_currency>]
    [-v|--validators <number>]
    [--validators-root-dir <validators_root_dir>]
    [--validator-accounts-dir <validator_accounts_dir>]
    [--validator-tokens <validators_initial_tokens>]
    [--validator-stake <tokens_validator_stakes>]
    [--wasm-script-path <wasm_script_path>]
    [--wasm-code-path <wasm_code_path>]
    [--treasury-nls-u128 <treasury_initial_Nolus_tokens>]
    [--reserve-tokens <initial_reserve_tokens>]
    [--lpp-native <lpp_native>]
    [--user-dir <client_user_dir>]
    [--hermes-binary-dir <hermes_binary_dir>]
    [--hermes-address <hermes_address_nolus]
    [--a-chain <configured_hermes_chain_1_id]
    [--b-chain <configured_hermes_chain_2_id]" \
    "$0"
    exit 0
    ;;

  --chain-id)
    CHAIN_ID="$2"
    shift
    shift
    ;;

  --nolus-net-rpc)
    NOLUS_NET_RPC="$2"
    shift
    shift
    ;;

  --currency)
    NATIVE_CURRENCY="$2"
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

  --validators-root-dir)
    VALIDATORS_ROOT_DIR="$2"
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

  --treasury-nls-u128)
    TREASURY_NLS_U128="$2"
    shift
    shift
    ;;

  --reserve-tokens)
    RESERVE_TOKENS="$2"
    shift
    shift
    ;;

  --lpp-native)
    LPP_NATIVE_TICKER="$2"
    shift
    shift
    ;;

  --user-dir)
    USER_DIR="$2"
    shift
    shift
    ;;

  --hermes-binary-dir)
    HERMES_BINARY_DIR="$2"
    shift
    shift
    ;;

  --hermes-address)
    HERMES_ADDRESS="$2"
    shift
    shift
    ;;

  --a-chain)
    A_CHAIN="$2"
    shift
    shift
    ;;

  --b-chain)
    B_CHAIN="$2"
    shift
    shift
    ;;

  *)
    echo >&2 "The provided option '$key' is not recognized"
    exit 1
    ;;

  esac
done

__config_client() {
  run_cmd "$USER_DIR" config chain-id "$CHAIN_ID"
  run_cmd "$USER_DIR" config keyring-backend "test"
  run_cmd "$USER_DIR" config node "tcp://localhost:$(first_node_rpc_port)"
}

verify_dir_exist "$WASM_SCRIPT_PATH" "wasm sripts path"
verify_dir_exist "$WASM_CODE_PATH" "wasm code path"
verify_mandatory "$A_CHAIN" "configured hermes chain 1 id"
verify_mandatory "$B_CHAIN" "configured hermes chain 2 id"
verify_mandatory "$HERMES_ADDRESS" "hermes address nolus"

rm -fr "$VALIDATORS_ROOT_DIR"
rm -fr "$VAL_ACCOUNTS_DIR"
rm -fr "$USER_DIR"

accounts_spec=$(echo "[]" | add_account "$(generate_account "$RESERVE_NAME" "$USER_DIR")" "$RESERVE_TOKENS")
contracts_owner_name="contracts_owner"
contracts_owner_addr=$(generate_account "$contracts_owner_name" "$USER_DIR")
# We handle the contracts_owner account as normal address.
treasury_init_tokens="$TREASURY_NLS_U128$NATIVE_CURRENCY"
accounts_spec=$(echo "$accounts_spec" | add_account "$contracts_owner_addr" "$treasury_init_tokens")
# accounts_spec=$(echo "$accounts_spec" | add_vesting_account "$contracts_owner_addr" "1000020000000$NATIVE_CURRENCY" \
#                 "20000000" "2022-10-31T17:15:59+02:00" "2022-10-31T17:30:00+02:00")

source "$SCRIPT_DIR"/internal/setup-validator-local.sh
init_setup_validator_local_sh "$SCRIPT_DIR" "$VALIDATORS_ROOT_DIR"

source "$SCRIPT_DIR"/internal/init-network.sh
init_network "$VAL_ACCOUNTS_DIR" "$VALIDATORS" "$CHAIN_ID" "$NATIVE_CURRENCY" \
              "$VAL_TOKENS" "$VAL_STAKE" "$accounts_spec" \
              "$WASM_SCRIPT_PATH" "$WASM_CODE_PATH" \
              "$contracts_owner_addr" "$TREASURY_NLS_U128" \
              "$LPP_NATIVE_TICKER" "$CONTRACTS_INFO_FILE"

__config_client

run_cmd "$VALIDATORS_ROOT_DIR/local-validator-1" start &>"$USER_DIR"/nolus_logs.txt & disown;

declare nolus_node_status=""
while [ "$nolus_node_status" != "STARTED" ]
do
   sleep 1
   run_cmd "$USER_DIR" status && nolus_node_status="STARTED" || nolus_node_status="ERROR"
done

leaser_dex_setup "$NOLUS_NET_RPC" "$USER_DIR" "$contracts_owner_name" "$RESERVE_NAME" "$CONTRACTS_INFO_FILE" "$HERMES_BINARY_DIR" "$HERMES_ADDRESS" "$A_CHAIN" "$B_CHAIN"

"$HERMES_BINARY_DIR"/hermes start &>"$HERMES_BINARY_DIR"/hermes_logs.txt & disown;
