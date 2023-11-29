#!/bin/bash
set -euxo pipefail

# Add new DEX.
# Extending the existing Hermes settings and creating a connection between Nolus and the new DEX.

SCRIPTS_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
source "$SCRIPTS_DIR"/internal/verify.sh
source "$SCRIPTS_DIR"/internal/add-dex-support.sh

NOLUS_NET="http://localhost:26612/"
NOLUS_HOME_DIR="$HOME/.nolus"
NOLUS_MONEY_MARKET_DIR="$SCRIPTS_DIR/../../nolus-money-market"
ACCOUNT_KEY_TO_FEED_HERMES_ADDRESS="reserve"
DEX_ADMIN_KEY=""
STORE_CODE_PRIVILEGED_USER_KEY=""
WASM_ARTIFACTS_PATH=""

HERMES_CONFIG_DIR_PATH="$HOME/.hermes"
HERMES_BINARY_DIR_PATH="$HOME/hermes"
DEX_NAME=""
CHAIN_ID=""
CHAIN_IP_ADDR_RPC=""
CHAIN_IP_ADDR_GRPC=""
CHAIN_ACCOUNT_PREFIX=""
CHAIN_PRICE_DENOM=""
CHAIN_TRUSTING_PERIOD=""
IF_INTERCHAIN_SECURITY="true"

PROTOCOL_CURRENCY=""
ADMIN_CONTRACT_ADDRESS="nolus1gurgpv8savnfw66lckwzn4zk7fp394lpe667dhu7aw48u40lj6jsqxf8nd"
TRASURY_CONTRACT_ADDRESS="nolus14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9s0k0puz"
TIMEALARMS_CONTRACT_ADDRESS="nolus1zwv6feuzhy6a9wekh96cd57lsarmqlwxdypdsplw6zhfncqw6ftqmx7chl"
SWAP_TREE=""

while [[ $# -gt 0 ]]; do
  key="$1"

  case $key in

  -h | --help)
    printf \
    "Usage: %s
    [--nolus-home-dir <nolus_home_dir>]
    [--nolus-money-market-dir <nolus_money_market_dir>]
    [--account-key-to-feed-hermes-address <account_key_to_feed_hermes_address>]
    [--dex-admin-key <dex_admin_key>]
    [--store-code-privileged-user-key <store_code_privileged_user_key>]
    [--wasm-artifacts-path <wasm_artifacts_path>]
    [--hermes-config-dir-path <config.toml_and_hermes.seed_dir_path>]
    [--hermes-binary-dir-path <hermes_binary_dir_path>]
    [--dex-name <dex_name>]
    [--dex-chain-id <new_dex_chain_id>]
    [--dex-ip-addr-rpc-host <new_dex_chain_ip_addr_rpc_fully_host>]
    [--dex-ip-addr-grpc-host <new_dex_chain_ip_addr_grpc_fully_host>]
    [--dex-account-prefix <new_dex_account_prefix>]
    [--dex-price-denom <new_dex_price_denom>]
    [--dex-trusting-period-secs <new_dex_trusting_period_in_seconds>]
    [--dex-if-interchain-security <new_dex_if_interchain_security_true/false>]
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
    shift
    shift
    ;;

  --nolus-money-market-dir)
    NOLUS_MONEY_MARKET_DIR="$2"
    shift
    shift
    ;;

  --account-key-to-feed-hermes-address)
    ACCOUNT_KEY_TO_FEED_HERMES_ADDRESS="$2"
    shift
    shift
    ;;

  --dex-admin-key)
    DEX_ADMIN_KEY="$2"
    shift
    shift
    ;;

  --store-code-privileged-user-key)
    STORE_CODE_PRIVILEGED_USER_KEY="$2"
    shift
    shift
    ;;

  --wasm-artifacts-path)
    WASM_ARTIFACTS_PATH="$2"
    shift
    shift
    ;;

  --hermes-config-dir-path)
    HERMES_CONFIG_DIR_PATH="$2"
    shift
    shift
    ;;

  --hermes-binary-dir-path)
    HERMES_BINARY_DIR_PATH="$2"
    shift
    shift
    ;;

  --dex-name)
    DEX_NAME="$2"
    shift
    shift
    ;;

  --dex-chain-id)
    CHAIN_ID="$2"
    shift
    shift
    ;;

  --dex-ip-addr-rpc-host)
    CHAIN_IP_ADDR_RPC="$2"
    shift
    shift
    ;;

  --dex-ip-addr-grpc-host)
    CHAIN_IP_ADDR_GRPC="$2"
    shift
    shift
    ;;

  --dex-account-prefix)
    CHAIN_ACCOUNT_PREFIX="$2"
    shift
    shift
    ;;

  --dex-price-denom)
    CHAIN_PRICE_DENOM="$2"
    shift
    shift
    ;;

  --dex-trusting-period-secs)
    CHAIN_TRUSTING_PERIOD="$2"
    shift
    shift
    ;;

  --dex-if-interchain-security)
    IF_INTERCHAIN_SECURITY="$2"
    shift
    shift
    ;;

  --protocol-currency)
    PROTOCOL_CURRENCY="$2"
    shift
    shift
    ;;

  --admin-contract-address)
    ADMIN_CONTRACT_ADDRESS="$2"
    shift
    shift
    ;;

  --treasury-contract-address)
    TRASURY_CONTRACT_ADDRESS="$2"
    shift
    shift
    ;;

  --timealarms-contract-address)
    TIMEALARMS_CONTRACT_ADDRESS="$2"
    shift
    shift
    ;;

  --protocol-swap-tree)
    SWAP_TREE="$2"
    shift
    shift
    ;;

  *)
    echo >&2 "The provided option '$key' is not recognized"
    exit 1
    ;;

  esac
