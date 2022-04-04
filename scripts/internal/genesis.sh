#!/bin/bash

SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
source "$SCRIPT_DIR"/validators-manager.sh
source "$SCRIPT_DIR"/accounts.sh
source "$SCRIPT_DIR"/../common/cmd.sh
"$SCRIPT_DIR"/check-jq.sh

# start "instance" variables
genesis_home_dir=$(mktemp -d)
genesis_file="$genesis_home_dir"/config/genesis.json
# end "instance" variables

WASM_BIN_PATH="$SCRIPT_DIR/wasmbin"

cleanup_genesis_sh() {
  if [[ -n "${genesis_home_dir:-}" ]]; then
    rm -rf "$genesis_home_dir"
  fi
}

generate_genesis() {
  local -r val_accounts_dir="$1"
  local -r chain_id="$2"
  local -r native_currency="$3"
  local -r val_tokens="$4"
  local -r val_stake="$5"
  local -r genesis_accounts_spec="$6"
  local -r node_id_and_val_pubkeys="$7"

  local -r proto_genesis_file="$val_accounts_dir/penultimate-genesis.json"
  local -r final_genesis_file="$val_accounts_dir/genesis.json"

  local -r acl_bpath="$WASM_BIN_PATH/acl.wasm"
  local -r treasury_bpath="$WASM_BIN_PATH/treasury.wasm"
  # TODO Add proper account
  local -r smartcontract_admin_addr="nolus1ga3l8gj8kpddksvgdly4qrs597jejkf8yl8kly"


  init_val_mngr_sh "$val_accounts_dir" "$chain_id"
  val_addrs="$(__gen_val_accounts "$node_id_and_val_pubkeys")"
  local accounts_spec="$genesis_accounts_spec"
  accounts_spec="$(__add_val_accounts "$accounts_spec" "$val_addrs" "$val_tokens")"
  generate_proto_genesis "$chain_id" "$accounts_spec" "$native_currency" "$proto_genesis_file"
  create_validator_txs="$(__init_validators "$proto_genesis_file" "$node_id_and_val_pubkeys" "$val_stake")"
  integrate_genesis_txs "$proto_genesis_file" "$create_validator_txs" "$final_genesis_file"
  add_wasm_genesis_message "$acl_bpath" "$treasury_bpath" "$smartcontract_admin_addr" "$final_genesis_file"
}

generate_proto_genesis() {
  local chain_id="$1"
  local accounts_spec="$2"
  local currency="$3"
  local proto_genesis_file="$4"

  run_cmd "$genesis_home_dir" init genesis_manager --chain-id "$chain_id"
  run_cmd "$genesis_home_dir" config keyring-backend test
  run_cmd "$genesis_home_dir" config chain-id "$chain_id"

  __set_token_denominations "$genesis_file" "$currency"

  while IFS= read -r account_spec ; do
    add_genesis_account "$account_spec" "$currency" "$genesis_home_dir"
  done <<< "$(echo "$accounts_spec" | jq -c '.[]')"

  cp "$genesis_file" "$proto_genesis_file"
}

#
# Takes a json object and creates a genesis account
#
# JSON specification object:
# "address" - mandatory string
# "amount" - mandatory string in the form '<number><currency>[,<number><currency>]*'
# "vesting" - optional object
# "vesting.start-time" - optional string representing a datetime in ISO 8601 format with max precision in seconds,
#                         for example "2022-01-28T13:15:59+02:00"
# "vesting.end-time" - mandatory string representing a datetime in ISO 8601 format with max precision in seconds,
#                         for example "2022-01-30T15:15:59-06:00"
# "vesting.amount" - mandatory number in native currency, e.g. 100 means "100 unolus"
add_genesis_account() {
  local specification="$1"
  local currency="$2"
  # TBD remove the following argument once the periodic vesting testing is deleted
  local home_dir="$3"

  local address
  address=$(jq -r '.address' <<< "$specification")
  local amount
  amount=$(jq -r '.amount' <<< "$specification")
  if [[ "$(jq -r '.vesting' <<< "$specification")" != 'null' ]]; then
    local vesting_start_time=""
    if [[ "$(jq -r '.vesting."start-time"' <<< "$specification")" != 'null' ]]; then
      vesting_start_time="--vesting-start-time $(__read_unix_time "$specification" start-time)"
    fi

    local vesting_end_time
    vesting_end_time=$(echo "$specification" | jq -r '.vesting."end-time"' | __as_unix_time )
    local vesting_amount
    vesting_amount=$(jq -r '.vesting.amount' <<< "$row")
    run_cmd "$home_dir" add-genesis-account "$address" "$amount" \
                --vesting-amount "$vesting_amount$currency" \
                --vesting-end-time "$vesting_end_time" $vesting_start_time
  else
    run_cmd "$home_dir" add-genesis-account "$address" "$amount"
  fi
}

