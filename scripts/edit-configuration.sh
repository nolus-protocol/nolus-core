#!/bin/bash
set -euxo pipefail


POSITIONAL=()
while [[ $# -gt 0 ]]; do
  key="$1"

  case $key in
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
  --persistent_peers)
    PERSISTENT_PEERS="$2"
    shift
    shift
    ;;
  --home)
    HOME="$2"
    shift
    shift
    ;;
  *) # unknown option
    POSITIONAL+=("$1") # save it in an array for later
    shift              # past argument
    ;;
  esac
done


update_app () {
      tomlq -t "$1=$2" < "$HOME/config/app.toml" > "$HOME/config/app.toml.tmp"
      mv "$HOME/config/app.toml.tmp" "$HOME/config/app.toml"
}

update_config () {
      tomlq -t "$1=$2" < "$HOME/config/config.toml" > "$HOME/config/config.toml.tmp"
      mv "$HOME/config/config.toml.tmp" "$HOME/config/config.toml"
}


command -v tomlq > /dev/null 2>&1 || { echo >&2 "tomlq not installed. More info: https://tomlq.readthedocs.io/en/latest/installation.html"; exit 1; }


if [[ -z "$HOME" ]]; then
  echo "HOME is unset"
   exit 1
fi

if [[ -n "${MINIMUM_GAS_PRICES+x}" ]]; then
  update_app '."minimum-gas-prices"' "\"$MINIMUM_GAS_PRICES\""
fi

if [[ -n "${ENABLE_API+x}" ]]; then
  update_app '."api"."enable"' "$ENABLE_API"
fi

if [[ -n "${API_ADDRESS+x}" ]]; then
  update_app '."api"."address"' "\"$API_ADDRESS"
fi

if [[ -n "${ENABLE_GRPC+x}" ]]; then
  update_app '."grpc"."enable"' "$ENABLE_GRPC"
fi

if [[ -n "${GRPC_ADDRESS+x}" ]]; then
  update_app '."grpc"."address"' "\"$GRPC_ADDRESS\""
fi

if [[ -n "${ENABLE_GRPC_WEB+x}" ]]; then
  update_app '."grpc-web"."enable"' "$ENABLE_GRPC_WEB"
fi

if [[ -n "${GRPC_WEB_ADDRESS+x}" ]]; then
  update_app '."grpc-web"."address"' "\"$GRPC_WEB_ADDRESS\""
fi



if [[ -n "${PROXY_APP_ADDRESS+x}" ]]; then
  update_config '."proxy_app"' "\"$PROXY_APP_ADDRESS\""
fi

if [[ -n "${TENDERMINT_RPC_ADDRESS+x}" ]]; then
  update_config '."rpc"."laddr"' "\"$TENDERMINT_RPC_ADDRESS\""
fi

if [[ -n "${TENDERMINT_P2P_ADDRESS+x}" ]]; then
  update_config '."p2p"."laddr"' "\"$TENDERMINT_P2P_ADDRESS\""
fi

if [[ -n "${PERSISTENT_PEERS+x}" ]]; then
  update_config '."p2p"."laddr"' "\"$PERSISTENT_PEERS\""
fi
