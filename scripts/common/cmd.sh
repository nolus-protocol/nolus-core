#!/bin/bash
CMD="nolusd"
command -v "$CMD" >/dev/null 2>&1 || {
  echo >&2 "$CMD is not found in \$PATH."
  exit 1
}

run_cmd() {
  local home="$1"
  shift

  "$CMD" "$@" --home "$home" 2>&1
}


