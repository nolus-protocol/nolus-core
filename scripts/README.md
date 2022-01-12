# Scripts

## Prerequisites

The binary `cosmzoned` must be on the system path to allow scripts to run it.

## penultimate-genesis.sh

Script that generates a non-final genesis, namely without any gen_txs. For more details check the Cosmos hub [GENESIS CEREMONY](https://github.com/cosmos/mainnet/blob/master/GENESIS-CEREMONY.md) process.

Sample usage:
```shell
  penultimate-genesis.sh --output "proto-genesis.json"
```

## init-test-network.sh

Initialize the directory structure for 1 or more validators. Internally it invokes calls to `penultimate-genesis.sh` to create a proto genesis. It then initializes validator nodes via `init-validator-node.sh` by also passing them the previously generated proto-genesis and finally, it creates a final genesis by combining the generated gentx transactions via `collect-validator-gentxs.sh`.

Sample usage which generates 2 validator nodes:
```shell
init-validator-node.sh -v 2 --output validator_setup
```

## init-validator-node.sh

Generate full node file structure by receiving a penultimate genesis and validator specific configurations. The script calls `cosmzoned gentx` thus it will have a gentx file in the `<node_dir>/config/gentx` directory.

Sample usage:
```shell
init-validator-node.sh -g "penultimate-genesis.json" -d "node1" --moniker "validator-1" --mnemonic "<24 words mnemonic>"
```

## edit-configuration.sh

Edits full node configuration files such as `app.toml` and `config.toml`.

Sample usage:
```shell
edit-configuration.sh --home ./validator_setup/node1 --enable-api true
```

## create-vesting-account.sh

Support library for vesting account creation (internally used in the `penultimate-genesis.sh`).

Sample usage:
```shell
  source create-vesting-account.sh
  row="{\"address\": \"$addr\", \"amount\": \"$amnt\", \"vesting\": { \"type\": \"periodic\", \"start-time\": \"$start_date\", \"end-time\": \"$end_date\", \"amount\": \"$amnt\", \"periods\": 4, \"length\": 14400}}"
  add_vesting_account "$row" "./validator_setup/node1"

```

## collect-validator-gentxs.sh

Used to collect a directory of messages gentx into a single genesis file.

Sample usage:

```shell
collect-validator-gentxs.sh --collector "node1" --gentxs "gentxs"
```