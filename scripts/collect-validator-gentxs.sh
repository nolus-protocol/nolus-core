#!/bin/bash
set -euxo pipefail

COLLECTOR_DIR=""
GENTXS_FILES_DIR=""
MODE="local"

POSITIONAL=()
while [[ $# -gt 0 ]]; do
  key="$1"

  case $key in
  -m | --mode)
    MODE="$2"
    [[ "$MODE" == "local" || "$MODE" == "docker" ]] || {
      echo >&2 "mode must be either local or docker"
      exit 1
    }
    shift
    shift
    ;;
  --collector)
    COLLECTOR_DIR="$2"
    shift
    shift
    ;;
  --gentxs)
    GENTXS_FILES_DIR="$2"
    shift
    shift
    ;;
  --help)
    echo "Usage: collect-validator-gentxs.sh [--collector <collector_idr>] [-gentxs <gentx_dir>] [-m|--mode <local|docker>]"
    exit 0
    ;;
  *) # unknown option
    POSITIONAL+=("$1") # save it in an array for later
    shift              # past argument
    ;;
  esac
done

run_cmd() {
  local DIR="$1"
  shift
  case $MODE in
  local) cosmzoned $@ --home "$DIR" ;;
  docker) docker run --rm -u "$(id -u)":"$(id -u)" -v "$DIR:/tmp/.cosmzone:Z" nomo/node $@ --home /tmp/.cosmzone ;;
  esac
}

if [[ ! -d "$COLLECTOR_DIR" ]]; then
  echo "collector node directory does not exist"
   exit 1
fi


if [[ ! -d "$GENTXS_FILES_DIR" ]]; then
  echo "gentx directory does not exist"
   exit 1
fi

rm -rf "$COLLECTOR_DIR/config/gentx"
mkdir "$COLLECTOR_DIR/config/gentx"
cp -a "$GENTXS_FILES_DIR/." "$COLLECTOR_DIR/config/gentx"

run_cmd "$COLLECTOR_DIR" collect-gentxs