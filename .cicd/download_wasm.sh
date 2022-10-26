#!/bin/bash
set -euox pipefail

version="$1"

TOKEN_TYPE="JOB-TOKEN"
TOKEN_VALUE="$CI_JOB_TOKEN"
GITLAB_API="https://gitlab-nomo.credissimo.net/api/v4"
# nolus-money-market project id obtained from $GITLAB_API/projects
PROJECT_ID=8
JOB_NAME="build-and-optimization:cargo"
ARCHVE_NAME="$JOB_NAME.zip"

curl --output "$ARCHVE_NAME" --header "$TOKEN_TYPE: $TOKEN_VALUE" \
        "$GITLAB_API/projects/$PROJECT_ID/jobs/artifacts/$version/download?job=$JOB_NAME"
echo 'A' | unzip "$ARCHVE_NAME"
