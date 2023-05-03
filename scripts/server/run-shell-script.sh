#!/bin/bash
# Used to execute commands on a remote server via ssh
#
# arg1: command to be executed
# arg3: ssh user
# arg4: server ip
set -euo pipefail

cmd="$1"
ssh_user="$2"
ssh_ip="$3"

ssh $ssh_user@$ssh_ip $cmd
