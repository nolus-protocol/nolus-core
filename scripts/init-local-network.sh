#!/bin/bash
set -euxo pipefail

INIT_LOCAL_NETWORK_SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
source "$INIT_LOCAL_NETWORK_SCRIPT_DIR"/common/cmd.sh
source "$INIT_LOCAL_NETWORK_SCRIPT_DIR"/internal/accounts.sh
source "$INIT_LOCAL_NETWORK_SCRIPT_DIR"/internal/verify.sh
source "$INIT_LOCAL_NETWORK_SCRIPT_DIR"/internal/wait_services.sh
source "$INIT_LOCAL_NETWORK_SCRIPT_DIR"/internal/leaser-dex-setup.sh

cleanup() {
  cleanup_init_network_sh
  exit
}
trap cleanup INT TERM EXIT

VALIDATORS=1
VALIDATORS_ROOT_DIR="networks/nolus"
VAL_ACCOUNTS_DIR="$VALIDATORS_ROOT_DIR/val-accounts"
USER_DIR="$HOME/.nolus"
NATIVE_CURRENCY="unls"

TAG=$(git describe --tags)
VAL_TOKENS="1000000000""$NATIVE_CURRENCY"
VAL_STAKE="1000000""$NATIVE_CURRENCY"
WASM_SCRIPT_PATH="$INIT_LOCAL_NETWORK_SCRIPT_DIR/../../nolus-money-market/scripts"
WASM_CODE_PATH="$INIT_LOCAL_NETWORK_SCRIPT_DIR/../../nolus-money-market/artifacts"
# Tags would look like (v0.1.37-60-g321dbd1), we want to cut the last part(abbreviated object name)
VERSION=$(cut -f1,2 -d'-' <<< "$TAG")
# date +%s returns the number of seconds since the epoch
CHAIN_ID="nolus-local""-$VERSION-$(date +%s)"
TREASURY_NLS_U128="1000000000000"
RESERVE_NAME="reserve"
RESERVE_TOKENS="1000000000""$NATIVE_CURRENCY"
LPP_NATIVE_TICKER="USDC"
CONTRACTS_INFO_FILE="contracts-info.json"

NOLUS_NETWORK_ADDR="127.0.0.1"
NOLUS_NETWORK_RPC_PORT="26612"
NOLUS_NETWORK_GRPC_PORT="26615"
DEX_NETWORK_ID="osmo-test-4"
DEX_NETWORK_ADDR="osmo-net.nolus.io"
DEX_NETWORK_RPC_PORT="26657"
DEX_NETWORK_GRPC_PORT="9090"
HERMES_ACCOUNT_MNEMONIC=""

GOV_VOTING_PERIOD="300s"
FEEREFUNDER_ACK_FEE_MIN="1"
FEEREFUNDER_TIMEOUT_FEE_MIN="1"

NOLUS_NET="http://localhost:$NOLUS_NETWORK_RPC_PORT/"

while [[ $# -gt 0 ]]; do
  key="$1"

  case $key in

  -h | --help)
    printf \
    "Usage: %s
    [--chain-id <string>]
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
    [--gov-voting-period <voting_period>]
    [--user-dir <client_user_dir>]
    [--dex-network-id <dex_network_id]
    [--dex-network-addr <dex_network_addr>]
    [--dex-network-rpc-port <dex_network_rpc_port>]
    [--dex-network-grpc-port <dex_network_grpc_port>]
    [--hermes-mnemonic <hermes_account_mnemonic]
    [--feerefunder-ack-fee-min <feerefunder_ack_fee_min_amount>]
    [--feerefunder-timeout-fee-min <feerefunder_timeout_fee_min_amount>]" \
    "$0"
    exit 0
    ;;

  --chain-id)
    CHAIN_ID="$2"
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

  --gov-voting-period)
    GOV_VOTING_PERIOD="$2"
    shift
    shift
    ;;

  --user-dir)
    USER_DIR="$2"
    shift
    shift
    ;;

  --dex-network-id)
    DEX_NETWORK_ID="$2"
    shift
    shift
    ;;


  --dex-network-addr)
    DEX_NETWORK_ADDR="$2"
    shift
    shift
    ;;

  --dex-network-rpc-port)
    DEX_NETWORK_RPC_PORT="$2"
    shift
    shift
    ;;

  --dex-network-grpc-port)
    DEX_NETWORK_GRPC_PORT="$2"
    shift
    shift
    ;;

  --hermes-mnemonic)
    HERMES_ACCOUNT_MNEMONIC="$2"
    shift
    shift
    ;;

  --feerefunder-ack-fee-min)
    FEEREFUNDER_ACK_FEE_MIN="$2"
    shift
    shift
    ;;

  --feerefunder-timeout-fee-min)
    FEEREFUNDER_TIMEOUT_FEE_MIN="$2"
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

