#!/bin/bash
set -euxo pipefail

THIS_SCRIPT_DIR=$(cd $(dirname "${BASH_SOURCE[0]}") && pwd)
SCRIPTS_DIR="$THIS_SCRIPT_DIR"/..
source "$SCRIPTS_DIR"/create-vesting-account.sh

ROOT_DIR=$1
shift
if [[ -n ${ROOT_DIR+} ]]; then
  echo "root directory was not set"
  exit 1
fi

TESTS_DIR="$ROOT_DIR/tests/integration"
NET_ROOT_DIR="$ROOT_DIR/networks/nolus"
HOME_DIR="$NET_ROOT_DIR/dev-validator-1"
IBC_TOKEN='ibc/11DFDFADE34DCE439BA732EBA5CD8AA804A544BA1ECC0882856289FAF01FE53F'
LOG_DIR="/tmp"

cleanup() {
  if [ -n "${COSMZONED_PID:-}" ]; then
    echo "Stopping cosmzone..."
    kill -7 "$COSMZONED_PID"
  fi
  if [ -n "${MARS_PID:-}" ]; then
    echo "Stopping ibc network..."
    kill -7 "$MARS_PID"
  fi
  exit
}

trap cleanup INT TERM EXIT

create_vested_account() {
  cosmzoned keys add periodic-vesting-account --keyring-backend "test"  --home "$HOME_DIR"
  PERIODIC_VEST=$(cosmzoned keys show periodic-vesting-account -a --home "$HOME_DIR" --keyring-backend "test")
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
  suspend_admin_output="$(cosmzoned keys add suspend-admin --keyring-backend "test" --home "$HOME_DIR" --output json)"
  SUSPEND_ADMIN_ADDR="$(echo "$suspend_admin_output" | jq -r '.address')"
  SUSPEND_ADMIN_PRIV_KEY="$(echo 'y' | cosmzoned keys export suspend-admin --unsafe --unarmored-hex --home "$HOME_DIR" --keyring-backend "test" 2>&1)"
  "$SCRIPTS_DIR"/init-dev-network.sh -v 1 --validator-tokens "100000000000unolus,1000000000$IBC_TOKEN" --suspend-admin "$SUSPEND_ADMIN_ADDR"  --output "$NET_ROOT_DIR" 2>&1
  "$SCRIPTS_DIR"/edit-configuration.sh --home "$HOME_DIR" --timeout-commit '1s'

  create_ibc_network

  create_vested_account
  cosmzoned keys add test-user-1 --keyring-backend "test" --home "$HOME_DIR" # force no password
  cosmzoned keys add test-user-2 --keyring-backend "test" --home "$HOME_DIR" # force no password
  cosmzoned keys add test-delayed-vesting --keyring-backend "test" --home "$HOME_DIR" # force no password

  VALIDATOR_PRIV_KEY=$(echo 'y' | cosmzoned keys export dev-validator-1 --unsafe --unarmored-hex --home "$HOME_DIR" --keyring-backend "test" 2>&1)
  PERIODIC_PRIV_KEY=$(echo 'y' | cosmzoned keys export periodic-vesting-account --unsafe --unarmored-hex --home "$HOME_DIR" --keyring-backend "test" 2>&1)
  USR_1_PRIV_KEY=$(echo 'y' | cosmzoned keys export test-user-1 --unsafe --unarmored-hex --home "$HOME_DIR" --keyring-backend "test" 2>&1)
  USR_2_PRIV_KEY=$(echo 'y' | cosmzoned keys export test-user-2 --unsafe --unarmored-hex --home "$HOME_DIR" --keyring-backend "test" 2>&1)
  DELAYED_VESTING_PRIV_KEY=$(echo 'y' | cosmzoned keys export test-delayed-vesting --unsafe --unarmored-hex --home "$HOME_DIR" --keyring-backend "test" 2>&1)
  DOT_ENV=$(cat <<-EOF
NODE_URL=http://localhost:26657
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

  cosmzoned start --home "$HOME_DIR" >$LOG_DIR/cosmzone-run.log 2>&1 &
  COSMZONED_PID=$!
  sleep 5
}

create_ibc_network() {
    local MARS_ROOT_DIR="$ROOT_DIR/networks/ibc_network/"
    local MARS_HOME_DIR="$MARS_ROOT_DIR/dev-validator-1"
    "$SCRIPTS_DIR"/init-dev-network.sh -v 1 --currency 'mars' --validator-tokens '100000000000mars' --validator-stake '1000000mars'\
      --chain-id 'mars-private' --suspend-admin 'nolus1jxguv8equszl0xus8akavgf465ppl2tzd8ac9k' --output "$MARS_ROOT_DIR"
    "$SCRIPTS_DIR"/edit-configuration.sh --home "$MARS_HOME_DIR" \
      --tendermint-rpc-address "tcp://127.0.0.1:26667" --tendermint-p2p-address "tcp://0.0.0.0:26666" \
      --enable-api false --enable-grpc false --grpc-address "0.0.0.0:9095" \
      --enable-grpc-web false --grpc-web-address "0.0.0.0:9096" \
      --timeout-commit '1s'
    cosmzoned start --home "$MARS_HOME_DIR" >$LOG_DIR/mars-run.log 2>&1 &
    MARS_PID=$!
}

prepare_env


cd "$TESTS_DIR"
yarn install
yarn test $@
