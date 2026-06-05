# x/interchaintxs

Interchain Accounts (ICA) controller module. It lets a Nolus account (typically a
CosmWasm contract) register an interchain account on a remote chain over IBC and
submit transactions to be executed there.

Originally a Neutron module, internalized into nolus-core to allow direct
modification (e.g. unordered ICA channels); see
[ADR 20241025](../../doc/adr/20241025-decouple-and-minting-formula.md).

## Messages

| Message                        | Effect                                                                            |
|--------------------------------|-----------------------------------------------------------------------------------|
| `MsgRegisterInterchainAccount` | Open an ICA channel for `(owner, interchain_account_id)`; charges `register_fee`. |
| `MsgSubmitTx`                  | Send `msgs` to the registered ICA over `connection_id` with a `timeout`.          |
| `MsgUpdateParams`              | Governance: update `Params`.                                                      |

## Params

- `msg_submit_tx_max_messages` — upper bound on the number of messages a single
  `MsgSubmitTx` may carry.

## Wiring

The keeper implements the IBC controller callbacks (`ibc_module.go`) and forwards
packet lifecycle events to the registering contract. ICA fees are escrowed through
[`x/feerefunder`](../feerefunder/README.md); IBC acknowledgement/timeout errors are
recorded by [`x/contractmanager`](../contractmanager/README.md).
