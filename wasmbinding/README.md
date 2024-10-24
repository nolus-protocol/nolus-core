# CosmWasm support

This package allows for custom queries and custom messages sends from contract.


### What is supported 

- Queries:
  - InterchainAccountAddress - Get the interchain account address by owner_id and connection_id
- Messages:
  - RegisterInterchainAccount - register an interchain account
  - SubmitTx - submit a transaction for execution on a remote chain


## Command line interface (CLI)

- Commands

```sh
  nolusd tx wasm -h
```

- Query

```sh
  nolusd query wasm -h
```

## Tests
