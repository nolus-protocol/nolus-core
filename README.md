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

#### Initialize, set up and run

`init-local-network.sh` generates a network setup, including the deployment of platform contracts (only) and initial Hermes setup (Nolus chain configuration).

First, generate the mnemonic you will use for Hermes:

```sh
nolusd keys mnemonic
```

Initialize and start (run `./scripts/init-local-network.sh --help` for additional configuration options):

```sh
./scripts/init-local-network.sh --reserve-tokens <reserve_account_init_tokens> --hermes-mnemonic <the_mnemonic_generated_by_the_previous_steps> --dex-admin-mnemonic <mnemonic_phrase> --store-code-privileged-account-mnemonic <mnemonic_phrase>
```

*Notes:

* Make sure the `nolus-money-market` repository is checked out as a sibling to this one.

* !!! Before running the `./scripts/init-local-network.sh` again, make sure the `nolusd` and `hermes` processes are killed.

* The nolusd logs are stored in `~/.nolus`.

### Run an already configured single-node

```sh
nolusd start --home "networks/nolus/local-validator-1"
```

## Set up a new DEX

On a live network, a new DEX can be deployed using the following steps.

### Manual step - Prerequisites

* provide a certain amount for the Hermes account (DEX side)

    Recover your Hermes wallet on the DEX network and use a faucet to obtain some amount.

    Example for the Osmosis DEX ([Osmo-test-5 faucet](https://faucet.osmotest5.osmosis.zone/)):

    ```sh
    osmosisd keys add hermes_key --recover
    ```

* start hermes

### Аutomated step

```sh
./scripts/add-new-dex.sh --dex-admin-key <dex_admin_key> --store-code-privileged-user-key <store_code_privileged_user_key> --wasm-artifacts-path <wasm_artifacts_dir_path> --dex-name <dex_name> --dex-chain-id <new_dex_chain_id> --dex-ip-addr-rpc-host <new_dex_ip_addr_rpc_host_part> --dex-ip-addr-grpc-host <new_dex_ip_addr_grpc_host_part> --dex-account-prefix <new_dex_account_prefix> --dex-price-denom <new_dex_price_denom> --dex-trusting-period-secs <new_dex_trusting_period_in_seconds>  --dex-if-interchain-security <if_interchain_security_true/false> --protocol-currency <new_protocol_currency> --stable-currency <new_protocol_stable_currency> --protocol-swap-tree <new_protocol_swap_tree>
```

The script takes care of setting up Hermes to work with the new DEX and, for now, deploying DEX-specific contracts (More about deploying contracts on a live network can be found [here](https://github.com/nolus-protocol/nolus-money-market)).

*Notes:

* Execute `./scripts/add-new-dex.sh --help` for additional configuration options

* The `protocol-swap-tree` must be passed in single quotes (for example: **--protocol-swap-tree '{"value":[0,"USDC"],"children":[{"value":[5,"OSMO"],"children":[{"value":[12,"ATOM"]}]}]}'**)

* The script will locate the Hermes account from the Hermes configuration directory and link it to the new DEX

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
