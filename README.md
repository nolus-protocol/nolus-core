# cosmzone
**cosmzone** is a blockchain built using Cosmos SDK and Tendermint and created with [Starport](https://github.com/tendermint/starport).

## Prerequisites

Install [golang](https://golang.org/) and [jq](https://stedolan.github.io/jq/).

## Get started

```
./init.sh
```

The init script will compile and start the network

### Configure

Your blockchain in development can be configured within the `init.sh` script. Check out the `update_genesis` method and its usages

## Integration testing

Tests can be run by installing dependencies via: `cd tests/integration && yarn install` and then running `./run-integration.sh`