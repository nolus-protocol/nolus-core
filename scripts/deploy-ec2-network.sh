#!/bin/bash

rm -rf ./validator_setup

./init-test-network.sh -v 3 -ips '172.28.5.2,172.28.5.3,172.28.5.4'

# make validator 1 apis available to the outside world
./remote/edit.sh --home ./validator_setup/node1 --enable-api true --enable-grpc true --enable-grpc-web true --tendermint-rpc-address "tcp://0.0.0.0:26657"

ssh -i ./gitlab.pem ec2-user@ec2-35-158-128-53.eu-central-1.compute.amazonaws.com 'sudo rm -rf /tmp/validator_setup'
scp -i ./gitlab.pem -r ./validator_setup ec2-user@ec2-35-158-128-53.eu-central-1.compute.amazonaws.com:/tmp

ssh -i ./gitlab.pem ec2-user@ec2-35-158-128-53.eu-central-1.compute.amazonaws.com << EOF
  if [ "$(docker network list | grep 'blockchain-network' || echo 'none')" = 'none' ]; then
    docker network create --subnet=172.28.0.0/16 --ip-range=172.28.5.0/24 --gateway=172.28.5.1  blockchain-network
  fi
  docker kill $(docker ps -q)
  docker run -d --network=blockchain-network -v "/tmp/validator_setup/node1:/root/.cosmzone/:Z" -p 1317:1317 -p 9090:9090 -p 9091:9091 -p 26657:26657 --ip=172.28.5.2 public.ecr.aws/nolus/node:0.1 start
  docker run -d --network=blockchain-network -v "/tmp/validator_setup/node2:/root/.cosmzone/:Z" --ip=172.28.5.3 public.ecr.aws/nolus/node:0.1 start
  docker run -d --network=blockchain-network -v "/tmp/validator_setup/node3:/root/.cosmzone/:Z" --ip=172.28.5.4 public.ecr.aws/nolus/node:0.1 start
EOF
