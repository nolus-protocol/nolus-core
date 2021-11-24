#!/bin/bash
set -euxo pipefail


command -v common-util.sh >/dev/null 2>&1 || {
  echo >&2 "scripts are not found in \$PATH."
  exit 1
}

source common-util.sh

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

run_cmd "$MODE" "$COLLECTOR_DIR" collect-gentxs