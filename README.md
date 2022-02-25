# nolus
**nolus** is a blockchain built using Cosmos SDK and Tendermint and created with [Starport](https://github.com/tendermint/starport).

## Prerequisites

Install [golang](https://golang.org/), [tomlq](https://tomlq.readthedocs.io/en/latest/installation.html) and [jq](https://stedolan.github.io/jq/).

## Get started

```
make install
./scripts/init-local-network.sh
nolusd start --home "networks/nolus/local-validator-1"
```

The `make install` command will compile and locally install nolusd on your machine. `init-local-network.sh` generates a node setup (run `init-local-network.sh --help` for more configuration options) and `nolusd start` starts the network. For more details check the [scripts README](./scripts/README.md)

## Integration testing

Tests can be run by installing dependencies via

```
cd tests/integration
yarn install
```
and run with `make test-integration` from the base project root.