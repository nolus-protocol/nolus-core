#!/bin/bash
set -euxo pipefail

generate_proto_genesis() {
  local genesis_home_dir="$1"
  local chain_id="$2"
  local accounts_file="$3"
  local currency="$4"
  local proto_genesis_file="$5"

  run_cmd "$genesis_home_dir" init genesis_manager --chain-id "$chain_id"
  run_cmd "$genesis_home_dir" config keyring-backend test
  run_cmd "$genesis_home_dir" config chain-id "$chain_id"

  local genesis_file="$genesis_home_dir/config/genesis.json"
  fix_token_denominations "$genesis_file" "$currency"

  if [[ -n "${accounts_file+x}" ]]; then
    for i in $(jq '. | keys | .[]' "$accounts_file"); do
      row=$(jq ".[$i]" "$accounts_file")
      address=$(jq -r '.address' <<<"$row")
      amount=$(jq -r '.amount' <<<"$row")
      if [[ "$(jq -r '.vesting' <<<"$row")" != 'null' ]]; then
        add_vesting_account "$row" "$genesis_home_dir"
      else
        run_cmd "$genesis_home_dir" add-genesis-account "$address" "$amount"
      fi
    done
  fi

  cp "$genesis_file" "$proto_genesis_file"
}

integrate_genesis_txs() {
  local genesis_home_dir="$1"
  local genesis_in_file="$2"
  local txs="$3"
  local genesis_out_file="$4"

  local genesis_basedir="$genesis_home_dir"/config
  local genesis_file="$genesis_basedir"/genesis.json
  cp "$genesis_in_file" "$genesis_file"

  local txs_dir="$genesis_home_dir"/txs
  {
    mkdir "$txs_dir"
    local index=0
    for tx in $txs; do
        echo "$tx" > "$txs_dir"/tx"$index".json
        index=$((index+1))
    done
  }

  run_cmd "$genesis_home_dir" collect-gentxs --gentx-dir "$txs_dir"
  cp "$genesis_file" "$genesis_out_file"
}

fix_token_denominations() {
  local genesis_file="$1"
  local currency="$2"

  local genesis_tmp_file="$genesis_file".tmp

  cat "$genesis_file" \
    | jq '.app_state["staking"]["params"]["bond_denom"]="'"$currency"'"' \
    | jq '.app_state["crisis"]["constant_fee"]["denom"]="'"$currency"'"' \
    | jq '.app_state["gov"]["deposit_params"]["min_deposit"][0]["denom"]="'"$currency"'"' \
    | jq '.app_state["gov"]["deposit_params"]["min_deposit"][0]["denom"]="'"$currency"'"' \
    | jq '.app_state["mint"]["params"]["mint_denom"]="'"$currency"'"' > "$genesis_tmp_file"
  mv "$genesis_tmp_file" "$genesis_file"
}