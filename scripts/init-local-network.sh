#!/bin/bash
set -euxo pipefail

INIT_LOCAL_NETWORK_SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
source "$INIT_LOCAL_NETWORK_SCRIPT_DIR"/common/cmd.sh
source "$INIT_LOCAL_NETWORK_SCRIPT_DIR"/common/rm-dir.sh
source "$INIT_LOCAL_NETWORK_SCRIPT_DIR"/internal/accounts.sh
source "$INIT_LOCAL_NETWORK_SCRIPT_DIR"/internal/verify.sh
source "$INIT_LOCAL_NETWORK_SCRIPT_DIR"/internal/wait_services.sh
source "$INIT_LOCAL_NETWORK_SCRIPT_DIR"/internal/add-dex-support.sh

cleanup() {
  cleanup_init_network_sh
  exit
}
trap cleanup INT TERM EXIT

TAG=$(git describe --tags)
# Tags would look like (v0.1.37-60-g321dbd1), we want to cut the last part(abbreviated object name)
VERSION=$(cut -f1,2 -d'-' <<< "$TAG")
# date +%s returns the number of seconds since the epoch
NOLUS_NETWORK_ADDR="127.0.0.1"
NOLUS_NETWORK_RPC_PORT="26612"
NOLUS_NETWORK_GRPC_PORT="26615"
NATIVE_CURRENCY="unls"
VALIDATORS=1
VALIDATORS_ROOT_DIR="networks/nolus"
VAL_ACCOUNTS_DIR="$VALIDATORS_ROOT_DIR/val-accounts"
USER_DIR="$HOME/.nolus"
VAL_TOKENS="1000000000""$NATIVE_CURRENCY"
VAL_STAKE="1000000""$NATIVE_CURRENCY"
CHAIN_ID="nolus-local""-$VERSION-$(date +%s)"
TREASURY_NLS_U128="1000000000000"
RESERVE_NAME="reserve"
RESERVE_TOKENS="1000000000""$NATIVE_CURRENCY"
GOV_VOTING_PERIOD="300s"
FEEREFUNDER_ACK_FEE_MIN="1"
FEEREFUNDER_TIMEOUT_FEE_MIN="1"
DEX_ADMIN_MNEMONIC=""
HERMES_ACCOUNT_MNEMONIC=""
HERMES_VERSION=""
STORE_CODE_PRIVILEGED_ACCOUNT_MNEMONIC=""

# Hermes - Nolus chain configuration
RPC_TIMEOUT_SECS="10"
DEFAULT_GAS="1000000"
MAX_GAS="4000000"
GAS_PRICE_PRICE="0.0025"
GAS_MULTIPLIER="1.1"
MAX_MSG_NUM="30"
MAX_TX_SIZE="2097152"
CLOCK_DRIFT_SECS="5"
MAX_BLOCK_TIME_SECS="30"
TRUSTING_PERIOD_SECS="1209600"
TRUST_THRESHOLD_NUMERATOR="1"
TRUST_THRESHOLD_DENOMINATOR="3"

WASM_SCRIPT_PATH="$INIT_LOCAL_NETWORK_SCRIPT_DIR/../../nolus-money-market/scripts"
WASM_CODE_ARTIFACTS_PATH_PLATFORM="$INIT_LOCAL_NETWORK_SCRIPT_DIR/../../nolus-money-market/artifacts/platform"
ADMINS_BALANCE="10000000""$NATIVE_CURRENCY"

