#!/bin/bash
set -euxo pipefail

ROOT_DIR=$1
shift
if [[ -n ${ROOT_DIR+} ]]; then
  echo "root directory was not set"
  exit 1
fi

TESTS_DIR="$ROOT_DIR/tests/integration"
HOME_DIR="$ROOT_DIR/validator_setup/node1"
IBC_TOKEN='ibc/11DFDFADE34DCE439BA732EBA5CD8AA804A544BA1ECC0882856289FAF01FE53F'
LOG_DIR="/tmp"

command -v create-vesting-account.sh >/dev/null 2>&1 || {
  echo >&2 "scripts are not found in \$PATH."
  exit 1
}

source "create-vesting-account.sh"

cleanup() {
  if [ -n "${COSMZONED_PID:-}" ]; then
    echo "Stopping cosmzone..."
    kill -7 "$COSMZONED_PID"
    exit
  fi
}

trap cleanup INT TERM EXIT

create_vested_account() {
  cosmzoned keys add vested-user-1 --keyring-backend "test"  --home "$HOME_DIR"
  VESTED_USR_1_ADDR=$(cosmzoned keys show vested-user-1 -a --home "$HOME_DIR" --keyring-backend "test")
  local TILL4H
  TILL4H=$(($(date +%s) + 14400))
  local amnt
  amnt='546652nomo'
  row="{\"address\": \"$VESTED_USR_1_ADDR\", \"amount\": \"$amnt\", \"vesting\": { \"type\": \"periodic\", \"start-time\": \"$(date +%s)\", \"end-time\": \"$TILL4H\", \"amount\": \"$amnt\", \"periods\": 4, \"length\": 14400}}"
  add_vesting_account "$row" "$HOME_DIR"
}

prepare_env() {
  create_vested_account
  cosmzoned keys add test-user-1 --keyring-backend "test" --home "$HOME_DIR" # force no password
  cosmzoned keys add test-user-2 --keyring-backend "test" --home "$HOME_DIR" # force no password

  VALIDATOR_ADDR=$(cosmzoned keys show validator-key -a --home "$HOME_DIR" --keyring-backend "test")
  VALIDATOR_PRIV_KEY=$(echo 'y' | cosmzoned keys  export validator-key --unsafe --unarmored-hex --home "$HOME_DIR" --keyring-backend "test" 2>&1)
  VESTED_USR_1_ADDR=$(cosmzoned keys show vested-user-1 -a --home "$HOME_DIR" --keyring-backend "test")
  USR_1_ADDR=$(cosmzoned keys show test-user-1 -a --home "$HOME_DIR" --keyring-backend "test")
  USR_2_ADDR=$(cosmzoned keys show test-user-2 -a --home "$HOME_DIR" --keyring-backend "test")
  VESTED_USR_1_PRIV_KEY=$(echo 'y' | cosmzoned keys  export vested-user-1 --unsafe --unarmored-hex --home "$HOME_DIR" --keyring-backend "test" 2>&1)
  USR_1_PRIV_KEY=$(echo 'y' | cosmzoned keys  export test-user-1 --unsafe --unarmored-hex --home "$HOME_DIR" --keyring-backend "test" 2>&1)
  USR_2_PRIV_KEY=$(echo 'y' | cosmzoned keys  export test-user-2 --unsafe --unarmored-hex --home "$HOME_DIR" --keyring-backend "test" 2>&1)
  DOT_ENV=$(cat <<-EOF
NODE_URL=http://localhost:26657
VALIDATOR_ADDR=${VALIDATOR_ADDR}
VALIDATOR_PRIV_KEY=${VALIDATOR_PRIV_KEY}
USR_1_ADDR=${USR_1_ADDR}
USR_1_PRIV_KEY=${USR_1_PRIV_KEY}
USR_2_ADDR=${USR_2_ADDR}
USR_2_PRIV_KEY=${USR_2_PRIV_KEY}
VESTED_USR_1_ADDR=${VESTED_USR_1_ADDR}
VESTED_USR_1_PRIV_KEY=${VESTED_USR_1_PRIV_KEY}
IBC_TOKEN=${IBC_TOKEN}
EOF
  )
  echo "$DOT_ENV" > .env
}

init-test-network.sh -v 1 --validator-tokens "100000000000nomo,1000000000$IBC_TOKEN" 2>&1

edit-configuration.sh --home "$HOME_DIR" --enable-api true --enable-grpc true --enable-grpc-web true --timeout-commit '1s'


cd "$TESTS_DIR"
prepare_env
cosmzoned start --home "$HOME_DIR" >$LOG_DIR/cosmzone-run.log 2>&1 &
COSMZONED_PID=$!
sleep 5

yarn install
yarn test $@
