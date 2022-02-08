#!/bin/bash
set -euxo pipefail

SCRIPT_DIR=$(cd $(dirname "${BASH_SOURCE[0]}") && pwd)
source "$SCRIPT_DIR"/common/cmd.sh

COLLECTOR_DIR=""
GENTXS_FILES_DIR=""

POSITIONAL=()
while [[ $# -gt 0 ]]; do
  key="$1"

  case $key in
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
    echo "Usage: collect-validator-gentxs.sh [--collector <collector_idr>] [-gentxs <gentx_dir>]"
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

run_cmd "$COLLECTOR_DIR" collect-gentxs