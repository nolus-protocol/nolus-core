#!/bin/bash
set -euo pipefail

# This script is a modified version of Ethermint's init.sh script https://github.com/tharsis/ethermint/blob/main/init.sh

KEY="local-validator"
CHAINID="nomo-private"
MONIKER="localtestnet"
KEYRING="test"
CFG_DIR="$HOME/.cosmzone/config"

VESTING_ACC="local-vesting-account"

update_config () {
  if [[ "$OSTYPE" == "darwin"* ]]; then
    sed -i '' "$1" "$CFG_DIR/config.toml"
  else
    sed -i "$1" "$CFG_DIR/config.toml"
  fi
}

update_genesis () {
  jq "$1" < "$CFG_DIR/genesis.json" > "$CFG_DIR/tmp_genesis.json" && mv "$CFG_DIR/tmp_genesis.json" "$CFG_DIR/genesis.json"
}

# validate dependencies are installed
command -v jq > /dev/null 2>&1 || { echo >&2 "jq not installed. More info: https://stedolan.github.io/jq/download/"; exit 1; }

# remove existing daemon and client
rm -rf ~/.cosmzone

make install

cosmzoned config keyring-backend $KEYRING
cosmzoned config chain-id $CHAINID

# if $KEY exists it should be deleted
cosmzoned keys add $KEY --keyring-backend $KEYRING
cosmzoned keys add $VESTING_ACC --keyring-backend $KEYRING

cosmzoned init $MONIKER --chain-id $CHAINID

# Change parameter token denominations to nomo
update_genesis '.app_state["staking"]["params"]["bond_denom"]="nomo"'
update_genesis '.app_state["crisis"]["constant_fee"]["denom"]="nomo"'
update_genesis '.app_state["gov"]["deposit_params"]["min_deposit"][0]["denom"]="nomo"'
update_genesis '.app_state["gov"]["deposit_params"]["min_deposit"][0]["denom"]="nomo"'
update_genesis '.app_state["mint"]["params"]["mint_denom"]="nomo"'
# Allocate genesis accounts (cosmos formatted addresses)
cosmzoned add-genesis-account $KEY 1000000000nomo --keyring-backend $KEYRING

# Add DelayedVestingAccount
TILL=$(date -d "+10 minutes" +%s)
cosmzoned add-genesis-account $VESTING_ACC 4325346435455nomo --vesting-amount 4325346435455nomo --vesting-end-time $TILL --keyring-backend $KEYRING

# Add second DelayedVestingAccount
TILL1H=$(date -d "+1 hours" +%s)
cosmzoned add-genesis-account nomo1332gau6khc5xw5854ldsnd6vte4u6rl34j9s6v 1343144nomo --vesting-amount 1343144nomo --vesting-end-time $TILL1H

# Sign genesis transaction
cosmzoned gentx $KEY 1000000nomo --keyring-backend $KEYRING --chain-id $CHAINID

# Collect genesis tx
cosmzoned collect-gentxs

# Run this to ensure everything worked and that the genesis file is setup correctly
cosmzoned validate-genesis

if [[ "$*" =~ "integration" ]]; then
  update_config 's/timeout_commit = "5s"/timeout_commit = "1s"/g'
fi

if [[ "$*" =~ "prepare" ]]; then
  echo "Network prepared. You can start it with the command: 'cosmzoned start'"
else
  cosmzoned start
fi
