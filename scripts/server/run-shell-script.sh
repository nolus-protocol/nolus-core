#!/bin/bash
# Used to execute commands on a remote server via ssh
#
# arg1: command to be executed
# arg2: ssh user
# arg3: server ip
# arg4: ssh private key file path
set -euo pipefail

cmd="$1"
ssh_user="$2"
ssh_ip="$3"
ssh_key="$4"

ssh -o "IdentitiesOnly=yes" -i "$ssh_key" "$ssh_user"@"$ssh_ip" "$cmd"
