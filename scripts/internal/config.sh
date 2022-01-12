#!/bin/bash
set -euxo pipefail

command -v tomlq > /dev/null 2>&1 || {
  echo >&2 "tomlq not installed. More info: https://tomlq.readthedocs.io/en/latest/installation.html"
  exit 1
}

update_app() {
  local app_file="$1"/config/app.toml

  tomlq -t "$2=$3" < "$app_file" > "$app_file".tmp
  mv "$app_file".tmp "$app_file"
}

update_config() {
  local config_file="$1"/config/config.toml

  tomlq -t "$2=$3" < "$config_file" > "$config_file".tmp
  mv "$config_file".tmp "$config_file"
}
