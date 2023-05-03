#!/bin/bash
# Used to copy files between localhost and remote server via ssh
#
# arg1: file path to copy
# arg2: target path to paste
# arg3: ssh user
# arg4: server ip
set -euo pipefail

source="$1"
target="$2"
ssh_user="$3"
ssh_ip="$4"

scp $source $ssh_user@$ssh_ip:$target
