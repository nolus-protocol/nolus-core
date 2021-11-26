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
  docker) docker run -i --rm -u "$(id -u)":"$(id -u)" -v "$(realpath "$dir"):/tmp/.cosmzone:Z" public.ecr.aws/nolus/node:0.1 $@ --home /tmp/.cosmzone 2>&1 ;;
  esac
}