done

NOLUS_CHAIN_ID=$(grep -oP 'chain-id = "\K[^"]+' "$NOLUS_HOME_DIR"/config/client.toml)

verify_dir_exist "$NOLUS_MONEY_MARKET_DIR" "The NOLUS_MONEY_MARKET_DIR dir does not exist"
DEPLOY_CONTRACTS_SCRIPT="$NOLUS_MONEY_MARKET_DIR/scripts/deploy-contracts-live.sh"

verify_dir_exist "$WASM_ARTIFACTS_PATH" "The WASM_ARTIFACTS_PATH dir does not exist"
verify_mandatory "$DEX_NAME" "new DEX name"
verify_mandatory "$DEX_ADMIN_KEY" "dex-admin key name"
verify_mandatory "$STORE_CODE_PRIVILEGED_USER_KEY" "sotre-code privileged user key"
verify_mandatory "$CHAIN_ID" "new DEX chain_id"
verify_mandatory "$CHAIN_IP_ADDR_RPC" "new DEX RPC addr - fully host part"
verify_mandatory "$CHAIN_IP_ADDR_GRPC" "new DEX gRPC addr - fully host part"
verify_mandatory "$CHAIN_ACCOUNT_PREFIX" "new DEX account prefix"
verify_mandatory "$CHAIN_PRICE_DENOM" "new DEX price  denom"
verify_mandatory "$CHAIN_TRUSTING_PERIOD" "new DEX trusting period"
verify_mandatory "$PROTOCOL_CURRENCY" "new protocol currency"
verify_mandatory "$SWAP_TREE" "new protocol swap_tree"

verify_file_exist "$DEPLOY_CONTRACTS_SCRIPT" "The script does not exist"
source "$DEPLOY_CONTRACTS_SCRIPT"

 if [[ $IF_INTERCHAIN_SECURITY != "true" && $IF_INTERCHAIN_SECURITY != "false" ]]; then
    echo >&2 "the dex-if-interchain-security value must be true or false"
    exit 1
  fi

# Extend the existing Hermes configuration
add_new_chain_hermes "$HERMES_CONFIG_DIR_PATH" "$CHAIN_ID" "$CHAIN_IP_ADDR_RPC" "$CHAIN_IP_ADDR_GRPC" \
    "$CHAIN_ACCOUNT_PREFIX" "$CHAIN_PRICE_DENOM" "$CHAIN_TRUSTING_PERIOD" "$IF_INTERCHAIN_SECURITY"

# Link the Hermes account to the DEX
dex_account_setup "$HERMES_BINARY_DIR_PATH" "$CHAIN_ID" "$HERMES_CONFIG_DIR_PATH"/hermes.seed

NOLUS_HERMES_ADDRESS=$(get_hermes_address "$HERMES_BINARY_DIR_PATH" "$NOLUS_CHAIN_ID")

# Open a connection
open_connection "$NOLUS_NET" "$NOLUS_HOME_DIR" "$ACCOUNT_KEY_TO_FEED_HERMES_ADDRESS" "$HERMES_BINARY_DIR_PATH" \
    "$NOLUS_HERMES_ADDRESS" "$NOLUS_CHAIN_ID" "$CHAIN_ID"

# Deploy contracts
_=$(deploy_contracts "$NOLUS_NET" "$NOLUS_CHAIN_ID" "$NOLUS_HOME_DIR" "$DEX_ADMIN_KEY" "$STORE_CODE_PRIVILEGED_USER_KEY" \
"$ADMIN_CONTRACT_ADDRESS" "$WASM_ARTIFACTS_PATH/$DEX_NAME" "$DEX_NAME" "$PROTOCOL_CURRENCY" \
"$TRASURY_CONTRACT_ADDRESS" "$TIMEALARMS_CONTRACT_ADDRESS" "$SWAP_TREE")
