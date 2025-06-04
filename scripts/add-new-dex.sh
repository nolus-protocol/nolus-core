#!/bin/bash
set -euxo pipefail

# Add new DEX.
# Extending the existing Hermes settings and creating a connection between Nolus and the new DEX.

SCRIPTS_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
source "$SCRIPTS_DIR"/internal/verify.sh
source "$SCRIPTS_DIR"/internal/add-dex-support.sh

NOLUS_NETWORK_ADDR="localhost"
NOLUS_NETWORK_RPC_PORT="26612"
NOLUS_HOME_DIR="$HOME/.nolus"
NOLUS_MONEY_MARKET_DIR="$SCRIPTS_DIR/../../nolus-money-market"
ACCOUNT_KEY_TO_FEED_HERMES_ADDRESS="reserve"
DEX_ADMIN_KEY="wasmAdmin"
STORE_CODE_PRIVILEGED_USER_KEY="storeAdmin"
LEASE_ADMIN_ADDRESS=""
WASM_ARTIFACTS_PATH=""

HERMES_CONFIG_DIR_PATH="$HOME/.hermes"
HERMES_BINARY_DIR_PATH="$HOME/hermes"
DEX_NETWORK=""
DEX_NAME=""
DEX_TYPE_AND_PARAMS=""
CHAIN_ID=""
CHAIN_IP_ADDR_RPC=""
CHAIN_IP_ADDR_GRPC=""
CHAIN_ACCOUNT_PREFIX=""
CHAIN_GAS_PRICE_DENOM=""
CHAIN_TRUSTING_PERIOD_SECS=""

CHAIN_RPC_TIMEOUT_SECS="10"
CHAIN_DEFAULT_GAS="5000000"
CHAIN_MAX_GAS="15000000"
CHAIN_GAS_PRICE_PRICE="0.056"
CHAIN_GAS_MULTIPLIER="2"
CHAIN_MAX_MSG_NUM="20"
CHAIN_MAX_TX_SIZE="209715"
CHAIN_CLOCK_DRIFT_SECS="20"
CHAIN_MAX_BLOCK_TIME_SECS="10"
CHAIN_TRUST_THRESHOLD_NUMERATOR="2"
CHAIN_TRUST_THRESHOLD_DENOMINATOR="3"
IF_INTERCHAIN_SECURITY="true"

PROTOCOL_CURRENCY=""
SWAP_TREE=""

