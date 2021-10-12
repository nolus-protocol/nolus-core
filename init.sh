#!/bin/bash


# This script is a modified version of Ethermint's init.sh script https://github.com/tharsis/ethermint/blob/main/init.sh

KEY="local-validator"
CHAINID="nomo-private"
MONIKER="localtestnet"
KEYRING="test"

update_config () {
  if [[ "$OSTYPE" == "darwin"* ]]; then
    sed -i '' "$1" "$HOME/.cosmzone/config/config.toml"
  else
    sed -i 's/create_empty_blocks_interval = "0s"/create_empty_blocks_interval = "30s"/g' "$HOME/.cosmzone/config/config.toml"
  fi
}

update_genesis () {
  cat "$HOME/.cosmzone/config/genesis.json" | jq "$1" > "$HOME/.cosmzone/config/tmp_genesis.json" && mv "$HOME/.cosmzone/config/tmp_genesis.json" "$HOME/.cosmzone/config/genesis.json"
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

cosmzoned init $MONIKER --chain-id $CHAINID

# Change parameter token denominations to nomo
update_genesis '.app_state["staking"]["params"]["bond_denom"]="nomo"'
update_genesis '.app_state["crisis"]["constant_fee"]["denom"]="nomo"'
update_genesis '.app_state["gov"]["deposit_params"]["min_deposit"][0]["denom"]="nomo"'
update_genesis '.app_state["gov"]["deposit_params"]["min_deposit"][0]["denom"]="nomo"'
update_genesis '.app_state["mint"]["params"]["mint_denom"]="nomo"'
# Allocate genesis accounts (cosmos formatted addresses)
cosmzoned add-genesis-account $KEY 10000000nomo --keyring-backend $KEYRING

# Sign genesis transaction
cosmzoned gentx $KEY 1000000nomo --keyring-backend $KEYRING --chain-id $CHAINID

# Collect genesis tx
cosmzoned collect-gentxs

# Run this to ensure everything worked and that the genesis file is setup correctly
cosmzoned validate-genesis

if [[ $1 != "prepare" ]]; then
cosmzoned start
else
  echo "Network prepared. You can start it with the command: 'cosmzoned start'"
fi