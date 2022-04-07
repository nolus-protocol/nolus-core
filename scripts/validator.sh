#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR"/internal/setup-validator.sh

__print_usage() {
    printf \
    "Usage: %s
    <$COMMAND_STOP|$COMMAND_SETUP|$COMMAND_SEND_GENESIS|$COMMAND_START>
    [--artifact-bin <tar_gz_nolusd>]
    [--artifact-scripts <tar_gz_scripts>]
    [--genesis-file <genesis_file_path>]
    [--ec2-id-validator <AWS EC2 validator instance ID>]
    [--ec2-private-ip-validator <AWS EC2 validator private IP>]
    [--ec2-id-sentries <space delimited AWS EC2 sentry instance IDs>]
    [--ec2-public-ip-sentries <space delimited AWS EC2 sentry public IPs>]
    [--ec2-private-ip-sentries <space delimited AWS EC2 sentry private IPs>]" \
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

__verify_mandatory_array() {
  local -r length="$1"
  local -r description="$2"
  if [[ length -eq 0 ]]; then
    echo >&2 "$description was not set"
    exit 1
  fi
}

COMMAND_STOP="stop"
COMMAND_SETUP="setup"
COMMAND_SEND_GENESIS="send-genesis"
COMMAND_START="start"

AWS_S3_ARTIFACTS_MEDIUM_BUCKET="nolus-artifact-bucket/test"
AWS_EC2_VALIDATOR_INSTANCE_ID=""
AWS_EC2_VALIDATOR_PRIVATE_IP=""
declare -g -a AWS_EC2_SENTRY_INSTANCE_IDS=()
AWS_EC2_SENTRY_PUBLIC_IPS=()
AWS_EC2_SENTRY_PRIVATE_IPS=()
# AWS_EC2_VALIDATOR_INSTANCE_ID="i-095fbcf2670dee0ea"
# AWS_EC2_VALIDATOR_PRIVATE_IP="10.215.65.198"
# AWS_EC2_SENTRY_INSTANCE_IDS=("i-07cf9474ec35f8cd7" "i-05258ca94ed55d360" "i-04e8908952a9824db")
# AWS_EC2_SENTRY_PUBLIC_IPS=("52.16.45.178" "34.241.107.0" "52.215.189.123")
# AWS_EC2_SENTRY_PRIVATE_IPS=("10.215.65.197" "10.215.65.199" "10.215.65.196")

# format: "[node-id@ip:port,]*"
KNOWN_SENTRY_NODE_URLS=","

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

  --ec2-id-validator)
    AWS_EC2_VALIDATOR_INSTANCE_ID="$2"
    shift
    shift
    ;;

  --ec2-private-ip-validator)
    AWS_EC2_VALIDATOR_PRIVATE_IP="$2"
    shift
    shift
    ;;

  --ec2-id-sentries)
    read -r -a AWS_EC2_SENTRY_INSTANCE_IDS <<< "$2"
    shift
    shift
    ;;

  --ec2-public-ip-sentries)
    read -r -a AWS_EC2_SENTRY_PUBLIC_IPS <<< "$2"
    shift
    shift
    ;;

  --ec2-private-ip-sentries)
    read -r -a AWS_EC2_SENTRY_PRIVATE_IPS <<< "$2"
    shift
    shift
    ;;

  *)
    echo "unknown option '$key'"
    exit 1
    ;;

  esac
done

__verify_mandatory "$AWS_EC2_VALIDATOR_INSTANCE_ID" "AWS EC2 validator instance ID"
__verify_mandatory_array "${#AWS_EC2_SENTRY_INSTANCE_IDS[@]}" "AWS EC2 sentry instance IDs"

if [[ "$COMMAND" == "$COMMAND_STOP" ]]; then
  stop_nodes "$SCRIPT_DIR" "$AWS_EC2_VALIDATOR_INSTANCE_ID" AWS_EC2_SENTRY_INSTANCE_IDS
elif [[ "$COMMAND" == "$COMMAND_SETUP" ]]; then
  __verify_mandatory "$ARTIFACT_BIN" "Nolus binary actifact"
  __verify_mandatory "$ARTIFACT_SCRIPTS" "Nolus scipts actifact"
  __verify_mandatory "$AWS_EC2_VALIDATOR_PRIVATE_IP" "AWS EC2 validator private IP"
  __verify_mandatory_array "${#AWS_EC2_SENTRY_PUBLIC_IPS[@]}" "AWS EC2 sentry public IPs"
  __verify_mandatory_array "${#AWS_EC2_SENTRY_PRIVATE_IPS[@]}" "AWS EC2 sentry private IPs"
  deploy_nodes "$SCRIPT_DIR" "$ARTIFACT_BIN" "$ARTIFACT_SCRIPTS" "$AWS_S3_ARTIFACTS_MEDIUM_BUCKET" \
                "$AWS_EC2_VALIDATOR_INSTANCE_ID" AWS_EC2_SENTRY_INSTANCE_IDS
  setup_nodes "$SCRIPT_DIR" "$MONIKER_BASE" "$AWS_EC2_VALIDATOR_INSTANCE_ID" \
              "$AWS_EC2_VALIDATOR_PRIVATE_IP" \
              AWS_EC2_SENTRY_INSTANCE_IDS AWS_EC2_SENTRY_PUBLIC_IPS AWS_EC2_SENTRY_PRIVATE_IPS \
              "$KNOWN_SENTRY_NODE_URLS"
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