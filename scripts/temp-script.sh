#!/bin/bash
set -euxo pipefail

USER_DIR="/home/denislavivanov/.nolus"
WASM_CODE_ARTIFACTS_PATH_PLATFORM="/home/denislavivanov/go/github/nolus-money-market/artifacts-v0.7.1-dev/platform"

source "/home/denislavivanov/go/github/nolus-money-market/scripts/deploy-platform.sh"

deploy_contracts "$USER_DIR" "$WASM_CODE_ARTIFACTS_PATH_PLATFORM" "1000000000unls" "nolus1unzvj963cha6gfthjcxj5me6cmzr0w6yw3td7d" "nolus18rl557k4jcjjcx7wmcrw47wkg25e3k3vmata54"