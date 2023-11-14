# Nolus Core
<div align="center">
<br /><p align="center"><img alt="nolus-core" src="docs/nolus-core-logo.svg" width="100"/></p><br />


[![Go Report Card](https://goreportcard.com/badge/github.com/Nolus-Protocol/nolus-core)](https://goreportcard.com/report/github.com/Nolus-Protocol/nolus-core)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://github.com/Nolus-Protocol/nolus-core/blob/main/LICENSE)
[![Lint](https://github.com/Nolus-Protocol/nolus-core/actions/workflows/lint.yaml/badge.svg?branch=main)](https://github.com/Nolus-Protocol/nolus-core/actions/workflows/lint.yaml)
[![Test](https://github.com/Nolus-Protocol/nolus-core/actions/workflows/test.yaml/badge.svg?branch=main)](https://github.com/Nolus-Protocol/nolus-core/actions/workflows/test.yaml)
[![Test Fuzz](https://github.com/Nolus-Protocol/nolus-core/actions/workflows/test-fuzz.yaml/badge.svg?branch=main)](https://github.com/Nolus-Protocol/nolus-core/actions/workflows/test-fuzz.yaml)
[![Test cosmos-sdk](https://github.com/Nolus-Protocol/nolus-core/actions/workflows/test-cosmos.yaml/badge.svg?branch=main)](https://github.com/Nolus-Protocol/nolus-core/actions/workflows/test-cosmos.yaml)

**Nolus** is a blockchain built using Cosmos SDK and Tendermint.
</div>

## Prerequisites

Install [golang](https://golang.org/), [tomlq](https://tomlq.readthedocs.io/en/latest/installation.html) and [jq](https://stedolan.github.io/jq/).

## Get started

### Build, configure and run a single-node locally deployed Nolus chain

#### Build

  ```sh
  make install
  ```

The command will compile and install nolusd locally on your machine.

#### Initialize, set up the DEX parameters and run

`init-local-network.sh` generates a network setup, including setting up the initial DEX. The Hermes relayer is used to connect to the DEXes, and its configuration is also handled by the script.

First, generate the mnemonic you will use for Hermes:

```sh
nolusd keys mnemonic
```

Then recover it on the DEX (the network binary is required) and use a faucet to obtain some amount:

Example for the Osmosis DEX ([Osmo-test-5 faucet](https://faucet.osmotest5.osmosis.zone/)):

```sh
osmosisd keys add hermes_key --recover
```

Initialize and start (run `./scripts/init-local-network.sh --help` for additional configuration options):

```sh
./scripts/init-local-network.sh --reserve-tokens <reserve_account_init_tokens> --hermes-mnemonic <the_mnemonic_generated_by_the_previous_steps> --dex-network-addr-rpc-host <dex_network_addr_rpc_host_part> --dex-network-addr-grpc-host <dex_network_addr_grpc_host_part> --dex-admin-mnemonic <mnemonic_phrase> --store-code-privileged-account-mnemonic <mnemonic_phrase>
```

Set up the DEX parameters: [Set up the DEX parameters manually](#set-up-the-dex-parameters-manually)

*Notes:

* Make sure the `nolus-money-market` repository is checked out as a sibling to this one.

* !!! Before running the `./scripts/init-local-network.sh` again, make sure the `nolusd` and `hermes` processes are killed.

* The `hermes` and `nolusd` logs are stored in `~/hermes` and `~/.nolus` respectively.

### Run an already configured single-node

```sh
nolusd start --home "networks/nolus/local-validator-1"
```

## Set up the DEX parameters manually

The goal is to let smart contracts know the details of the connectivity to the selected DEX. Herebelow is a sample request for setting up the DEX.
This should be done via sudo gov proposal:

```sh
nolusd tx gov submit-proposal sudo-contract nolus1wn625s4jcmvk0szpl85rj5azkfc6suyvf75q6vrddscjdphtve8s5gg42f '{"setup_dex": {"connection_id": "connection-0", "transfer_channel": {"local_endpoint": "channel-0", "remote_endpoint": "channel-1499"}}}' --title "Set up the DEX parameter" --description "Thе proposal aims to set the DEX parameters in the Leaser contract" --deposit 10000000unls --fees 900unls --gas auto --gas-adjustment 1.1 --from wallet
```

Check if the transaction has passed:

```sh
nolusd q wasm contract-state smart nolus1wn625s4jcmvk0szpl85rj5azkfc6suyvf75q6vrddscjdphtve8s5gg42f '{"config":{}}'
```

*Notes:

* `*nolus1wn625s4jcmvk0szpl85rj5azkfc6suyvf75q6vrddscjdphtve8s5gg42f*` is the Leaser contract instance associated with the first DEX already configured on the network at genesis time /by default this is Osmosis/. Each DEX is associated with a separate instance.

* `*connection-0*` is the connection to the first DEX already configured on the network /by default this is Osmosis/. Should be replaced with the connection to the selected DEX.

* `*channel-0*` refers to the first DEX already configured on the network /by default this is Osmosis/. Should be replaced with the channel to the selected DEX.

* `*channel-1499*` should be replaced, so you can get the actual channel ID of the remote endpoint with:

```sh
nolusd q ibc channel connections <connection> --output json | jq '.channels[0].counterparty.channel_id' | tr -d '"'
```

## Set up a new DEX

On a live network, a new DEX can be deployed using the following script. It takes care of setting up Hermes to work with the new DEX. (TO DO: It will also deploy DEX-specific contracts using `./deploy-contracts-live.sh`)

```sh
./scripts/add-new-dex.sh --dex-chain-id <new_dex_chain_id> --dex-ip-addr-rpc-host <new_dex_ip_addr_rpc_host_part> --dex-ip-addr-grpc-host <new_dex_ip_addr_grpc_host_part> --dex-account-prefix <new_dex_account_prefix> --dex-price-denom <new_dex_price_denom> --dex-trusting-period-secs <new_dex_trusting_period_in_seconds>
```

(Execute `./scripts/add-new-dex.sh --help` for additional configuration options)

The script will locate the Hermes account from the Hermes configuration directory and link it to the new DEX.

!!! Prerequisites: Before running, the address should have a certain amount on the DEX network in order to be used by Hermes. This can be accomplished by using the DEX network binary and a public faucet, as demonstrated for Osmosis [here](#initialize-set-up-the-DEX-parameters-and-run).

## Build a statically linked binary

By default, `make build` generates a dynamically linked binary. In case someone would like to reproduce the way the binary is built in the pipeline then the command to achieve it locally is:

```sh
docker run --rm -it -v "$(pwd)":/code public.ecr.aws/nolus/builder:<replace_with_the latest_tag> make build -C /code
```

## Upgrade wasmvm

* Update the Go modules
* Update the wasmvm version in the builder Dockerfile at .github/images/builder.Dockerfile
* Increment the IMAGE_TAG and use the same version in the build-binary step in .github/workflows/build.yaml

## Run a full node with docker

[https://github.com/Nolus-Protocol/nolus-networks](https://github.com/Nolus-Protocol/nolus-networks)
