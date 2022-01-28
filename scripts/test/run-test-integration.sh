#!/bin/bash
set -euxo pipefail

SCRIPTS_DIR=$(cd $(dirname "${BASH_SOURCE[0]}")/.. && pwd)
source "$SCRIPTS_DIR"/create-vesting-account.sh
source "$SCRIPTS_DIR"/internal/cmd.sh

ROOT_DIR=$1
shift
if [[ -n ${ROOT_DIR+} ]]; then
  echo "root directory was not set"
  exit 1
fi

TESTS_DIR="$ROOT_DIR/tests/integration"
NET_ROOT_DIR="$ROOT_DIR/networks/nolus"
HOME_DIR="$NET_ROOT_DIR/dev-validator-1"
VAL_ACCOUNTS_DIR="$NET_ROOT_DIR/val-accounts"
USER_DIR="$ROOT_DIR/users/test-integration"
IBC_TOKEN='ibc/11DFDFADE34DCE439BA732EBA5CD8AA804A544BA1ECC0882856289FAF01FE53F'
LOG_DIR="/tmp"

cleanup() {
  if [ -n "${NOLUSD_PID:-}" ]; then
    echo "Stopping nolus..."
    kill -7 "$NOLUSD_PID"
  fi
  if [ -n "${MARS_PID:-}" ]; then
    echo "Stopping ibc network..."
    kill -7 "$MARS_PID"
  fi
  exit
}

trap cleanup INT TERM EXIT

create_vested_account() {
  run_cmd "$USER_DIR" keys add periodic-vesting-account --keyring-backend "test"
  PERIODIC_VEST=$(run_cmd "$USER_DIR" keys show periodic-vesting-account -a --keyring-backend "test")
  local TILL4H
  TILL4H=$(($(date +%s) + 14400))
  local amnt
  amnt='546652unolus'
  row="{\"address\": \"$PERIODIC_VEST\", \"amount\": \"$amnt\", \"vesting\": { \"type\": \"periodic\", \"start-time\": \"$(($(date +%s) - 3600))\", \"end-time\": \"$TILL4H\", \"amount\": \"$amnt\", \"periods\": 4, \"length\": 14400}}"
  add_vesting_account "$row" "$HOME_DIR"
}

prepare_env() {
  # note that the suspend admin account will be deleted by the init-dev-network script, so we first need to save his private key for later use in the tests
  local suspend_admin_output
  suspend_admin_output="$(run_cmd "$USER_DIR" keys add suspend-admin --keyring-backend "test" --output json)"
  SUSPEND_ADMIN_ADDR="$(run_cmd "$USER_DIR" keys show suspend-admin -a --keyring-backend "test")"
  SUSPEND_ADMIN_PRIV_KEY="$(echo 'y' | nolusd keys export suspend-admin --unsafe --unarmored-hex --keyring-backend "test" --home "$USER_DIR" 2>&1 )"
# TBD switch to using 'local' network - a single node locally run network with all ports opened
  "$SCRIPTS_DIR"/init-dev-network.sh --validators_dir "$NET_ROOT_DIR" -v 1 --validator_accounts_dir "$VAL_ACCOUNTS_DIR" \
      --validator-tokens "100000000000unolus,1000000000$IBC_TOKEN" \
      --suspend-admin "$SUSPEND_ADMIN_ADDR" 2>&1
# TBD incorporate it in the local network node configuration
  "$SCRIPTS_DIR"/config/edit.sh --home "$HOME_DIR" --timeout-commit '1s'

  create_ibc_network

# TBD do create vesting accounts feeding `init-dev-network` with necessary account data
  create_vested_account
  run_cmd "$USER_DIR" keys add test-user-1 --keyring-backend "test" # force no password
  run_cmd "$USER_DIR" keys add test-user-2 --keyring-backend "test" # force no password
  run_cmd "$USER_DIR" keys add test-delayed-vesting --keyring-backend "test" # force no password

  VALIDATOR_KEY_NAME=$(run_cmd "$VAL_ACCOUNTS_DIR" keys list --list-names)
  VALIDATOR_PRIV_KEY=$(echo 'y' | nolusd keys export "$VALIDATOR_KEY_NAME" --unsafe --unarmored-hex --keyring-backend "test" --home "$VAL_ACCOUNTS_DIR" 2>&1)
  PERIODIC_PRIV_KEY=$(echo 'y' | nolusd keys export periodic-vesting-account --unsafe --unarmored-hex --keyring-backend "test" --home "$USER_DIR" 2>&1)
  USR_1_PRIV_KEY=$(echo 'y' | nolusd keys export test-user-1 --unsafe --unarmored-hex --keyring-backend "test" --home "$USER_DIR" 2>&1)
  USR_2_PRIV_KEY=$(echo 'y' | nolusd keys export test-user-2 --unsafe --unarmored-hex --keyring-backend "test" --home "$USER_DIR" 2>&1)
  DELAYED_VESTING_PRIV_KEY=$(echo 'y' | nolusd keys export test-delayed-vesting --unsafe --unarmored-hex --keyring-backend "test" --home "$USER_DIR" 2>&1)
  DOT_ENV=$(cat <<-EOF
NODE_URL=http://localhost:26612
VALIDATOR_PRIV_KEY=${VALIDATOR_PRIV_KEY}
USR_1_PRIV_KEY=${USR_1_PRIV_KEY}
USR_2_PRIV_KEY=${USR_2_PRIV_KEY}
PERIODIC_PRIV_KEY=${PERIODIC_PRIV_KEY}
DELAYED_VESTING_PRIV_KEY=${DELAYED_VESTING_PRIV_KEY}
SUSPEND_ADMIN_PRIV_KEY=${SUSPEND_ADMIN_PRIV_KEY}
IBC_TOKEN=${IBC_TOKEN}
EOF
  )
  echo "$DOT_ENV" > "$TESTS_DIR/.env"

  nolusd start --home "$HOME_DIR" >$LOG_DIR/nolus-run.log 2>&1 &
  
  NOLUSD_PID=$!
  sleep 5
}

create_ibc_network() {
    local MARS_ROOT_DIR="$ROOT_DIR/networks/ibc_network"
    local MARS_HOME_DIR="$MARS_ROOT_DIR/dev-validator-1"
    local MARS_VAL_ACCOUNTS_DIR="$MARS_ROOT_DIR/val-accounts"
    "$SCRIPTS_DIR"/init-dev-network.sh --currency 'mars' --chain-id 'mars-private' \
      --validators_dir "$MARS_ROOT_DIR" -v 1 --validator_accounts_dir "$MARS_VAL_ACCOUNTS_DIR" \
      --validator-tokens '100000000000mars' --validator-stake '1000000mars' \
      --suspend-admin 'nolus1jxguv8equszl0xus8akavgf465ppl2tzd8ac9k'
    "$SCRIPTS_DIR"/config/edit.sh --home "$MARS_HOME_DIR" \
      --tendermint-rpc-address "tcp://127.0.0.1:26667" --tendermint-p2p-address "tcp://0.0.0.0:26666" \
      --enable-api false --enable-grpc false --grpc-address "0.0.0.0:9095" \
      --enable-grpc-web false --grpc-web-address "0.0.0.0:9096" \
      --timeout-commit '1s'
    nolusd start --home "$MARS_HOME_DIR" >$LOG_DIR/mars-run.log 2>&1 &
    MARS_PID=$!
}

rm -fr "$USER_DIR"
prepare_env


cd "$TESTS_DIR"
yarn install
yarn test $@
