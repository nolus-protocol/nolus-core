#!/bin/bash
set -euxo pipefail

cleanup() {
  if [[ -n "${COSMZONED_PID:-}" ]]; then
    echo "Stopping cosmzone..."
    kill -7 "$COSMZONED_PID"
    exit
  fi
}

trap cleanup INT TERM EXIT

prepare_env() {
  cosmzoned keys add test-user-1 --keyring-backend "test" --home "$HOME_DIR" # force no password
  cosmzoned keys add test-user-2 --keyring-backend "test" --home "$HOME_DIR" # force no password

  VALIDATOR_ADDR=$(cosmzoned keys show validator-key -a --home "$HOME_DIR" --keyring-backend "test")
  VALIDATOR_PRIV_KEY=$(echo 'y' | cosmzoned keys  export validator-key --unsafe --unarmored-hex --home "$HOME_DIR" --keyring-backend "test" 2>&1)
  USR_1_ADDR=$(cosmzoned keys show test-user-1 -a --home "$HOME_DIR" --keyring-backend "test")
  USR_2_ADDR=$(cosmzoned keys show test-user-2 -a --home "$HOME_DIR" --keyring-backend "test")
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
IBC_TOKEN=$IBC_TOKEN
EOF
  )
  echo "$DOT_ENV" > .env
}

TEST_DIR=$(pwd)
cd ../../scripts
IBC_TOKEN='ibc/11DFDFADE34DCE439BA732EBA5CD8AA804A544BA1ECC0882856289FAF01FE53F'
./init-test-network.sh -v 1 --validator-tokens "100000000000nomo,1000000000$IBC_TOKEN" 2>&1
HOME_DIR=$(realpath ./validator_setup/node1)
./edit-configuration.sh --home "$HOME_DIR" --enable-api true --enable-grpc true --enable-grpc-web true --timeout_commit '1s'

cd "$TEST_DIR"

cosmzoned start --home "$HOME_DIR" >/tmp/cosmzone-run.log 2>&1 &
COSMZONED_PID=$!
sleep 5

prepare_env
yarn test "$@"
