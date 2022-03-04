#!/bin/bash
set -euxo pipefail

SCRIPTS_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
source "$SCRIPTS_DIR"/create-vesting-account.sh
source "$SCRIPTS_DIR"/common/cmd.sh

ROOT_DIR=$1
shift
if [[ -n ${ROOT_DIR+} ]]; then
  echo "root directory was not set"
  exit 1
fi

TESTS_DIR="$ROOT_DIR/tests/integration"
NET_ROOT_DIR="$ROOT_DIR/networks/nolus"
HOME_DIR="$NET_ROOT_DIR/local-validator-1"
VAL_ACCOUNTS_DIR="$NET_ROOT_DIR/val-accounts"
USER_DIR="$NET_ROOT_DIR/users"
IBC_TOKEN='ibc/11DFDFADE34DCE439BA732EBA5CD8AA804A544BA1ECC0882856289FAF01FE53F'
NLS_CURRENCY="unolus"
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

__now_shifted_with_hours() {
  local delta_hours="$1"
  date -d @$(($(date +%s) + delta_hours*60*60)) -Iseconds
}
create_vested_account() {
  run_cmd "$USER_DIR" keys add periodic-vesting-account --keyring-backend "test"
  PERIODIC_VEST=$(run_cmd "$USER_DIR" keys show periodic-vesting-account -a --keyring-backend "test")
  local start_time
  start_time=$(__now_shifted_with_hours -1)
  local end_time
  end_time=$(__now_shifted_with_hours 4)
  local amnt="546652"
  local spec
  spec="{\"address\": \"$PERIODIC_VEST\", \"amount\": \"$amnt$NLS_CURRENCY\", \
        \"vesting\": { \"type\": \"periodic\", \"start-time\": \"$start_time\", \"end-time\": \"$end_time\", \"amount\": \"$amnt\", \
                        \"periods\": 4, \"length\": 14400}}"
  add_vesting_account "$spec" "$NLS_CURRENCY" "$HOME_DIR"
}

prepare_env() {
  rm -fr "$NET_ROOT_DIR"
  "$SCRIPTS_DIR"/init-local-network.sh --validators-root-dir "$NET_ROOT_DIR" -v 1 \
      --validator-accounts-dir "$VAL_ACCOUNTS_DIR" \
      --validator-tokens "100000000000$NLS_CURRENCY,1000000000$IBC_TOKEN" \
      --treasury-tokens "1000000000000$NLS_CURRENCY, \
                        1000000000000ibc/0954E1C28EB7AF5B72D24F3BC2B47BBB2FDF91BDDFD57B74B99E133AED40972A, \
                        1000000000000ibc/0EF15DF2F02480ADE0BB6E85D9EBB5DAEA2836D3860E9F97F9AADE4F57A31AA0, \
                        1000000000000ibc/8A34AF0C1943FD0DFCDE9ADBF0B2C9959C45E87E6088EA2FC6ADACD59261B8A2" \
      --user-dir "$USER_DIR" 2>&1

  # create_ibc_network

# TBD when finalize vesting type, periodic vs continuos, do create vesting accounts
# feeding `init-local-network` with necessary account data
  create_vested_account
  run_cmd "$USER_DIR" keys add test-user-1 --keyring-backend "test" # force no password
  run_cmd "$USER_DIR" keys add test-user-2 --keyring-backend "test" # force no password
  run_cmd "$USER_DIR" keys add test-delayed-vesting --keyring-backend "test" # force no password

  VALIDATOR_KEY_NAME=$(run_cmd "$VAL_ACCOUNTS_DIR" keys list --list-names)
  VALIDATOR_PRIV_KEY=$(echo 'y' | nolusd keys export "$VALIDATOR_KEY_NAME" --unsafe --unarmored-hex --keyring-backend "test" --home "$VAL_ACCOUNTS_DIR" 2>&1)
  SUSPEND_ADMIN_PRIV_KEY="$(echo 'y' | nolusd keys export suspend-admin --unsafe --unarmored-hex --keyring-backend "test" --home "$USER_DIR" 2>&1 )"
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
    rm -fr "$MARS_ROOT_DIR"
    local MARS_HOME_DIR="$MARS_ROOT_DIR/local-validator-1"
    local MARS_VAL_ACCOUNTS_DIR="$MARS_ROOT_DIR/val-accounts"
    local MARS_USER_DIR="$MARS_ROOT_DIR/users"
    "$SCRIPTS_DIR"/init-local-network.sh --currency 'mars' --chain-id 'mars-private' \
      --validators-root-dir "$MARS_ROOT_DIR" -v 1 --validator-accounts-dir "$MARS_VAL_ACCOUNTS_DIR" \
      --validator-tokens '100000000000mars' --validator-stake '1000000mars' \
      --user-dir "$MARS_USER_DIR"
    "$SCRIPTS_DIR"/remote/edit.sh --home "$MARS_HOME_DIR" \
      --tendermint-rpc-address "tcp://127.0.0.1:26667" --tendermint-p2p-address "tcp://0.0.0.0:26666" \
      --enable-api false --enable-grpc false --grpc-address "0.0.0.0:9095" \
      --enable-grpc-web false --grpc-web-address "0.0.0.0:9096" \
      --timeout-commit '1s'
    nolusd start --home "$MARS_HOME_DIR" >$LOG_DIR/mars-run.log 2>&1 &
    MARS_PID=$!
}

prepare_env

cd "$TESTS_DIR"
yarn install
yarn test "$@"
