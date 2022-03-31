#!/bin/bash
set -euox pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR"/internal/setup-validator.sh

__print_usage() {
    printf \
    "Usage: %s
    <$COMMAND_STOP|$COMMAND_SETUP|$COMMAND_SEND_GENESIS|$COMMAND_START>
    [--artifact-bin <tar_gz_nolusd>]
    [--artifact-scripts <tar_gz_scripts>]
    [--genesis-file <genesis_file_path>]" \
     "$1"
}

__verify_mandatory() {
  local value="$1"
  local description="$2"

  if [[ -z "$value" ]]; then
    echo >&2 "$description was not set"
    exit 1
  fi
}

COMMAND_STOP="stop"
COMMAND_SETUP="setup"
COMMAND_SEND_GENESIS="send-genesis"
COMMAND_START="start"

AWS_S3_ARTIFACTS_MEDIUM_BUCKET="nolus-artifact-bucket/test"
AWS_EC2_VALIDATOR_INSTANCE_ID="i-095fbcf2670dee0ea"
AWS_EC2_VALIDATOR_PRIVATE_IP="10.215.65.198"
AWS_EC2_SENTRY_INSTANCE_IDS=("i-07cf9474ec35f8cd7" "i-05258ca94ed55d360" "i-04e8908952a9824db")
AWS_EC2_SENTRY_PUBLIC_IPS=("52.16.45.178" "34.241.107.0" "52.215.189.123")

MONIKER_BASE="rila1"
ARTIFACT_BIN=""
ARTIFACT_SCRIPTS=""
GENESIS_FILE=""

if [[ $# -lt 1 ]]; then
  echo "Missing command!"
  __print_usage "$0"
  exit 1
fi
COMMAND="$1"
shift

while [[ $# -gt 0 ]]; do
  key="$1"

  case $key in

  -h | --help)
    __print_usage "$0"
    exit 0
    ;;

   --artifact-bin)
    ARTIFACT_BIN="$2"
    shift
    shift
    ;;

   --artifact-scripts)
    ARTIFACT_SCRIPTS="$2"
    shift
    shift
    ;;

  --genesis-file)
    GENESIS_FILE="$2"
    shift
    shift
    ;;
  
  *)
    echo "unknown option '$key'"
    exit 1
    ;;

  esac
done

if [[ "$COMMAND" == "$COMMAND_STOP" ]]; then
  stop_nodes "$SCRIPT_DIR" "$AWS_EC2_VALIDATOR_INSTANCE_ID" AWS_EC2_SENTRY_INSTANCE_IDS
elif [[ "$COMMAND" == "$COMMAND_SETUP" ]]; then
  __verify_mandatory "$ARTIFACT_BIN" "Nolus binary actifact"
  __verify_mandatory "$ARTIFACT_SCRIPTS" "Nolus scipts actifact"
  deploy_nodes "$SCRIPT_DIR" "$ARTIFACT_BIN" "$ARTIFACT_SCRIPTS" "$AWS_S3_ARTIFACTS_MEDIUM_BUCKET" \
                "$AWS_EC2_VALIDATOR_INSTANCE_ID" AWS_EC2_SENTRY_INSTANCE_IDS
  setup_nodes "$SCRIPT_DIR" "$MONIKER_BASE" "$AWS_EC2_VALIDATOR_INSTANCE_ID" \
              "$AWS_EC2_VALIDATOR_PRIVATE_IP" \
              AWS_EC2_SENTRY_INSTANCE_IDS AWS_EC2_SENTRY_PUBLIC_IPS
elif [[ "$COMMAND" == "$COMMAND_SEND_GENESIS" ]]; then
  __verify_mandatory "$GENESIS_FILE" "Nolus genesis file"
  propagate_genesis "$SCRIPT_DIR" "$GENESIS_FILE" "$AWS_S3_ARTIFACTS_MEDIUM_BUCKET" \
                    "$AWS_EC2_VALIDATOR_INSTANCE_ID" AWS_EC2_SENTRY_INSTANCE_IDS
elif [[ "$COMMAND" == "$COMMAND_START" ]]; then
  start_nodes "$SCRIPT_DIR" "$AWS_EC2_VALIDATOR_INSTANCE_ID" AWS_EC2_SENTRY_INSTANCE_IDS
else
  echo "Unknown command!"
  exit 1
fi