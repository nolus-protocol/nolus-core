#!/bin/bash
# DEX setup

set -euox pipefail

SCRIPTS_DIR=$(cd $(dirname "${BASH_SOURCE[0]}") && pwd)
source "$SCRIPTS_DIR"/common/cmd.sh
source "$SCRIPTS_DIR"/internal/verify.sh

key_recover() {
  local -r nolus_home_dir="$1"
  local -r mnemonic="$1"
  local -r key="$2"

  echo "$mnemonic" | run_cmd "$nolus_home_dir" keys add "$key" --recover --keyring-backend "test"
}

NOLUS_NET_ADDRESS=""
NOLUS_HOME_DIR=""
CONTRACTS_OWNER_KEY=""
WALLET_WITH_FUNDS_KEY=""
CONTRACTS_INFO_FILE_PATH=""
CONTRACTS_OWNER_MNEMONIC=""
FAUCET_MNEMONIC=""
HERMES_BINARY_DIR=""
HERMES_MNEMONIC=""
A_CHAIN=""
B_CHAIN=""

while [[ $# -gt 0 ]]; do
  key="$1"

  case $key in

  -h | --help)
    printf \
    "Usage: %s
    [--nolus-net-address <nolus_net_address>]
    [--nolus-home-dir <nolus_home_dir_path>]
    [--contracts-owner-key <contracts_owner_key (if exists)>]
    [--wallet-with-funds-key <wallet_with_funds_key (if exists)>]
    [--contracts-info-file-path <contracts_info_file_full_path>]
    [--contracts-owner-mnemonic <contracts_owner_mnemonic_to_be_recovered  (if exists)>]
    [--faucet-mnemonic <faucet_mnemonic_to_be_recovered (if exists)>]
    [--hermes-binary-dir <hermes_binary_dir_path>]
    [--hermes-mnemonic <hermes_account_mnemonic>]
    [--a-chain-id <configured_a_chain_id>]
    [--b-chain-id <configured_a_chain_id>]" \
    "$0"
    exit 0
    ;;

  --nolus-net-address)
    NOLUS_NET_ADDRESS="$2"
    shift
    shift
    ;;

  --nolus-home-dir)
    NOLUS_HOME_DIR="$2"
    shift
    shift
    ;;

  --contracts-owner-key)
    CONTRACTS_OWNER_KEY="$2"
    shift
    shift
    ;;

  --wallet-with-funds-key)
    WALLET_WITH_FUNDS_KEY="$2"
    shift
    shift
    ;;

  --contracts-info-file-path)
    CONTRACTS_INFO_FILE_PATH="$2"
    shift
    shift
    ;;

  --contracts-owner-mnemonic)
    CONTRACTS_OWNER_MNEMONIC="$2"
    shift
    shift
    ;;

  --faucet-mnemonic)
    FAUCET_MNEMONIC="$2"
    shift
    shift
    ;;

  --hermes-binary-dir)
    HERMES_BINARY_DIR="$2"
    shift
    shift
    ;;

  --hermes-mnemonic)
    HERMES_MNEMONIC="$2"
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

verify_mandatory "$NOLUS_NET_ADDRESS" "Nolus address"
verify_mandatory "$NOLUS_HOME_DIR" "Nolus home directory path"
verify_mandatory "$CONTRACTS_INFO_FILE_PATH" "Smart Contracts information file path"
verify_mandatory "$HERMES_BINARY_DIR" "Hermes binary directory path"
verify_mandatory "$HERMES_MNEMONIC" "Hermes account mnemonic"
verify_mandatory "$A_CHAIN" "Configured A chain id in Hermes config"
verify_mandatory "$B_CHAIN" "Configured B chain id in Hermes config"

if [ -z "$CONTRACTS_OWNER_MNEMONIC" ]; then
    verify_mandatory "$CONTRACTS_OWNER_KEY" "Smart Contracts owner key"
fi

if [ -z "$FAUCET_MNEMONIC" ]; then
    verify_mandatory "$WALLET_WITH_FUNDS_KEY" "Active key, with funds"
fi

if [ -n "$CONTRACTS_OWNER_MNEMONIC" ]; then
    CONTRACTS_OWNER_KEY="contracts_owner"
    key_recover "$NOLUS_HOME_DIR" "$CONTRACTS_OWNER_MNEMONIC" "$CONTRACTS_OWNER_KEY"
fi

if [ -n "$FAUCET_MNEMONIC" ]; then
    WALLET_WITH_FUNDS_KEY="faucet"
    key_recover "$NOLUS_HOME_DIR" "$FAUCET_MNEMONIC" "$WALLET_WITH_FUNDS_KEY"
fi


# Prepare Hermes

source "$SCRIPTS_DIR"/internal/hermes-setup.sh

setup_accounts "$NOLUS_HOME_DIR" "$NOLUS_NET_ADDRESS" "$WALLET_WITH_FUNDS_KEY" "$HERMES_BINARY_DIR" "$A_CHAIN" "$B_CHAIN" "$HERMES_MNEMONIC"

CONNECTION="connection-0"
COUNTERPARTY_CHANNEL_ID=$(open_connection "$NOLUS_HOME_DIR" "$NOLUS_NET_ADDRESS" "$HERMES_BINARY_DIR" "$A_CHAIN" "$B_CHAIN" "$CONNECTION")

# Setup Leaser

CONTRACTS_OWNER_ADDRESS=$(run_cmd "$NOLUS_HOME_DIR" keys show "$CONTRACTS_OWNER_KEY" -a)
FLAGS="--fees 1000unls --gas auto --gas-adjustment 1.3 --node $NOLUS_NET_ADDRESS"
echo 'y' | run_cmd "$NOLUS_HOME_DIR" tx bank send "$WALLET_WITH_FUNDS_KEY" "$CONTRACTS_OWNER_ADDRESS" 10000unls --broadcast-mode block $FLAGS

LEASER_CONTRACT_ADDRESS=$(jq .contracts_info[5].leaser.instance "$CONTRACTS_INFO_FILE_PATH" | tr -d '"')
SETUP_DEX_MSG='{"setup_dex":{"connection_id":"'$CONNECTION'","transfer_channel":{"local_endpoint":"channel-0","remote_endpoint":"'$COUNTERPARTY_CHANNEL_ID'"}}}'
echo 'y' | run_cmd "$NOLUS_HOME_DIR" tx wasm execute "$LEASER_CONTRACT_ADDRESS" "$SETUP_DEX_MSG" --from "$CONTRACTS_OWNER_KEY" $FLAGS