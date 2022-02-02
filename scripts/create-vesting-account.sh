#!/bin/bash
set -euxo pipefail

SCRIPT_DIR=$(cd $(dirname "${BASH_SOURCE[0]}") && pwd)

source "$SCRIPT_DIR"/common/cmd.sh
source "$SCRIPT_DIR"/internal/genesis.sh

create_periodic_vesting () {
	local amount="$1"
	local currency="$2"
	local LENGTH=$3 # in seconds
	local VESTING_PERIODS=$4
	local ACC_NUM=$5
	local home=$6

	local P
	P=$(( "$LENGTH" / "$VESTING_PERIODS"))
	local R
	R=$(( "$LENGTH" % "$VESTING_PERIODS"))

	local PA
	PA=$(( "$amount" / "$VESTING_PERIODS"))
	local RA
	RA=$(( "$amount" % "$VESTING_PERIODS"))

	if [ "$R" -gt 0 ]
	then
		echo 'Fix length' $P
		exit 1
	fi

	if [ "$RA" -gt 0 ]
	then
		echo 'Fix amount' $RA
		exit 1
	fi

	for (( i=0; i<"$VESTING_PERIODS"; i++ ))
	do
		jq --arg a "$PA" --argjson n "$ACC_NUM" --arg p "$P" --argjson i "$i" '.app_state["auth"]["accounts"][$n]["vesting_periods"][$i] = { "length": $p, "amount": [ { "amount": $a, "denom": "'"$currency"'" } ] }' <"$home/config/genesis.json" >"$home/config/genesis.json.tmp" && mv "$home/config/genesis.json.tmp" "$home/config/genesis.json"
	done
}

add_vesting_account() {
  local row="$1"
  local currency="$2"
  local home="$3"
  add_genesis_account "$row" "$currency" "$home"

  local type
  type=$(jq -r '.vesting.type' <<< "$row")
  if [[ "$type" == "periodic" ]]; then
    local address
    address=$(jq -r  '.address' <<< "$row")
    local vesting_amount
    vesting_amount=$(jq -r '.vesting.amount' <<< "$row")
    local vesting_length
    vesting_length=$(jq -r '.vesting.length' <<< "$row")
    local vesting_periods
    vesting_periods=$(jq -r '.vesting.periods' <<< "$row")
    index=$(jq '."app_state"["auth"]["accounts"] | map(."base_vesting_account"."base_account"."address" == "'"$address"'") | index(true)' "$home/config/genesis.json")
    jq --arg i "$index" '.app_state["auth"]["accounts"][$i|tonumber]["@type"]="/cosmos.vesting.v1beta1.PeriodicVestingAccount"' <"$home/config/genesis.json" >"$home/config/genesis.json.tmp" && mv "$home/config/genesis.json.tmp" "$home/config/genesis.json"
    jq --arg i "$index" '.app_state["auth"]["accounts"][$i|tonumber]["vesting_periods"]+=[]' <"$home/config/genesis.json" >"$home/config/genesis.json.tmp" && mv "$home/config/genesis.json.tmp" "$home/config/genesis.json"

    create_periodic_vesting "$vesting_amount" "$currency" "$vesting_length" "$vesting_periods" "$index" "$home"
  fi
}