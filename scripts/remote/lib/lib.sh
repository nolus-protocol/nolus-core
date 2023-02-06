#!/bin/bash
command -v tomlq > /dev/null 2>&1 || {
  echo >&2 "tomlq not installed. More info: https://tomlq.readthedocs.io/en/latest/installation.html"
  exit 1
}

update_app() {
  local app_file="$1"/app.toml

  _update_toml "$app_file" "$2" "$3"
}

update_config() {
  local config_file="$1"/config.toml

  _update_toml "$config_file" "$2" "$3"
}

update_client() {
  local config_file="$1"/client.toml

  _update_toml "$config_file" "$2" "$3"
}

_update_toml() {
  local file="$1"

  tomlq -t "$2=$3" < "$file" > "$file".tmp
  mv "$file".tmp "$file"
}
