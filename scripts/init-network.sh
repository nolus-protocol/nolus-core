#!/bin/bash
set -euxo pipefail

SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
source "$SCRIPT_DIR"/common/cmd.sh
source "$SCRIPT_DIR"/common/rm-dir.sh
source "$SCRIPT_DIR"/internal/accounts.sh
source "$SCRIPT_DIR"/internal/verify.sh

cleanup() {
  cleanup_init_network_sh
  exit
}
trap cleanup INT TERM EXIT

determine_faucet_addr() {
  local -r faucet_mnemonic="$1"
  local -r faucet_dir="$(mktemp -d)"
  local -r key_name="anonymous"

  recover_account "$faucet_dir" "$faucet_mnemonic" "$key_name"

  rm_dir "$faucet_dir"
}

MONIKER_BASE="validator"
VALIDATORS=1
MINIMUM_GAS_PRICE="0.0025unls"
QUERY_GAS_LIMIT="3500000"
VAL_ACCOUNTS_DIR="networks/nolus/val-accounts"
ARTIFACT_BIN=""
ARTIFACT_SCRIPTS=""
SSH_USER=""
SSH_IP=""
SSH_KEY=""
NATIVE_CURRENCY="unls"
VAL_TOKENS="1000000000""$NATIVE_CURRENCY"
VAL_STAKE="1000000""$NATIVE_CURRENCY"
CHAIN_ID=""
WASM_SCRIPT_PATH=""
TREASURY_NLS_U128="1000000000000"
FAUCET_MNEMONIC=""
FAUCET_TOKENS="100000000000""$NATIVE_CURRENCY"
GOV_VOTING_PERIOD="3600s"
GOV_MAX_DEPOSIT_PERIOD="43200s"
GOV_EXPEDITED_VOTING_PERIOD="2m"
STAKING_MAX_VALIDATORS="40"
FEEREFUNDER_ACK_FEE_MIN="1"
FEEREFUNDER_TIMEOUT_FEE_MIN="1"
DEX_ADMIN_MNEMONIC=""
ADMINS_TOKENS="10000000""$NATIVE_CURRENCY"
STORE_CODE_PRIVILEGED_ACCOUNT_MNEMONIC=""

while [[ $# -gt 0 ]]; do
  key="$1"

  case $key in

  -h | --help)
    printf \
    "Usage: %s
    [--artifact-bin <tar_gz_nolusd>]
    [--artifact-scripts <tar_gz_scripts>]
    [--ip <string - ip of the remote host>]
    [--user <string - ssh key user>]
    [--ssh-key <string - ssh pvt key file path>]
    [--chain-id <string>]
    [--validators <number>]
    [--minimum-gas-price <minimum_gas_price - X.XXunls>]
    [--query-gas-limit <query_gas_limit>]
    [--validator-accounts-dir <validator_accounts_dir>]
    [--validator-tokens <tokens_for_val_genesis_accounts>]
    [--validator-stake <tokens_val_will_stake>]
    [--wasm-script-path <wasm_script_path>]
    [--wasm-code-path <wasm_code_path>]
    [--treasury-nls-u128 <treasury_initial_Nolus_tokens>]
    [--faucet-mnemonic <mnemonic_phrase>]
    [--faucet-tokens <initial_balance>]
    [--gov-voting-period <voting_period>]
    [--gov-max-deposit-period <max_deposit_period - XXs>]
    [--gov-expedited-voting-period <expedited_voting_period>]
    [--staking-max-validators <staking_max_validators>]
    [--feerefunder-ack-fee-min <feerefunder_ack_fee_min_amount>]
    [--feerefunder-timeout-fee-min <feerefunder_timeout_fee_min_amount>]
    [--moniker <string - node moniker (default: $MONIKER_BASE>]
    [--dex-admin-mnemonic <dex_admin_mnemonic>]
    [--store-code-privileged-account-mnemonic <store_code_privileged_account_mnemonic>]" \
      "$0"
    exit 0
    ;;

  --artifact-bin)
    ARTIFACT_BIN="$2"
    shift
    shift
    ;;

  --artifact-scripts)
    ARTIFACT_SCRIPTS="$2"
    shift
    shift
    ;;

  --ip)
    SSH_IP=$2
    shift
    shift
    ;;

  --user)
    SSH_USER=$2
    shift
    shift
    ;;

  --ssh-key)
    SSH_KEY=$2
    shift
    shift
    ;;

  --chain-id)
    CHAIN_ID="$2"
    shift
    shift
    ;;

  --validators)
    VALIDATORS="$2"
    [ "$VALIDATORS" -gt 0 ] || {
      echo >&2 "validators must be a positive number"
      exit 1
    }
    shift
    shift
    ;;

  --minimum-gas-price)
    MINIMUM_GAS_PRICE="$2"
    shift
    shift
    ;;

  --query-gas-limit)
    QUERY_GAS_LIMIT="$2"
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

  --treasury-nls-u128)
    TREASURY_NLS_U128="$2"
    shift
    shift
    ;;

  --faucet-mnemonic)
    FAUCET_MNEMONIC="$2"
    shift
    shift
    ;;

  --faucet-tokens)
    FAUCET_TOKENS="$2"
    shift
    shift
    ;;

  --gov-voting-period)
    GOV_VOTING_PERIOD="$2"
    shift
    shift
    ;;

  --gov-max-deposit-period)
    GOV_MAX_DEPOSIT_PERIOD="$2"
    shift
    shift
    ;;

  --gov-expedited-voting-period)
    GOV_EXPEDITED_VOTING_PERIOD="$2"
    shift
    shift
    ;;

  --staking-max-validators)
    STAKING_MAX_VALIDATORS="$2"
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

  --moniker)
    MONIKER_BASE="$2"
    shift
    shift
    ;;

  --dex-admin-mnemonic)
    DEX_ADMIN_MNEMONIC="$2"
    shift
    shift
    ;;

  --store-code-privileged-account-mnemonic)
    STORE_CODE_PRIVILEGED_ACCOUNT_MNEMONIC="$2"
    shift
    shift
    ;;

  *)
    echo >&2 "The provided option '$key' is not recognized"
    exit 1
    ;;

  esac
