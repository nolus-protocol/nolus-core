# cosmzone
**cosmzone** is a blockchain built using Cosmos SDK and Tendermint and created with [Starport](https://github.com/tendermint/starport).

## Prerequisites

Install [golang](https://golang.org/), [tomlq](https://tomlq.readthedocs.io/en/latest/installation.html) and [jq](https://stedolan.github.io/jq/).

## Get started

```
make install
scripts/init-dev-network.sh -v 1 --output validators
cosmzoned start --home "./validators/node1"
```

The `make install` command will compile and locally install cosmzoned on your machine. `init-dev-network.sh` generates a node setup (run `init-dev-network.sh --help` for more configuration options) and `cosmzoned start` starts the network. For more details check the [scripts README](./scripts/README.md)

### Configure

Your blockchain in development can be configured within the `init.sh` script. Check out the `update_genesis` method and its usages

## Integration testing

Tests can be run by installing dependencies via: `cd tests/integration && yarn install` and then running `./run-integration.sh`