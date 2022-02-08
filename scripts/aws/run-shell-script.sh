#!/bin/bash
set -euxo pipefail

cmd="$1"
instance_id="$2"

aws_cmd=$(aws ssm send-command --document-name "AWS-RunShellScript" --document-version "1" --targets '[{"Key":"InstanceIds","Values":["'"$instance_id"'"]}]' --parameters '{"workingDirectory":[""],"executionTimeout":["3600"],"commands":["'"$cmd"'"]}' --timeout-seconds 600 --max-concurrency "50" --max-errors "0" --region eu-west-1 --query 'Command.{CommandId:CommandId}' --output text )

aws ssm wait command-executed  --command-id "$aws_cmd" --instance-id "$instance_id"

cmd_invocations=$(aws ssm list-command-invocations --command-id "$aws_cmd" --output json --details)

echo "$cmd_invocations" | jq -r '.CommandInvocations[].CommandPlugins[].Output'

[[ $( echo "$cmd_invocations" | jq '.CommandInvocations[].Status' | xargs ) == Success ]]