done

verify_mandatory "$ARTIFACT_BIN" "Nolus binary actifact"
verify_mandatory "$ARTIFACT_SCRIPTS" "Nolus scripts actifact"
verify_mandatory "$WASM_SCRIPT_PATH" "Wasm script path"
verify_mandatory "$FAUCET_MNEMONIC" "Faucet mnemonic"
verify_mandatory "$CHAIN_ID" "Nolus Chain ID"
verify_mandatory "$SSH_USER" "Server ssh user"
verify_mandatory "$SSH_IP" "Server ip"
verify_mandatory "$SSH_KEY" "SSH pvt key file path"
verify_mandatory "$DEX_ADMIN_MNEMONIC" "DEX-Admin account mnemonic"
verify_mandatory "$STORE_CODE_PRIVILEGED_ACCOUNT_MNEMONIC" "WASM store-code privileged account mnemonic"

rm_dir "$VAL_ACCOUNTS_DIR"

FAUCET_ADDR=$(determine_faucet_addr "$FAUCET_MNEMONIC")
accounts_spec=$(echo "[]" | add_account "$FAUCET_ADDR" "$FAUCET_TOKENS")

source "$SCRIPT_DIR"/internal/setup-validator.sh

init_setup_validator "$SCRIPT_DIR" "$ARTIFACT_BIN" "$ARTIFACT_SCRIPTS" "$MONIKER_BASE" "$SSH_USER" "$SSH_IP" "$SSH_KEY"
deploy_binary
deploy_scripts
setup_services "$VALIDATORS"

source "$SCRIPT_DIR"/internal/init-network.sh
init_network "$VAL_ACCOUNTS_DIR" "$VALIDATORS" "$MINIMUM_GAS_PRICE" "$QUERY_GAS_LIMIT" "$CHAIN_ID" "$NATIVE_CURRENCY" \
  "$VAL_TOKENS" "$VAL_STAKE" "$accounts_spec" "$WASM_SCRIPT_PATH" "$TREASURY_NLS_U128" \
  "$GOV_VOTING_PERIOD" "$GOV_MAX_DEPOSIT_PERIOD" "$GOV_EXPEDITED_VOTING_PERIOD" "$STAKING_MAX_VALIDATORS" \
  "$FEEREFUNDER_ACK_FEE_MIN" "$FEEREFUNDER_TIMEOUT_FEE_MIN" \
  "$DEX_ADMIN_MNEMONIC" "$STORE_CODE_PRIVILEGED_ACCOUNT_MNEMONIC" "$ADMINS_TOKENS"

start_validators "$VALIDATORS"