integrate_genesis_txs() {
  local genesis_in_file="$1"
  local txs="$2"
  local genesis_out_file="$3"

  cp "$genesis_in_file" "$genesis_file"

  local txs_dir="$genesis_home_dir"/txs
  {
    mkdir "$txs_dir"
    local index=0
    while IFS= read -r tx ; do
        echo "$tx" > "$txs_dir"/tx"$index".json
        index=$((index+1))
    done <<< "$txs"
  }

  run_cmd "$genesis_home_dir" collect-gentxs --gentx-dir "$txs_dir"
  cp "$genesis_file" "$genesis_out_file"
}

add_wasm_genesis_message() {
  local acl_bpath="$1"
  local treasury_bpath="$2"
  local admin_addr="$3"
  local genesis_in_out_file="$4"
  local trs_inst='{"acl":"nolus14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9s0k0puz"}'

  cp "$genesis_in_out_file" "$genesis_file"

  run_cmd "$genesis_home_dir" add-wasm-genesis-message store "$acl_bpath" --run-as "$admin_addr"
  run_cmd "$genesis_home_dir" add-wasm-genesis-message instantiate-contract 1 {} --label acl --run-as "$admin_addr" --admin "$admin_addr"
  run_cmd "$genesis_home_dir" add-wasm-genesis-message store "$treasury_bpath" --run-as "$admin_addr"
  run_cmd "$genesis_home_dir" add-wasm-genesis-message instantiate-contract 2 "$trs_inst" --label treasury --run-as "$admin_addr" --admin "$admin_addr"

  cp "$genesis_file" "$genesis_in_out_file"
}

#####################
# private functions #
#####################
__set_token_denominations() {
  local genesis_file="$1"
  local currency="$2"

  local genesis_tmp_file="$genesis_file".tmp

  < "$genesis_file" \
    jq '.app_state["staking"]["params"]["bond_denom"]="'"$currency"'"' \
    | jq '.app_state["crisis"]["constant_fee"]["denom"]="'"$currency"'"' \
    | jq '.app_state["gov"]["deposit_params"]["min_deposit"][0]["denom"]="'"$currency"'"' \
    | jq '.app_state["gov"]["deposit_params"]["min_deposit"][0]["denom"]="'"$currency"'"' \
    | jq '.app_state["mint"]["params"]["mint_denom"]="'"$currency"'"' > "$genesis_tmp_file"
  mv "$genesis_tmp_file" "$genesis_file"
}

__read_unix_time() {
  local spec="$1"
  local time_prop="$2"

  echo "$spec" | jq -r ".vesting.\"$time_prop\"" | __as_unix_time
}

__as_unix_time() {
  local datetime;
  read -r datetime
  date --date "$datetime" +%s
}

__gen_val_accounts() {
  local node_id_and_val_pubkeys="$1"
  while IFS= read -r node_id_and_val_pubkey ; do
    local account_name
    read -r account_name __val_pub_key <<< $node_id_and_val_pubkey
    local address
    address=$(gen_val_account "$account_name")
    echo "$address"
  done <<< "$node_id_and_val_pubkeys"
}

__add_val_accounts() {
  local account_spec="$1"
  local val_addrs="$2"
  local val_tokens="$3"

  while IFS= read -r address ; do
    account_spec=$(echo "$account_spec" | add_account "$address" "$val_tokens")
  done <<< "$val_addrs"
  echo "$account_spec"
}

__init_validators() {
  local proto_genesis_file="$1"
  local node_id_and_val_pubkeys="$2"
  local val_stake="$3"

  while IFS= read -r node_id_and_val_pubkey ; do
    local node_id
    local val_pub_key
    read -r node_id val_pub_key <<< $node_id_and_val_pubkey
    local create_validator_tx
    create_validator_tx=$(gen_val_txn "$proto_genesis_file" "$node_id" "$val_pub_key" "$val_stake")
    echo "$create_validator_tx"
  done <<< "$node_id_and_val_pubkeys"
}