while [[ $# -gt 0 ]]; do
  key="$1"

  case $key in

  -h | --help)
    printf \
    "Usage: %s
    [--nolus-home-dir <nolus_home_dir>]
    [--nolus-network-addr <nolus_node_listen_address>]
    [--nolus-network-rpc-port <nolus_node_rpc_port>]
    [--nolus-money-market-dir <nolus_money_market_dir>]
    [--account-key-to-feed-hermes-address <account_key_to_feed_hermes_address>]
    [--dex-admin-key <dex_admin_key>]
    [--store-code-privileged-user-key <store_code_privileged_user_key>]
    [--lease-admin-address <lease_admin_address>]
    [--wasm-artifacts-path <wasm_artifacts_path>]
    [--hermes-config-dir-path <config.toml_and_hermes.seed_dir_path>]
    [--hermes-binary-dir-path <hermes_binary_dir_path>]
    [--dex-network <dex_network>]
    [--dex-name <dex_name>]
    [--dex-type-and-params <dex_type_and_params>]
    [--dex-chain-id <new_dex_chain_id>]
    [--dex-ip-addr-rpc-host <new_dex_chain_ip_addr_rpc_fully_host>]
    [--dex-ip-addr-grpc-host <new_dex_chain_ip_addr_grpc_fully_host>]
    [--dex-account-prefix <new_dex_account_prefix>]
    [--dex-trusting-period-secs <new_dex_trusting_period_in_seconds>]
    [--dex-if-interchain-security <new_dex_if_interchain_security_true/false>]
    [--dex-rpc-timeout-secs <dex_rpc_timeout_in_seconds>]
    [--dex-default-gas <dex_default_gas>]
    [--dex-max-gas <dex_max_gas>]
    [--dex-gas-price-denom <dex_gas_price_denom>]
    [--dex-gas-multiplier <dex_gas_multiplier>]
    [--dex-max-msg-num <dex_max_msg_num>]
    [--dex-max-tx-size <dex_max_tx_size>]
    [--dex-clock-drift-secs <dex_clock_drift_in_seconds>]
    [--dex-max-block-time-secs <dex_max_block_time_in_seconds>]
    [--dex-trusting-period-secs <dex_trusting_period_in_seconds>]
    [--dex-trust-threshold-numerator <dex_trust_threshold_numerator>]
    [--dex-trust-threshold-denominator <dex_trust_threshold_denominator>]
    [--protocol-currency <new_protocol_currency>]
    [--admin-contract-address <admin_contract_address>]
    [--treasury-contract-address <treasury_contract_address>]
    [--timealarms-contract-address <timealarms_contract_address>]
    [--protocol-swap-tree <new_protocol_swap_tree>]" \
    "$0"
    exit 0
    ;;

  --nolus-home-dir)
    NOLUS_HOME_DIR="$2"
    shift 2
    ;;

  --nolus-network-addr)
    NOLUS_NETWORK_ADDR="$2"
    shift 2
    ;;

  --nolus-network-rpc-port)
    NOLUS_NETWORK_RPC_PORT="$2"
    shift 2
    ;;

  --nolus-money-market-dir)
    NOLUS_MONEY_MARKET_DIR="$2"
    shift 2
    ;;

  --account-key-to-feed-hermes-address)
    ACCOUNT_KEY_TO_FEED_HERMES_ADDRESS="$2"
    shift 2
    ;;

  --dex-admin-key)
    DEX_ADMIN_KEY="$2"
    shift 2
    ;;

  --store-code-privileged-user-key)
    STORE_CODE_PRIVILEGED_USER_KEY="$2"
    shift 2
    ;;

  --lease-admin-address)
    LEASE_ADMIN_ADDRESS="$2"
    shift 2
    ;;

  --wasm-artifacts-path)
    WASM_ARTIFACTS_PATH="$2"
    shift 2
    ;;

  --hermes-config-dir-path)
    HERMES_CONFIG_DIR_PATH="$2"
    shift 2
    ;;

  --hermes-binary-dir-path)
    HERMES_BINARY_DIR_PATH="$2"
    shift 2
    ;;

  --dex-network)
    DEX_NETWORK="$2"
    shift 2
    ;;

  --dex-name)
    DEX_NAME="$2"
    shift 2
    ;;

  --dex-type-and-params)
    DEX_TYPE_AND_PARAMS="$2"
    shift 2
    ;;

  --dex-chain-id)
    CHAIN_ID="$2"
    shift 2
    ;;

  --dex-ip-addr-rpc-host)
    CHAIN_IP_ADDR_RPC="$2"
    shift 2
    ;;

  --dex-ip-addr-grpc-host)
    CHAIN_IP_ADDR_GRPC="$2"
    shift 2
    ;;

  --dex-account-prefix)
    CHAIN_ACCOUNT_PREFIX="$2"
    shift 2
    ;;

  --dex-gas-price-denom)
    CHAIN_GAS_PRICE_DENOM="$2"
    shift 2
    ;;

  --dex-trusting-period-secs)
    CHAIN_TRUSTING_PERIOD_SECS="$2"
    shift 2
    ;;

  --dex-rpc-timeout-secs)
    CHAIN_RPC_TIMEOUT_SECS="$2"
    shift 2
      ;;

  --dex-default-gas)
    CHAIN_DEFAULT_GAS="$2"
    shift 2
    ;;

  --dex-max-gas)
    CHAIN_MAX_GAS="$2"
    shift 2
    ;;

  --dex-gas-multiplier)
    CHAIN_GAS_MULTIPLIER="$2"
    shift 2
    ;;

  --dex-max-msg-num)
    CHAIN_MAX_MSG_NUM="$2"
    shift 2
    ;;

  --dex-max-tx-size)
    CHAIN_MAX_TX_SIZE="$2"
    shift 2
    ;;

  --dex-clock-drift-secs)
    CHAIN_CLOCK_DRIFT_SECS="$2"
    shift 2
    ;;

  --dex-max-block-time-secs)
    CHAIN_MAX_BLOCK_TIME_SECS="$2"
    shift 2
    ;;

  --dex-trust-threshold-numerator)
    CHAIN_TRUST_THRESHOLD_NUMERATOR="$2"
    shift 2
    ;;

  --dex-trust-threshold-denominator)
    CHAIN_TRUST_THRESHOLD_DENOMINATOR="$2"
    shift 2
    ;;

  --dex-if-interchain-security)
    IF_INTERCHAIN_SECURITY="$2"
    shift 2
    ;;

  --protocol-currency)
    PROTOCOL_CURRENCY="$2"
    shift 2
    ;;

  --admin-contract-address)
    ADMIN_CONTRACT_ADDRESS="$2"
    shift 2
    ;;

  --treasury-contract-address)
    TREASURY_CONTRACT_ADDRESS="$2"
    shift 2
    ;;

  --timealarms-contract-address)
    TIMEALARMS_CONTRACT_ADDRESS="$2"
    shift 2
    ;;

  --protocol-swap-tree)
    SWAP_TREE="$2"
    shift 2
    ;;

  *)
    echo >&2 "The provided option '$key' is not recognized"
    exit 1
    ;;

  esac
