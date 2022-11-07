# nolus
**nolus** is a blockchain built using Cosmos SDK and Tendermint.
## Prerequisites

Install [golang](https://golang.org/), [tomlq](https://tomlq.readthedocs.io/en/latest/installation.html) and [jq](https://stedolan.github.io/jq/).

## Get started

Make sure nolus-money-market repo is checked out as a sibling to this repo.

```
make install
./scripts/init-local-network.sh
nolusd start --home "networks/nolus/local-validator-1"
```

The `make install` command will compile and locally install nolusd on your machine. `init-local-network.sh` generates a node setup (run `init-local-network.sh --help` for more configuration options) and `nolusd start` starts the network. For more details check the [scripts README](./scripts/README.md)

## Build statically linked binary

By default, `make build` generates a dynamically linked binary. In case someone would like to reproduce the way the binary is built in the pipeline then the command to achieve it locally is:

```shell
docker run --rm -it -v "$(pwd)":/code public.ecr.aws/nolus/builder:<replace_with_the latest_tag> make build -C /code
```

## Upgrade wasmvm
- Update the Go modules
- Update the wasmvm version in the builder Dockerfile at build/builder_spec
- Increment the NOLUS_BUILDER_TAG in the Gitlab pipeline definition at .gitlab-ci.yml
- (optional step if the branch is not Gitlab protected) In order to let the pipeline build and publish the new Nolus builder image, the build should be done on a protected branch. By default only main is protected. If the upgrade is done in another one then turn it protected until a successfull build finishes.


## Build image locally and run a full node with docker 

In order to run, you should put the artifact binary in the nolus-core directory and rename it to `nolus.tar.gz`. 
The artifact binary is in the git pipeline under the `build-binary` command.

#Testnet rila.

```
docker build -f build/node_spec.Dockerfile -t rila-testnet-image .
docker run -d -it \
  --name testnet-rila \
  -v nolusDataVol:/.nolus/data \
  rila-testnet-image
```