while [[ $# -gt 0 ]]; do
  key="$1"

  case $key in

  -h | --help)
    printf \
    "Usage: %s
    [--nolus-network-addr <nolus_node_listen_address>]
    [--nolus-network-rpc-port <nolus_network_rpc_port>]
    [--nolus-network-grpc-port <nolus_network_grpc_port>]
    [--native-currency <native_currency>]
    [-v|--validators <count>]
    [--validators-root-dir <validators_root_dir>]
    [--validator-accounts-dir <validator_accounts_dir>]
    [--validator-tokens <validators_initial_tokens>]
    [--user-dir <client_user_dir>]
    [--validator-stake <tokens_validator_stakes>]
    [--chain-id <chain_id>]
    [--treasury-nls-u128 <treasury_initial_nolus_tokens>]
    [--reserve-name <reserve_key_name>]
    [--reserve-tokens <initial_reserve_tokens>]
    [--gov-voting-period <voting_period>]
    [--feerefunder-ack-fee-min <feerefunder_ack_fee_min_amount>]
    [--feerefunder-timeout-fee-min <feerefunder_timeout_fee_min_amount>]
    [--dex-admin-mnemonic <dex_admin_mnemonic>]
    [--hermes-mnemonic <hermes_account_mnemonic>]
    [--hermes-version <hermes_version>]
    [--store-code-privileged-account-mnemonic <store_code_privileged_account_mnemonic>]
    [--rpc-timeout-secs <rpc_timeout_seconds>]
    [--default-gas <default_gas_amount>]
    [--max-gas <maximum_gas_amount>]
    [--gas-price-price <gas_price_price>]
    [--gas-multiplier <gas_multiplier>]
    [--max-msg-num <maximum_message_number>]
    [--max-tx-size <maximum_transaction_size>]
    [--clock-drift-secs <clock_drift_seconds>]
    [--max-block-time-secs <maximum_block_time_seconds>]
    [--trusting-period-secs <trusting_period_seconds>]
    [--trust-threshold-numerator <trust_threshold_numerator>]
    [--trust-threshold-denominator <trust_threshold_denominator>]
    [--wasm-script-path <wasm_script_path>]
    [--wasm-code-artifacts-path-platform <wasm_code_artifacts_path_platform>]
    [--admins-balance <admins_balance>]" \
    "$0"
    exit 0
    ;;

  --nolus-network-addr)
    NOLUS_NETWORK_ADDR="$2"
    shift 2
    ;;

  --nolus-network-rpc-port)
    NOLUS_NETWORK_RPC_PORT="$2"
    shift 2
    ;;

  --nolus-network-grpc-port)
    NOLUS_NETWORK_GRPC_PORT="$2"
    shift 2
    ;;

  --native-currency)
    NATIVE_CURRENCY="$2"
    shift 2
    ;;

  --validators)
    VALIDATORS="$2"
    shift 2
    ;;

  --validators-root-dir)
    VALIDATORS_ROOT_DIR="$2"
    shift 2
    ;;

  --validator-accounts-dir)
    VAL_ACCOUNTS_DIR="$2"
    shift 2
    ;;

  --user-dir)
    USER_DIR="$2"
    shift 2
    ;;

  --val-tokens)
    VAL_TOKENS="$2"
    shift 2
    ;;

  --val-stake)
    VAL_STAKE="$2"
    shift 2
    ;;

  --chain-id)
    CHAIN_ID="$2"
    shift 2
    ;;

  --treasury-nls-u128)
    TREASURY_NLS_U128="$2"
    shift 2
    ;;

  --reserve-name)
    RESERVE_NAME="$2"
    shift 2
    ;;

  --reserve-tokens)
    RESERVE_TOKENS="$2"
    shift 2
    ;;

  --gov-voting-period)
    GOV_VOTING_PERIOD="$2"
    shift 2
    ;;

  --feerefunder-ack-fee-min)
    FEEREFUNDER_ACK_FEE_MIN="$2"
    shift 2
    ;;

  --feerefunder-timeout-fee-min)
    FEEREFUNDER_TIMEOUT_FEE_MIN="$2"
    shift 2
    ;;

  --dex-admin-mnemonic)
    DEX_ADMIN_MNEMONIC="$2"
    shift 2
    ;;

  --hermes-mnemonic)
    HERMES_ACCOUNT_MNEMONIC="$2"
    shift 2
    ;;

  --hermes-version)
    HERMES_VERSION="$2"
    shift 2
    ;;

  --store-code-privileged-account-mnemonic)
    STORE_CODE_PRIVILEGED_ACCOUNT_MNEMONIC="$2"
    shift 2
    ;;

  --rpc-timeout-secs)
    RPC_TIMEOUT_SECS="$2"
    shift 2
    ;;

  --default-gas)
    DEFAULT_GAS="$2"
    shift 2
    ;;

  --max-gas)
    MAX_GAS="$2"
    shift 2
    ;;

  --gas-price-price)
    GAS_PRICE_PRICE="$2"
    shift 2
    ;;

  --gas-multiplier)
    GAS_MULTIPLIER="$2"
    shift 2
    ;;

  --max-msg-num)
    MAX_MSG_NUM="$2"
    shift 2
    ;;

  --max-tx-size)
    MAX_TX_SIZE="$2"
    shift 2
    ;;

  --clock-drift-secs)
    CLOCK_DRIFT_SECS="$2"
    shift 2
    ;;

  --max-block-time-secs)
    MAX_BLOCK_TIME_SECS="$2"
    shift 2
    ;;

  --trusting-period-secs)
    TRUSTING_PERIOD_SECS="$2"
    shift 2
    ;;

  --trust-threshold-numerator)
    TRUST_THRESHOLD_NUMERATOR="$2"
    shift 2
    ;;

  --trust-threshold-denominator)
    TRUST_THRESHOLD_DENOMINATOR="$2"
    shift 2
    ;;

  --wasm-script-path)
    WASM_SCRIPT_PATH="$2"
    shift 2
    ;;

  --wasm-code-artifacts-path-platform)
    WASM_CODE_ARTIFACTS_PATH_PLATFORM="$2"
    shift 2
    ;;

  --admins-balance)
    ADMINS_BALANCE="$2"
    shift 2
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
  run_cmd "$USER_DIR" config node "tcp://${NOLUS_NETWORK_ADDR}:$(first_node_rpc_port)"
}

