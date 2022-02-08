# Scripts

## Prerequisites

The binary `nolusd` must be on the system path to allow scripts to run it.

## penultimate-genesis.sh

Script that generates a non-final genesis, namely without any gen_txs. For more details check the Cosmos hub [GENESIS CEREMONY](https://github.com/cosmos/mainnet/blob/master/GENESIS-CEREMONY.md) process.

Sample usage:
```shell
  penultimate-genesis.sh --output "proto-genesis.json"
```

## init-local-network.sh

Initialize one or more validator nodes on the local file system. First it creates accounts for the validators and generates a proto genesis. Then it lets validator nodes to create validators and stake amount. Finally, the script collects the created transactions and produces the final genesis.

The nodes are ready to be started. The nolus client is configured at the default home, "$HOME/.nolus", to point to the first validator node.

Sample usage which generates 2 validator nodes:
```shell
init-local-node.sh -v 2
```

## init-validator-node.sh

Generate full node file structure by receiving a penultimate genesis and validator specific configurations. The script calls `nolusd gentx` thus it will have a gentx file in the `<node_dir>/config/gentx` directory.

Sample usage:
```shell
init-validator-node.sh -g "penultimate-genesis.json" -d "node1" --moniker "validator-1" --mnemonic "<24 words mnemonic>"
```

## Edit Configuration

Edits full node configuration files such as `app.toml` and `config.toml`.

Sample usage:
```shell
config/edit.sh --home ./validator_setup/node1 --enable-api true
```

## create-vesting-account.sh

Support library for vesting account creation (internally used in the `penultimate-genesis.sh`).

Sample usage:
```shell
  source create-vesting-account.sh
  row="{\"address\": \"$addr\", \"amount\": \"$amnt\", \"vesting\": { \"type\": \"periodic\", \"start-time\": \"$start_date\", \"end-time\": \"$end_date\", \"amount\": \"$amnt\", \"periods\": 4, \"length\": 14400}}"
  add_vesting_account "$row" "unolus" "./validator_setup/node1"

```

## collect-validator-gentxs.sh

Used to collect a directory of messages gentx into a single genesis file.

Sample usage:

```shell
collect-validator-gentxs.sh --collector "node1" --gentxs "gentxs"
```