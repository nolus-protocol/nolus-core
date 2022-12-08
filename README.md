# NOLUS

**nolus** is a blockchain built using Cosmos SDK and Tendermint.

## Prerequisites

Install [golang](https://golang.org/), [tomlq](https://tomlq.readthedocs.io/en/latest/installation.html) and [jq](https://stedolan.github.io/jq/).

## Get started

### Build, configure and run a single-node locally deployed Nolus chain

#### Build

  ```sh
  make install
  ```

#### Init && setup the DEX parameters && run

First generate the mnemonic you will use for Hermes:

```sh
nolusd keys mnemonic
```

Then recover osmosis key and use the [UI](https://faucet.osmosis.zone/#/) to get some uosmo:

```sh
osmosisd keys add hermes_key --recover
```

Init and start:

```sh
./scripts/init-local-network.sh --reserve-tokens <reserve_account_init_tokens> --hermes-mnemonic <the_mnemonic_generated_by_the_previous_steps>
```

The `make install` command will compile and locally install nolusd on your machine. `init-local-network.sh` generates a node setup, including setting the dex parameter (run `init-local-network.sh --help` for more configuration options). For more details check the [scripts README](./scripts/README.md)

*Notes:

* Make sure "nolus-money-market" repo is checked out as a sibling to this repo.

* The Osmosis binary is required: [Follow the steps](https://gitlab-nomo.credissimo.net/nomo/wiki/-/blob/main/hermes.md#download-the-latest-osmosis-binary).

* Before running the "./scripts/init-local-network.sh" again, make sure the nolusd and hermes processes are killed.

* The "hermes" and "nolusd" logs are stored in ~/hermes and ~/.nolus respectively.

### Run already configured single-node

```sh
nolusd start --home "networks/nolus/local-validator-1"
```

## Install, configure and run a local Hermes relayer manually

Follow the steps [here](https://gitlab-nomo.credissimo.net/nomo/wiki/-/blob/main/hermes.md#install-and-configure-hermes). Write down the connection and channel identifiers at Nolus and Osmosis for further usage.

## Setup the DEX parameters manually

The goal is to let smart contracts know the details of the connectivity to Osmosis. Herebelow is a sample request.

```sh
nolusd tx wasm execute nolus1wn625s4jcmvk0szpl85rj5azkfc6suyvf75q6vrddscjdphtve8s5gg42f '{"setup_dex": {"connection_id": "connection-0", "transfer_channel": {"local_endpoint": "channel-0", "remote_endpoint": "channel-1499"}}}' --fees 387unls --gas auto --gas-adjustment 1.1
```

*Notes:
"channel-1499" should be replaced, so you can get the actual channel ID of the remote endpoint with:

```sh
nolusd q ibc channel connections connection-0 --output json | jq '.channels[0].counterparty.channel_id' | tr -d '"'
```

Check the transaction has passed:

```sh
nolusd q wasm contract-state smart nolus1wn625s4jcmvk0szpl85rj5azkfc6suyvf75q6vrddscjdphtve8s5gg42f '{"config":{}}'
```

## Build statically linked binary

By default, `make build` generates a dynamically linked binary. In case someone would like to reproduce the way the binary is built in the pipeline then the command to achieve it locally is:

```sh
docker run --rm -it -v "$(pwd)":/code public.ecr.aws/nolus/builder:<replace_with_the latest_tag> make build -C /code
```

## Upgrade wasmvm

* Update the Go modules
* Update the wasmvm version in the builder Dockerfile at build/builder_spec
* Increment the NOLUS_BUILDER_TAG in the Gitlab pipeline definition at .gitlab-ci.yml
* (optional step if the branch is not Gitlab protected) In order to let the pipeline build and publish the new Nolus builder image, the build should be done on a protected branch. By default only main is protected. If the upgrade is done in another one then turn it protected until a successfull build finishes.

## Build image locally and run a full node with docker

In order to build the image, you should put the artifact binary in the nolus-core directory and rename it to `nolus.tar.gz`.
You should use the nolus version used in the testnet-rila, current is v0.1.37(as of November 8th 2022).
The artifact binary is in the git pipeline under the `build-binary` command.

### Testnet rila

```sh
ACCESS_TOKEN=<get token for gitlab wiki access>
docker build \
  --build-arg ACCESS_TOKEN=$ACCESS_TOKEN \
  -f build/node_spec.Dockerfile \
  -t rila-testnet-image .

docker run -d -it \
  --name testnet-rila \
  -v nolusDataVol:/.nolus/data \
  rila-testnet-image
```

*Notes:

* Make sure the genesis.json, presistent_peers.txt files in the wiki repo are for the same version of the artifact binary.