done

source "$NOLUS_MONEY_MARKET_DIR/scripts/deploy-platform.sh"

ADMIN_CONTRACT_ADDRESS="$(admin_contract_instance_addr)"
TREASURY_CONTRACT_ADDRESS="$(treasury_instance_addr)"
TIMEALARMS_CONTRACT_ADDRESS="$(timealarms_instance_addr)"

NOLUS_CHAIN_ID=$(grep -oP 'chain-id = "\K[^"]+' "$NOLUS_HOME_DIR"/config/client.toml)

verify_dir_exist "$NOLUS_MONEY_MARKET_DIR" "The NOLUS_MONEY_MARKET dir does not exist"
DEPLOY_CONTRACTS_SCRIPT="$NOLUS_MONEY_MARKET_DIR/scripts/deploy-protocol.sh"

verify_dir_exist "$WASM_ARTIFACTS_PATH" "The WASM_ARTIFACTS_PATH dir does not exist"
verify_mandatory "$LEASE_ADMIN_ADDRESS" "lease admin address"
verify_mandatory "$DEX_NAME" "new DEX name"
verify_mandatory "$DEX_TYPE_AND_PARAMS" "DEX type and parameters"
verify_mandatory "$DEX_NETWORK" "new DEX network"
verify_mandatory "$CHAIN_ID" "new DEX chain_id"
verify_mandatory "$CHAIN_IP_ADDR_RPC" "new DEX RPC addr - fully host part"
verify_mandatory "$CHAIN_IP_ADDR_GRPC" "new DEX gRPC addr - fully host part"
verify_mandatory "$CHAIN_ACCOUNT_PREFIX" "new DEX account prefix"
verify_mandatory "$CHAIN_GAS_PRICE_DENOM" "new DEX gas price denom"
verify_mandatory "$CHAIN_TRUSTING_PERIOD_SECS" "new DEX trusting period"
verify_mandatory "$PROTOCOL_CURRENCY" "new protocol lpn"
verify_mandatory "$SWAP_TREE" "new protocol swap_tree"

