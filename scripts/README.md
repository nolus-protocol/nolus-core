# Scripts

## Prerequisites

The binary `nolusd` must be on the system path to allow scripts to run it.

## Setup and run a localy hosted network

A script setups a network comprised of one or more validator nodes on the local file system. First it creates accounts for the validators and generates a proto genesis. Then it lets validator nodes to create validators and stake amount. Finally, the script collects the created transactions and produces the final genesis.

The genesis generation embeds the smart contracts. Therefore the nolus-money-market git repo should exist locally and be known for the network init script. By default, the scripts looks at a directory next to the this repo root. If necessary, that could be overridden providing --wasm-script-path and --wasm-code-path. For more details run
```shell
./init-local-network.sh --help
```
The nodes are ready to be started. The nolus client is configured at the default home, "$HOME/.nolus", to point to the first validator node.

Sample usage which generates 2 validator nodes:
```shell
./init-local-network.sh -v 2
```

For simplicity, and for the most use-cases, a single node local network would suffice, for example:
```shell
./init-local-network.sh
```

## Setup and run a network hosted remotely on a single machine

Script used to remotely manage network start, stop and replacing nolusd binary. Personal SSH key should be added on the remote host.

```shell
./node-operator.sh
```

This script setups a network comprised of one or more validator nodes on a remote machine.
Before executing this script make sure all services are stopped.

```shell
./init-network.sh
```

## Setup and run a network comprised of validator and sentry nodes hosted remotely on multiple machines

This setup involves multiple steps that are not orchestrated by a single main script. The aim is to allow the deployer to setup networks with different number of validators, and to have better control over the entire process.

The topology follows the Tendermint recommended model where each validator is guarded by a few sentry nodes, in our case three.  The steps are:
- stop an existing validator and its associated sentry nodes. Continue with the rest until done.
- setup a group of nodes, a validator and three sentry nodes. Continue with the rest until done.
- generate the genesis
- send the genesis to a validator and its associated sentry nodes. Continue with the rest until done.
- start a validator and its associated sentry nodes. Continue with the rest until done.

All steps are done by the commands implemented in `validator.sh` and `genesis.sh`.

### validator.sh

The script has commands for starting, setting up and stopping a group of a validator and sentry nodes.

```shell
./validator.sh --help
```

The `./validator.sh setup` outputs the fill ID and public key of the nodes. The first line contains the ones of the validator node followed by those of sentry nodes, one at a line.

### genesis.sh

```shell
./genesis.sh --help
```