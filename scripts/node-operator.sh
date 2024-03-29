#!/bin/bash
set -euo pipefail

SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
source "$SCRIPT_DIR"/internal/verify.sh

MONIKER_BASE=""
NODES=1
SSH_USER=""
SSH_IP=""
ARTIFACT_BIN="nolus.tar.gz"
BIN_TARGET_DIR=""
GENESIS_FILE=""
SSH_KEY=""

COMMAND_STOP="stop"
COMMAND_START="start"
COMMAND_SEND_GENESIS="send-genesis"
COMMAND_REPLACE_BIN="replace-bin"

cli_help() {
  cli_name=${0##*/}
  echo "$cli_name
Node operator CLI
Usage: $cli_name <command> [flags]

Available commands:
  <$COMMAND_STOP | $COMMAND_START | $COMMAND_SEND_GENESIS | $COMMAND_REPLACE_BIN>

Available Flags:
  [-h | --help]
  [--nodes <number - nodes count>]
  [--ip <string - ip of the remote host>]
  [--user <string - ssh user>]
  [--ssh-key <string - ssh pvt key file path>]
  [--genesis-file <genesis_file_path>]
  [--artifact-bin <*.tar.gz - archive with nolusd bin>]
  [--bin-target-dir <full path on the remote>]
  [--moniker <string - node moniker>]
"
  exit 1
}

if [[ $# -lt 1 ]]; then
  cli_help
fi
COMMAND="$1"
shift

while [[ $# -gt 0 ]]; do
  key="$1"

  case $key in
  -h | --help)
    cli_help
    ;;

  --ip)
    SSH_IP=$2
    shift
    shift
    ;;

  --user)
    SSH_USER=$2
    shift
    shift
    ;;

  --ssh-key)
    SSH_KEY=$2
    shift
    shift
    ;;

  --nodes)
    NODES="$2"
    [ "$NODES" -gt 0 ] || {
      echo >&2 "nodes must be a positive number"
      exit 1
    }
    shift
    shift
    ;;

  --genesis-file)
    GENESIS_FILE="$2"
    shift
    shift
    ;;

  --artifact-bin)
    ARTIFACT_BIN="$2"
    shift
    shift
    ;;

  --bin-target-dir)
    BIN_TARGET_DIR="$2"
    shift
    shift
    ;;

  --moniker)
    MONIKER_BASE="$2"
    shift
    shift
    ;;

  *)
    cli_help
    ;;
  esac
done

source "$SCRIPT_DIR"/internal/setup-validator.sh
verify_mandatory "$SSH_USER" "Remote server SSH user"
verify_mandatory "$SSH_IP" "Remote server IP"
verify_mandatory "$SSH_KEY" "SSH pvt key file path"

init_setup_validator "$SCRIPT_DIR" "$ARTIFACT_BIN" "" "$MONIKER_BASE" "$SSH_USER" "$SSH_IP" "$SSH_KEY"

case $COMMAND in
$COMMAND_STOP)
  verify_mandatory "$MONIKER_BASE" "Node moniker"
  stop_validators $NODES
  ;;

$COMMAND_START)
  verify_mandatory "$MONIKER_BASE" "Node moniker"
  start_validators $NODES
  ;;

$COMMAND_REPLACE_BIN)
  verify_mandatory "$ARTIFACT_BIN" "Nolus binary actifact"
  deploy_binary $BIN_TARGET_DIR
  ;;

$COMMAND_SEND_GENESIS)
  verify_mandatory "$GENESIS_FILE" "Nolus genesis file"
  propagate_genesis "$GENESIS_FILE" "$NODES"
  ;;
*)
  echo "Invalid command"
  ;;
esac
