#!/bin/bash

# start "instance" variables
admin_dev_home_dir=""
# end "instance" variables

init_admin_dev_sh() {
  admin_dev_home_dir="$1"
  local scripts_home_dir="$2"
  source "$scripts_home_dir"/common/cmd.sh
}

admin_dev_create_treasury_account() {
  __admin_dev_create_account "treasury"
}

__admin_dev_create_account() {
  local name="$1"

  run_cmd "$admin_dev_home_dir" keys add "$name" --keyring-backend "test" >/dev/null
  run_cmd "$admin_dev_home_dir" keys show "$name" -a --keyring-backend "test"
}