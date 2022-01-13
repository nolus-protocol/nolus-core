#!/bin/bash
set -euxo pipefail

SCRIPT_DIR=$(cd $(dirname "${BASH_SOURCE[0]}") && pwd)

source "$SCRIPT_DIR"/internal/cmd.sh

create_periodic_vesting () {
	local AMOUNT
	AMOUNT=$(tr -dc '0-9' <<< "$1")
	local CURRENCY
	CURRENCY=$(tr -d '0-9' <<< "$1")
	local LENGTH=$2 # in seconds
	local VESTING_PERIODS=$3
	local ACC_NUM=$4
	local home=$5

	local P
	P=$(( "$LENGTH" / "$VESTING_PERIODS"))
	local R
	R=$(( "$LENGTH" % "$VESTING_PERIODS"))

	local PA
	PA=$(( "$AMOUNT" / "$VESTING_PERIODS"))
	local RA
	RA=$(( "$AMOUNT" % "$VESTING_PERIODS"))

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
		jq --arg a "$PA" --argjson n "$ACC_NUM" --arg p "$P" --argjson i "$i" '.app_state["auth"]["accounts"][$n]["vesting_periods"][$i] = { "length": $p, "amount": [ { "amount": $a, "denom": "'"$CURRENCY"'" } ] }' <"$home/config/genesis.json" >"$home/config/genesis.json.tmp" && mv "$home/config/genesis.json.tmp" "$home/config/genesis.json"
	done
}

add_vesting_account() {
  local row
  row="$1"
  local home
  home="$2"
  local address
  address=$(jq -r  '.address' <<< "$row")
  local amount
  amount=$(jq -r  '.amount' <<< "$row")
  local vesting_start_time
  vesting_start_time=$(jq -r '.vesting."start-time"' <<< "$row")
  local vesting_end_time
  vesting_end_time=$(jq -r '.vesting."end-time"' <<< "$row")
  local vesting_amount
  vesting_amount=$(jq -r '.vesting.amount' <<< "$row")
  local vesting_length
  vesting_length=$(jq -r '.vesting.length' <<< "$row")
  local vesting_periods
  vesting_periods=$(jq -r '.vesting.periods' <<< "$row")
  local type
  type=$(jq -r '.vesting.type' <<< "$row")
  local VESTING_START=""
  if [[ -n "$vesting_start_time" ]]; then
    VESTING_START="--vesting-start-time $vesting_start_time"
  fi
  run_cmd "$home" add-genesis-account "$address" "$amount" --vesting-amount "$vesting_amount" \
    --vesting-end-time "$vesting_end_time" "$VESTING_START"
  if [[ "$type" == "periodic" ]]; then
    index=$(jq '."app_state"["auth"]["accounts"] | map(."base_vesting_account"."base_account"."address" == "'"$address"'") | index(true)' "$home/config/genesis.json")
    jq --arg i "$index" '.app_state["auth"]["accounts"][$i|tonumber]["@type"]="/cosmos.vesting.v1beta1.PeriodicVestingAccount"' <"$home/config/genesis.json" >"$home/config/genesis.json.tmp" && mv "$home/config/genesis.json.tmp" "$home/config/genesis.json"
    jq --arg i "$index" '.app_state["auth"]["accounts"][$i|tonumber]["vesting_periods"]+=[]' <"$home/config/genesis.json" >"$home/config/genesis.json.tmp" && mv "$home/config/genesis.json.tmp" "$home/config/genesis.json"

    create_periodic_vesting "$vesting_amount" "$vesting_length" "$vesting_periods" "$index" "$home"
  fi
}