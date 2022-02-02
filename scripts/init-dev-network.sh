#!/bin/bash
set -euxo pipefail

SCRIPT_DIR=$(cd $(dirname "${BASH_SOURCE[0]}") && pwd)
source "$SCRIPT_DIR"/common/cmd.sh
source "$SCRIPT_DIR"/internal/accounts.sh

cleanup() {
  cleanup_init_network_sh
  cleanup_setup_validator_dev_sh
  exit
}
trap cleanup INT TERM EXIT

VAL_ROOT_DIR="networks/nolus"
VALIDATORS=1
VAL_ACCOUNTS_DIR="$VAL_ROOT_DIR/val-accounts"
POSITIONAL=()
GIT_REF_SLUG=main

NATIVE_CURRENCY="unolus"
VAL_TOKENS="1000000000""$NATIVE_CURRENCY"
VAL_STAKE="1000000""$NATIVE_CURRENCY"
CHAIN_ID="nolus-private"
SUSPEND_ADMIN=""
FAUCET_MNEMONIC=""
FAUCET_TOKENS="1000000""$NATIVE_CURRENCY"

while [[ $# -gt 0 ]]; do
  key="$1"

  case $key in

  -h | --help)
    printf \
    "Usage: %s
    [--git-source-ref <branch_or_tag>]
    [--chain_id <string>]
    [--validators_dir <validators_root_dir>]
    [-v|--validators <number>]
    [--validator_accounts_dir <validator_accounts_dir>]
    [--currency <native_currency>]
    [--validator-tokens <tokens_for_val_genesis_accounts>]
    [--validator-stake <tokens_val_will_stake>]
    [-ips <ip_addrs>]
    [--suspend-admin <bech32address>]
    [--faucet-mnemonic <mnemonic_phrase>]
    [--faucet-tokens <initial_balance>]"
     "$0"
    exit 0
    ;;

   --git-source-ref)
    GIT_REF_SLUG="$2"
    shift
    shift
    ;;
  
   --chain-id)
    CHAIN_ID="$2"
    shift
    shift
    ;;

   --validators_dir)
    VAL_ROOT_DIR="$2"
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

  --suspend-admin)
    SUSPEND_ADMIN="$2"
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
  *) # unknown option
    POSITIONAL+=("$1") # save it in an array for later
    shift              # past argument
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

__add_faucet_account() {
  local mnemonic="$1"
  local amount="$2"

  local account_name="faucet"
  local tmp_faucet_dir
  tmp_faucet_dir="$(mktemp -d)"
  run_cmd "$tmp_faucet_dir" keys add --recover "$account_name" --keyring-backend test <<< "$mnemonic" 1>/dev/null
  local addr
  addr="$(run_cmd "$tmp_faucet_dir" keys show "$account_name" -a --keyring-backend test)"
  add_account "$addr" "$amount"
}


__verify_mandatory "$SUSPEND_ADMIN" "Suspend admin"
__verify_mandatory "$FAUCET_MNEMONIC" "Faucet mnemonic"

# TBD open a few sample private investor accounts
# TBD open admin accounts, e.g. a treasury and a suspender
#  and pass them to init_network
accounts_spec=$(echo "[]" | __add_faucet_account "$FAUCET_MNEMONIC" "$FAUCET_TOKENS")

ARTIFACT_URL="https://gitlab-nomo.credissimo.net/nomo/cosmzone/-/jobs/artifacts/$GIT_REF_SLUG/download?job="
ARTIFACT_BIN_URL="$ARTIFACT_URL"build-binary
ARTIFACT_REMOTE_SCRIPTS_URL="$ARTIFACT_URL"publish-remote-scripts
source "$SCRIPT_DIR"/internal/setup-validator-dev.sh
init_setup_validator_dev_sh "$SCRIPT_DIR" "$ARTIFACT_BIN_URL" "$VAL_ROOT_DIR"

source "$SCRIPT_DIR"/internal/init-network.sh
init_network "$VAL_ACCOUNTS_DIR" "$VALIDATORS" "$CHAIN_ID" "$NATIVE_CURRENCY" "$SUSPEND_ADMIN" "$VAL_TOKENS" "$VAL_STAKE" "$accounts_spec"
