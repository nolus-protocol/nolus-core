# Scripts

## Prerequisites

The binary `nolusd` must be on the system path to allow scripts to run it.

## Setup and run a locally hosted network

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

This script setups a network comprised of one or more validator nodes on a remote machine.
Before executing this script make sure all services are stopped.

```shell
./init-network.sh
```

## Setup full nodes hosted remotely on single machine
Steps:
- stop existing nodes
- setup full nodes
- generate the genesis
- send the genesis to full nodes
- start nodes

All steps are done by the commands implemented in `init-full-node.sh` and `node-operator.sh`.

### node-operator.sh

Script used to remotely manage network. Personal SSH key should be added on the remote host.

```shell
./node-operator.sh
```

### init-full-node.sh

The script deploys and setups full nodes.

```shell
./init-full-node.sh
```

### genesis.sh

Script used to generate genesis file.

```shell
./genesis.sh
```