#!/bin/bash
set -euxo pipefail

SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
source "$SCRIPT_DIR"/common/cmd.sh
source "$SCRIPT_DIR"/internal/accounts.sh

cleanup() {
  cleanup_init_network_sh
  exit
}
trap cleanup INT TERM EXIT

VALIDATORS=1
VAL_ACCOUNTS_DIR="networks/nolus/val-accounts"
ARTIFACT_BIN=""
ARTIFACT_SCRIPTS=""

NATIVE_CURRENCY="unolus"
VAL_TOKENS="1000000000""$NATIVE_CURRENCY"
VAL_STAKE="1000000""$NATIVE_CURRENCY"
CHAIN_ID="nolus-dev"
WASM_CODE_PATH=""
TREASURY_NLS_U128="1000000000000"
FAUCET_MNEMONIC=""
FAUCET_TOKENS="1000000""$NATIVE_CURRENCY"

while [[ $# -gt 0 ]]; do
  key="$1"

  case $key in

  -h | --help)
    printf \
    "Usage: %s
    [--artifact-bin <tar_gz_nolusd>]
    [--artifact-scripts <tar_gz_scripts>]
    [--chain-id <string>]
    [-v|--validators <number>]
    [--validator_accounts_dir <validator_accounts_dir>]
    [--validator-tokens <tokens_for_val_genesis_accounts>]
    [--validator-stake <tokens_val_will_stake>]
    [--wasm-code-path <wasm_code_path>]
    [--treasury-nls-u128 <treasury_initial_Nolus_tokens>]
    [--faucet-mnemonic <mnemonic_phrase>]
    [--faucet-tokens <initial_balance>]" \
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

   --validator_accounts_dir)
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
  *)
    echo >&2 "The provided option '$key' is not recognized"
    exit 1    
    ;;

  esac
done

__verify_mandatory() {
  local value="$1"
  local description="$2"

  if [[ -z "$value" ]]; then
    echo >&2 "$description was not set"
    exit 1
  fi
}

__verify_mandatory "$ARTIFACT_BIN" "Nolus binary actifact"
__verify_mandatory "$ARTIFACT_SCRIPTS" "Nolus scipts actifact"
__verify_mandatory "$WASM_CODE_PATH" "Wasm code path"
__verify_mandatory "$FAUCET_MNEMONIC" "Faucet mnemonic"

rm -fr "$VAL_ACCOUNTS_DIR"

accounts_spec=$(echo "[]" | add_account "$(recover_account "$FAUCET_MNEMONIC")" "$FAUCET_TOKENS")

source "$SCRIPT_DIR"/internal/setup-validator-dev.sh
init_setup_validator_dev_sh "$SCRIPT_DIR" "$ARTIFACT_BIN" "$ARTIFACT_SCRIPTS"
stop_validators "$VALIDATORS"
deploy_validators "$VALIDATORS"

source "$SCRIPT_DIR"/internal/init-network.sh
init_network "$VAL_ACCOUNTS_DIR" "$VALIDATORS" "$CHAIN_ID" "$NATIVE_CURRENCY" "$VAL_TOKENS" \
              "$VAL_STAKE" "$accounts_spec" "$WASM_CODE_PATH" "$TREASURY_NLS_U128"

start_validators "$VALIDATORS"