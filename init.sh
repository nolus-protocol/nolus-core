#!/bin/bash
set -euo pipefail

# This script is a modified version of Ethermint's init.sh script https://github.com/tharsis/ethermint/blob/main/init.sh

KEY="local-validator"
CHAINID="nomo-private"
MONIKER="localtestnet"
KEYRING="test"
CFG_DIR="$HOME/.cosmzone/config"

VESTING_ACC="local-vesting-account"
PERIODIC_VEST="periodic-vesting-account"

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

create_periodic_vesting () {
	local AMOUNT=$1 
	local HP=$2
	local S=$3
	local N=$4

	local P=$(expr $HP / $S)
	local R=$(expr $HP % $S)

	local PA=$(expr $AMOUNT / $S)
	local RA=$(expr $AMOUNT % $S)	

	if [ $R -gt 0 ]
	then
		echo 'Fix length' $P 
		exit 1
	fi

	if [ $RA -gt 0 ]
	then
		echo 'Fix amount' $RA
		exit 1
	fi

	for (( i=0; i<$S; i++ ))
	do
		jq --arg a $PA --argjson n $N --arg p $P --argjson i $i '.app_state["auth"]["accounts"][$n]["vesting_periods"][$i] = { "length": $p, "amount": [ { "amount": $a, "denom": "nomo" } ] }' < "$CFG_DIR/genesis.json" > "$CFG_DIR/tmp_genesis.json" && mv "$CFG_DIR/tmp_genesis.json" "$CFG_DIR/genesis.json"
	done
}

# validate dependencies are installed
command -v jq > /dev/null 2>&1 || { echo >&2 "jq not installed. More info: https://stedolan.github.io/jq/download/"; exit 1; }

# remove existing daemon and client
rm -rf ~/.cosmzone

make install_local

export PATH=$PATH:$(go env GOPATH)/bin

cosmzoned config keyring-backend $KEYRING
cosmzoned config chain-id $CHAINID

# if $KEY exists it should be deleted
cosmzoned keys add $KEY --keyring-backend $KEYRING
cosmzoned keys add $VESTING_ACC --keyring-backend $KEYRING
cosmzoned keys add $PERIODIC_VEST --keyring-backend $KEYRING

cosmzoned init $MONIKER --chain-id $CHAINID

# Change parameter token denominations to nomo
update_genesis '.app_state["staking"]["params"]["bond_denom"]="nomo"'
update_genesis '.app_state["crisis"]["constant_fee"]["denom"]="nomo"'
update_genesis '.app_state["gov"]["deposit_params"]["min_deposit"][0]["denom"]="nomo"'
update_genesis '.app_state["gov"]["deposit_params"]["min_deposit"][0]["denom"]="nomo"'
update_genesis '.app_state["mint"]["params"]["mint_denom"]="nomo"'

# Allocate genesis accounts (cosmos formatted addresses)
if [[ "$*" =~ "integration" ]]; then
  IBC_TOKEN="ibc/11DFDFADE34DCE439BA732EBA5CD8AA804A544BA1ECC0882856289FAF01FE53F"
  cosmzoned add-genesis-account $KEY "1000000000nomo, 1000000000$IBC_TOKEN" --keyring-backend $KEYRING
else
  cosmzoned add-genesis-account $KEY '1000000000nomo' --keyring-backend $KEYRING
fi

# Add DelayedVestingAccount
TILL=$(date -d "+10 minutes" +%s)
cosmzoned add-genesis-account $VESTING_ACC 4325346435455nomo --vesting-amount 4325346435455nomo --vesting-end-time $TILL --keyring-backend $KEYRING

# Add ContinuousVestingAccount
TILL1H=$(date -d "+1 hours" +%s)
cosmzoned add-genesis-account nomo1332gau6khc5xw5854ldsnd6vte4u6rl34j9s6v 1343144nomo --vesting-amount 1343144nomo --vesting-end-time $TILL1H --vesting-start-time $(date +%s)

# Periodic Vesting Account

TILL4H=$(date -d "+4 hours" +%s)
PV1=546652
LENGTH=14400
PERIODS=4
cosmzoned add-genesis-account $PERIODIC_VEST 1343144nomo --vesting-amount ${PV1}nomo --vesting-end-time $TILL4H --vesting-start-time $(date +%s)

update_genesis '.app_state["auth"]["accounts"][3]["@type"]="/cosmos.vesting.v1beta1.PeriodicVestingAccount"'
update_genesis '.app_state["auth"]["accounts"][3]["vesting_periods"]+=[]'

create_periodic_vesting $PV1 $LENGTH $PERIODS 3

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
