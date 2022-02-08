#!/bin/bash
SCRIPT_DIR=$(cd $(dirname "${BASH_SOURCE[0]}") && pwd)
source "$SCRIPT_DIR"/lib/lib.sh

POSITIONAL=()
while [[ $# -gt 0 ]]; do
  key="$1"

  case $key in
    --help)
      printf \
      "Usage: %s
      [-gas|--minimum-gas-prices <minimum_gas_prices>]
      [--enable-api <true|false>]
      [--api-address <ip_address>]
      [--enable-grpc <true|false>]
      [--grpc-address <ip_address>]
      [--enable-grpc-web <true|false>]
      [--grpc-web-address <ip_address>]
      [--proxy-app-address <ip_address>]
      [--tendermint-rpc-address <ip_address>]
      [--tendermint-p2p-address <ip_address>]
      [--persistent-peers <node_addresses>]
      [--timeout-commit <duration>]
      [--home <full_node_dir>]" "$0"
      exit 0
      ;;
  -gas | --minimum-gas-prices)
    MINIMUM_GAS_PRICES="$2"
    shift # past argument
    shift # past value
    ;;
  --enable-api)
    ENABLE_API="$2"
    shift
    shift
    ;;
  --api-address)
    API_ADDRESS="$2"
    shift
    shift
    ;;
  --enable-grpc)
    ENABLE_GRPC="$2"
    shift
    shift
    ;;
  --grpc-address)
    GRPC_ADDRESS="$2"
    shift
    shift
    ;;
  --enable-grpc-web)
    ENABLE_GRPC_WEB="$2"
    shift
    shift
    ;;
  --grpc-web-address)
    GRPC_WEB_ADDRESS="$2"
    shift
    shift
    ;;
  --proxy-app-address)
    PROXY_APP_ADDRESS="$2"
    shift
    shift
    ;;
  --tendermint-rpc-address)
    TENDERMINT_RPC_ADDRESS="$2"
    shift
    shift
    ;;
  --tendermint-p2p-address)
    TENDERMINT_P2P_ADDRESS="$2"
    shift
    shift
    ;;
  --persistent-peers)
    PERSISTENT_PEERS="$2"
    shift
    shift
    ;;
  --timeout-commit)
    TIMEOUT_COMMIT="$2"
    shift
    shift
    ;;
  --home)
    HOME_DIR="$2"
    shift
    shift
    ;;
  *) # unknown option
    POSITIONAL+=("$1") # save it in an array for later
    shift              # past argument
    ;;
  esac
done

if [[ -z "$HOME_DIR" ]]; then
  echo "HOME_DIR is unset"
   exit 1
fi

update_app_home() {
  update_app "$HOME_DIR" "$@"
}

update_config_home() {
  update_config "$HOME_DIR" "$@"
}

if [[ -n "${MINIMUM_GAS_PRICES+x}" ]]; then
  update_app_home '."minimum-gas-prices"' "\"$MINIMUM_GAS_PRICES\""
fi

if [[ -n "${ENABLE_API+x}" ]]; then
  update_app_home '."api"."enable"' "$ENABLE_API"
fi

if [[ -n "${API_ADDRESS+x}" ]]; then
  update_app_home '."api"."address"' "\"$API_ADDRESS\""
fi

if [[ -n "${ENABLE_GRPC+x}" ]]; then
  update_app_home '."grpc"."enable"' "$ENABLE_GRPC"
fi

if [[ -n "${GRPC_ADDRESS+x}" ]]; then
  update_app_home '."grpc"."address"' "\"$GRPC_ADDRESS\""
fi

if [[ -n "${ENABLE_GRPC_WEB+x}" ]]; then
  update_app_home '."grpc-web"."enable"' "$ENABLE_GRPC_WEB"
fi

if [[ -n "${GRPC_WEB_ADDRESS+x}" ]]; then
  update_app_home '."grpc-web"."address"' "\"$GRPC_WEB_ADDRESS\""
fi



if [[ -n "${PROXY_APP_ADDRESS+x}" ]]; then
  update_config_home '."proxy_app"' "\"$PROXY_APP_ADDRESS\""
fi

if [[ -n "${TENDERMINT_RPC_ADDRESS+x}" ]]; then
  update_config_home '."rpc"."laddr"' "\"$TENDERMINT_RPC_ADDRESS\""
fi

if [[ -n "${TENDERMINT_P2P_ADDRESS+x}" ]]; then
  update_config_home '."p2p"."laddr"' "\"$TENDERMINT_P2P_ADDRESS\""
fi

if [[ -n "${PERSISTENT_PEERS+x}" ]]; then
  update_config_home '."p2p"."laddr"' "\"$PERSISTENT_PEERS\""
fi

if [[ -n "${TIMEOUT_COMMIT+x}" ]]; then
  update_config_home '."consensus"."timeout_commit"'  "\"$TIMEOUT_COMMIT\""
fi
