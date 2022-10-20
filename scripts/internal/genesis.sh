#!/bin/bash

SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
source "$SCRIPT_DIR"/validators-manager.sh
source "$SCRIPT_DIR"/accounts.sh
source "$SCRIPT_DIR"/verify.sh
source "$SCRIPT_DIR"/../common/cmd.sh
"$SCRIPT_DIR"/check-jq.sh

# start "instance" variables
genesis_home_dir=$(mktemp -d)
genesis_file="$genesis_home_dir"/config/genesis.json
# end "instance" variables

cleanup_genesis_sh() {
  if [[ -n "${genesis_home_dir:-}" ]]; then
    rm -rf "$genesis_home_dir"
  fi
}

generate_genesis() {
  set -euo pipefail
  local -r chain_id="$1"
  local -r native_currency="$2"
  local -r val_tokens="$3"
  local -r val_stake="$4"
  local -r val_accounts_dir="$5"
  local -r accounts_spec_in="$6"
  local -r wasm_script_path="$7"
  local -r wasm_code_path="$8"
  local -r contracts_owner_addr="$9"
  local -r treasury_init_tokens_u128="${10}"
  local -r node_id_and_val_pubkeys="${11}"
  local -r lpp_native="${12}"
  local -r contracts_info_file="${13}"

  local -r treasury_init_tokens="$treasury_init_tokens_u128$native_currency"
  init_val_mngr_sh "$val_accounts_dir" "$chain_id"
  val_addrs="$(__gen_val_accounts "$node_id_and_val_pubkeys" "$val_accounts_dir")"
  local accounts_spec="$accounts_spec_in"
  accounts_spec="$(__add_val_accounts "$accounts_spec" "$val_addrs" "$val_tokens")"

  local -r wasm_script="$wasm_script_path/deploy-contracts-genesis.sh"
  verify_file_exist "$wasm_script" "wasm script file"
  source "$wasm_script"
  local treasury_addr
  treasury_addr="$(treasury_instance_addr)"
  # for PROD we decided to use the leaser's contract address(deterministic) as contracts_owner_addr which will be used to store and instantiate contracts,
  # because we would only change our contracts via gov proposals.
  # for local&&dev, we are having a normal address for contracts_owner which will be used for testing purposes
  leaser_addr=$(leaser_instance_addr)

  # use the below pattern to let the pipefail dump the failed command output
  _=$(__generate_proto_genesis_no_wasm "$chain_id" "$native_currency" "$accounts_spec" "$treasury_addr" "$leaser_addr")
  _=$(add_wasm_messages "$genesis_home_dir" "$wasm_code_path" "$contracts_owner_addr" \
                          "$treasury_init_tokens" "$lpp_native" "$contracts_info_file")

  create_validator_txs="$(__gen_val_txns "$genesis_file" "$node_id_and_val_pubkeys" "$val_stake")"
  _=$(__integrate_genesis_txs "$create_validator_txs")
  echo "$genesis_file"
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
# "vesting.amount" - mandatory number in native currency, e.g. 100 means "100 unls"
add_genesis_account() {
  local specification="$1"
  local currency="$2"
  # TBD remove the following argument once the periodic vesting testing is deleted
  local home_dir="$3"

  local address
  address=$(echo "$specification" | jq -r '.address')
  local amount
  amount=$(echo "$specification" | jq -r '.amount')
  if [[ "$(echo "$specification" | jq -r '.vesting')" != 'null' ]]; then
    local vesting_start_time=""
    if [[ "$(echo "$specification" | jq -r '.vesting."start-time"')" != 'null' ]]; then
      vesting_start_time="--vesting-start-time $(__read_unix_time "$specification" start-time)"
    fi

    local vesting_end_time
    vesting_end_time=$(echo "$specification" | jq -r '.vesting."end-time"' | __as_unix_time )
    local vesting_amount
    vesting_amount=$(echo "$specification" | jq -r '.vesting.amount')
    run_cmd "$home_dir" add-genesis-account "$address" "$amount" \
                --vesting-amount "$vesting_amount$currency" \
                --vesting-end-time "$vesting_end_time" "$vesting_start_time"
  else
    run_cmd "$home_dir" add-genesis-account "$address" "$amount"
  fi
}

#####################
# private functions #
#####################
__generate_proto_genesis_no_wasm() {
  local -r chain_id="$1"
  local -r currency="$2"
  local -r accounts_spec="$3"
  local -r treasury_addr="$4"
  local -r leaser_addr="$5"


  run_cmd "$genesis_home_dir" init genesis_manager --chain-id "$chain_id"
  run_cmd "$genesis_home_dir" config keyring-backend test
  run_cmd "$genesis_home_dir" config chain-id "$chain_id"

  __set_token_denominations "$genesis_file" "$currency"
  __set_tax_recipient "$genesis_file" "$treasury_addr"
  __set_wasm_permission_params "$genesis_file" "$contracts_owner_addr"

  while IFS= read -r account_spec ; do
    add_genesis_account "$account_spec" "$currency" "$genesis_home_dir"
  done <<< "$(echo "$accounts_spec" | jq -c '.[]')"

  # This will be used for MAINNET/TESTNET to have initial balance for the contracts_owner_addr(The leaser contract's address)
  if [ "$contracts_owner_addr" == "$leaser_addr" ]; then
    __add_bank_balances "$genesis_file" "$contracts_owner_addr" "$treasury_init_tokens_u128" "$native_currency"
  fi
}

__integrate_genesis_txs() {
  local txs="$1"

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
}

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

__set_tax_recipient() {
  local -r genesis_file="$1"
  local -r recipient_addr="$2"

  local genesis_tmp_file="$genesis_file".tmp
  < "$genesis_file" \
    jq '.app_state["tax"]["params"]["contractAddress"]="'"$recipient_addr"'"' > "$genesis_tmp_file"
  mv "$genesis_tmp_file" "$genesis_file"
}

__set_wasm_permission_params() {
  local -r genesis_file="$1"
  local -r allowed_addr="$2"

  local genesis_tmp_file="$genesis_file".tmp

  < "$genesis_file" \
    jq '.app_state["wasm"]["params"]["code_upload_access"]["permission"]="OnlyAddress"' \
    | jq '.app_state["wasm"]["params"]["code_upload_access"]["address"]="'"$allowed_addr"'"' \
    | jq '.app_state["wasm"]["params"]["instantiate_default_permission"]="Everybody"' > "$genesis_tmp_file"
  mv "$genesis_tmp_file" "$genesis_file"
}

__add_bank_balances() {
  local genesis_file="$1"
  local account_addr="$2"
  local init_tokens="$3"
  local currency="$4"

  local genesis_tmp_file="$genesis_file".tmp

  < "$genesis_file" \
    jq '.app_state["bank"]["balances"] += [{
          "address": "'"$account_addr"'",
          "coins": [
            {
              "denom": "'"$currency"'",
              "amount": "'"$init_tokens"'"
            }
          ]
        }]' > "$genesis_tmp_file"
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
  set -euo pipefail
  local node_id_and_val_pubkeys="$1"
  local destination_dir="$2"
  while IFS= read -r node_id_and_val_pubkey ; do
    local account_name
    read -r account_name __val_pub_key <<< "$node_id_and_val_pubkey"
    local address
    address=$(generate_account "$account_name" "$destination_dir")
    echo "$address"
  done <<< "$node_id_and_val_pubkeys"
}

__add_val_accounts() {
  set -euo pipefail
  local account_spec="$1"
  local val_addrs="$2"
  local val_tokens="$3"

  while IFS= read -r address ; do
    account_spec=$(echo "$account_spec" | add_account "$address" "$val_tokens")
  done <<< "$val_addrs"
  echo "$account_spec"
}

__gen_val_txns() {
  set -euo pipefail
  local proto_genesis_file="$1"
  local node_id_and_val_pubkeys="$2"
  local val_stake="$3"

  while IFS= read -r node_id_and_val_pubkey ; do
    local node_id
    local val_pub_key
    read -r node_id val_pub_key <<< "$node_id_and_val_pubkey"
    local create_validator_tx
    create_validator_tx=$(gen_val_txn "$proto_genesis_file" "$node_id" "$val_pub_key" "$val_stake")
    echo "$create_validator_tx"
  done <<< "$node_id_and_val_pubkeys"
}
