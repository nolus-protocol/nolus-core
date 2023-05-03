#!/bin/bash
set -euxo pipefail

SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
VALIDATORS=1
ACTION=""
SSH_USER=""
SSH_IP=""
ARTIFACT_BIN="nolus.tar.gz"

cli_help() {
  cli_name=${0##*/}
  echo "$cli_name
Dev operator CLI
Usage: $cli_name [flags]

Available Flags:
  [--validators <number>]
  [--action <start|stop|replace - start or stop all validators or replace nolusd bin on the remote host>]
  [--ip <string - ip of the remote host>]
  [--user <string - ssh key user>]
  [--artifact-bin <*.tar.gz - archive with nolusd bin>]
"
  exit 1
}

while [[ $# -gt 0 ]]; do
  key="$1"

  case $key in
  --action)
    ACTION=$2
    shift
    shift
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

  --validators)
    VALIDATORS="$2"
    [ "$VALIDATORS" -gt 0 ] || {
      echo >&2 "validators must be a positive number"
      exit 1
    }
    shift
    shift
    ;;

  --artifact-bin)
    ARTIFACT_BIN="$2"
    shift
    shift
    ;;

  *)
    cli_help
    ;;
  esac
done

source "$SCRIPT_DIR"/internal/setup-validator-dev.sh
init_setup_validator_dev_sh $SCRIPT_DIR $ARTIFACT_BIN "" $SSH_USER $SSH_IP

case $ACTION in
"stop")
  stop_validators $VALIDATORS
  ;;

"start")
  start_validators $VALIDATORS
  ;;

"replace")
  deploy_binary
  ;;

*)
  echo "Invalid action"
  ;;
esac
