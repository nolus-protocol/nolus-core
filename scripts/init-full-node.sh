#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR"/internal/setup-full-node.sh
source "$SCRIPT_DIR"/internal/verify.sh

MONIKER_BASE="full-node"
ARTIFACT_BIN=""
ARTIFACT_SCRIPTS=""
PERSISTENT_PEERS=""
COUNT=2
SSH_USER=""
SSH_IP=""
SSH_KEY=""

__print_usage() {
    printf "Usage: %s
    [--count <number>]
    [--artifact-bin <tar_gz_nolusd>]
    [--artifact-scripts <tar_gz_scripts>]
    [--persistent-peers <string - comma delimited list of peers>]
    [--ip <string - ip of the remote host>]
    [--user <string - ssh key user>]
    [--ssh-key <string - ssh pvt key file path>]
    [--moniker <string - node moniker (default: $MONIKER_BASE)>]" \
        "$1"
}

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

    --count)
        COUNT="$2"
        [ "$COUNT" -gt 0 ] || {
            echo >&2 "Nodes count must be a positive number"
            exit 1
        }
        shift
        shift
        ;;

    --persistent-peers)
        PERSISTENT_PEERS=$2
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

    --ssh-key)
        SSH_KEY=$2
        shift
        shift
        ;;

    --moniker)
        MONIKER_BASE="$2"
        shift
        shift
        ;;

    *)
        __print_usage "$0"
        exit 1
        ;;

    esac
done

verify_mandatory "$ARTIFACT_BIN" "Nolus binary actifact"
verify_mandatory "$ARTIFACT_SCRIPTS" "Nolus scipts actifact"
verify_mandatory "$SSH_USER" "Server ssh user"
verify_mandatory "$SSH_IP" "Server ip"
verify_mandatory "$SSH_KEY" "SSH pvt key file path"
verify_mandatory "$PERSISTENT_PEERS" "Valilidator peer IDs(comma delimited)"

init_setup_full_node "$SCRIPT_DIR" "$ARTIFACT_BIN" "$ARTIFACT_SCRIPTS" "$MONIKER_BASE" "$PERSISTENT_PEERS" "$SSH_USER" "$SSH_IP" "$SSH_KEY"
deploy_binary
deploy_scripts
setup_services "$COUNT"
setup_full_node "$COUNT"