verify_file_exist "$DEPLOY_CONTRACTS_SCRIPT" "The script does not exist"
source "$DEPLOY_CONTRACTS_SCRIPT"

 if [[ $IF_INTERCHAIN_SECURITY != "true" && $IF_INTERCHAIN_SECURITY != "false" ]]; then
    echo >&2 "the dex-if-interchain-security value must be true or false"
    exit 1
  fi

# Extend the existing Hermes configuration
add_new_chain_hermes "$HERMES_CONFIG_DIR_PATH" "$CHAIN_ID" "$CHAIN_IP_ADDR_RPC" "$CHAIN_IP_ADDR_GRPC" \
   "$CHAIN_RPC_TIMEOUT_SECS"  "$CHAIN_ACCOUNT_PREFIX" "$CHAIN_DEFAULT_GAS" "$CHAIN_MAX_GAS" "$CHAIN_GAS_PRICE_PRICE" \
    "$CHAIN_GAS_PRICE_DENOM" "$CHAIN_GAS_MULTIPLIER" "$CHAIN_MAX_MSG_NUM" "$CHAIN_MAX_TX_SIZE" \
    "$CHAIN_CLOCK_DRIFT_SECS" "$CHAIN_MAX_BLOCK_TIME_SECS" "$CHAIN_TRUSTING_PERIOD_SECS" \
    "$CHAIN_TRUST_THRESHOLD_NUMERATOR" "$CHAIN_TRUST_THRESHOLD_DENOMINATOR" "$IF_INTERCHAIN_SECURITY"

# Link the Hermes account to the DEX
dex_account_setup "$HERMES_BINARY_DIR_PATH" "$CHAIN_ID" "$HERMES_CONFIG_DIR_PATH"/hermes.seed

NOLUS_HERMES_ADDRESS=$(get_hermes_address "$HERMES_BINARY_DIR_PATH" "$NOLUS_CHAIN_ID")

NOLUS_NET="http://${NOLUS_NETWORK_ADDR}:${NOLUS_NETWORK_RPC_PORT}/"

# Open a connection (exports CONNECTION_ID)
open_connection "$NOLUS_NET" "$NOLUS_HOME_DIR" "$ACCOUNT_KEY_TO_FEED_HERMES_ADDRESS" "$HERMES_BINARY_DIR_PATH" \
    "$NOLUS_HERMES_ADDRESS" "$NOLUS_CHAIN_ID" "$CHAIN_ID"

DEX_CONNECTION_ID="$CONNECTION_ID"

CONNECTION_INFO=$(get_connection_info "$NOLUS_HOME_DIR" "$DEX_CONNECTION_ID")
DEX_CHANNEL_LOCAL=$(echo "$CONNECTION_INFO" |  jq -r '.channels[0].channel_id')
DEX_CHANNEL_REMOTE=$(echo "$CONNECTION_INFO" | jq -r '.channels[0].counterparty.channel_id')

# TO DO - Remove and run manually
# Deploy contracts
_=$(
  deploy_contracts "$NOLUS_NET" "$NOLUS_CHAIN_ID" "$NOLUS_HOME_DIR" \
    "$DEX_ADMIN_KEY" "$STORE_CODE_PRIVILEGED_USER_KEY" "$LEASE_ADMIN_ADDRESS" \
    "$ADMIN_CONTRACT_ADDRESS" "$WASM_ARTIFACTS_PATH/$DEX_NAME" "$DEX_NETWORK" \
    "$DEX_NAME" "$DEX_TYPE_AND_PARAMS" "$DEX_CONNECTION_ID" \
    "$DEX_CHANNEL_LOCAL" "$DEX_CHANNEL_REMOTE" "$PROTOCOL_CURRENCY" \
    "$TREASURY_CONTRACT_ADDRESS" \
    "$TIMEALARMS_CONTRACT_ADDRESS" "$SWAP_TREE"
)