verify_dir_exist "$WASM_SCRIPT_PATH" "WASM sripts path"
verify_dir_exist "$WASM_CODE_ARTIFACTS_PATH_PLATFORM" "WASM code path - platform"
verify_mandatory "$HERMES_ACCOUNT_MNEMONIC" "Hermes account mnemonic"
verify_mandatory "$HERMES_VERSION" "Hermes version (for example: 1.8.0)"
verify_mandatory "$DEX_ADMIN_MNEMONIC" "DEX-Admin account mnemonic"
verify_mandatory "$STORE_CODE_PRIVILEGED_ACCOUNT_MNEMONIC" "WASM store-code privileged account mnemonic"

rm_dir "$VALIDATORS_ROOT_DIR"
rm_dir "$VAL_ACCOUNTS_DIR"
rm_dir "$USER_DIR"

accounts_spec=$(echo "[]" | add_account "$(generate_account "$RESERVE_NAME" "$USER_DIR")" "$RESERVE_TOKENS")

source "$INIT_LOCAL_NETWORK_SCRIPT_DIR"/internal/setup-validator-local.sh
init_setup_validator_local_sh "$INIT_LOCAL_NETWORK_SCRIPT_DIR" "$VALIDATORS_ROOT_DIR"

source "$INIT_LOCAL_NETWORK_SCRIPT_DIR"/internal/init-network.sh
init_network "$VAL_ACCOUNTS_DIR" "$VALIDATORS" "$CHAIN_ID" "$NATIVE_CURRENCY" \
              "$VAL_TOKENS" "$VAL_STAKE" "$accounts_spec" "$WASM_SCRIPT_PATH" "$WASM_CODE_ARTIFACTS_PATH_PLATFORM" \
              "$TREASURY_NLS_U128" "$GOV_VOTING_PERIOD" "$FEEREFUNDER_ACK_FEE_MIN" "$FEEREFUNDER_TIMEOUT_FEE_MIN" \
              "$DEX_ADMIN_MNEMONIC" "$STORE_CODE_PRIVILEGED_ACCOUNT_MNEMONIC" "$ADMINS_BALANCE"

__config_client

run_cmd "$VALIDATORS_ROOT_DIR/local-validator-1" start &>"$USER_DIR"/nolus_logs.txt & disown;

/bin/bash "$INIT_LOCAL_NETWORK_SCRIPT_DIR"/remote/hermes-initial-config.sh "$HOME" "$CHAIN_ID" "$NOLUS_NETWORK_ADDR" \
                                                "$NOLUS_NETWORK_RPC_PORT" "$NOLUS_NETWORK_GRPC_PORT" "$RPC_TIMEOUT_SECS" \
                                                "$DEFAULT_GAS" "$MAX_GAS" "$GAS_PRICE_PRICE" \
                                                "$GAS_MULTIPLIER" "$MAX_MSG_NUM" "$MAX_TX_SIZE" \
                                                "$CLOCK_DRIFT_SECS" "$MAX_BLOCK_TIME_SECS" \
                                                "$TRUSTING_PERIOD_SECS" "$TRUST_THRESHOLD_NUMERATOR" \
                                                "$TRUST_THRESHOLD_DENOMINATOR" "$HERMES_ACCOUNT_MNEMONIC" "$HERMES_VERSION"

HERMES_BINARY_DIR="$HOME"/hermes

wait_nolus_gets_ready "$USER_DIR"
wait_hermes_config_gets_healthy "$HERMES_BINARY_DIR"
