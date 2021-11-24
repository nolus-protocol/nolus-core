#!/bin/bash
set -euxo pipefail

run_cmd() {
  local mode
  mode="$1"
  shift
  local dir
  dir="$1"
  shift
  case $mode in
  local) cosmzoned $@ --home "$dir" 2>&1 ;;
  docker) docker run --rm -u "$(id -u)":"$(id -u)" -v "$dir:/tmp/.cosmzone:Z" nomo/node $@ --home /tmp/.cosmzone 2>&1 ;;
  esac
}