__nolus_hermes_address() {
  local -r hermes_binary_dir="$1"
  local -r chain_id="$2"

  local address
  address="$("$hermes_binary_dir"/hermes keys list --chain "$chain_id")"
  # Look for 'Parameter Expansion' at https://linux.die.net/man/1/bash
  # First strip off the prefix
  address="${address#*nolus (}"
  # Then the suffix
  echo "${address%)}"
}

verify_dir_exist "$WASM_SCRIPT_PATH" "wasm sripts path"
verify_dir_exist "$WASM_CODE_PATH" "wasm code path"
verify_mandatory "$HERMES_ACCOUNT_MNEMONIC" "hermes account mnemonic"

rm -fr "$VALIDATORS_ROOT_DIR"
rm -fr "$VAL_ACCOUNTS_DIR"
rm -fr "$USER_DIR"

accounts_spec=$(echo "[]" | add_account "$(generate_account "$RESERVE_NAME" "$USER_DIR")" "$RESERVE_TOKENS")

source "$INIT_LOCAL_NETWORK_SCRIPT_DIR"/internal/setup-validator-local.sh
init_setup_validator_local_sh "$INIT_LOCAL_NETWORK_SCRIPT_DIR" "$VALIDATORS_ROOT_DIR"

source "$INIT_LOCAL_NETWORK_SCRIPT_DIR"/internal/init-network.sh
init_network "$VAL_ACCOUNTS_DIR" "$VALIDATORS" "$CHAIN_ID" "$NATIVE_CURRENCY" \
              "$VAL_TOKENS" "$VAL_STAKE" "$accounts_spec" \
              "$WASM_SCRIPT_PATH" "$WASM_CODE_PATH" \
              "$TREASURY_NLS_U128" \
              "$LPP_NATIVE_TICKER" "$CONTRACTS_INFO_FILE" \
              "$GOV_VOTING_PERIOD" "$FEEREFUNDER_ACK_FEE_MIN" "$FEEREFUNDER_TIMEOUT_FEE_MIN"

__config_client

run_cmd "$VALIDATORS_ROOT_DIR/local-validator-1" start &>"$USER_DIR"/nolus_logs.txt & disown;

/bin/bash "$INIT_LOCAL_NETWORK_SCRIPT_DIR"/remote/hermes-config.sh "$HOME" "$CHAIN_ID" "$NOLUS_NETWORK_ADDR" \
                                                "$NOLUS_NETWORK_RPC_PORT" "$NOLUS_NETWORK_GRPC_PORT" \
                                                "$DEX_NETWORK_ID" "$DEX_NETWORK_ADDR" "$DEX_NETWORK_RPC_PORT" \
                                                "$DEX_NETWORK_GRPC_PORT" "$HERMES_ACCOUNT_MNEMONIC"

HERMES_BINARY_DIR="$HOME"/hermes

wait_nolus_gets_ready "$USER_DIR"
wait_hermes_config_gets_healthy "$HERMES_BINARY_DIR"

NOLUS_HERMES_ADDRESS=$(__nolus_hermes_address "$HERMES_BINARY_DIR" "$CHAIN_ID")

leaser_dex_setup "$NOLUS_NET" "$USER_DIR" "$RESERVE_NAME" "$HERMES_BINARY_DIR" "$NOLUS_HERMES_ADDRESS" \
                 "$CHAIN_ID" "$DEX_NETWORK_ID"

"$HERMES_BINARY_DIR"/hermes start &>"$HERMES_BINARY_DIR"/hermes_logs.txt & disown;
