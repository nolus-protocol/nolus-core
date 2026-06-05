# x/transfer

A thin wrapper around the IBC `transfer` module that adds sudo callbacks so a
CosmWasm contract initiating an IBC transfer is notified of the packet's outcome
(acknowledgement / timeout), the same way [`x/interchaintxs`](../interchaintxs/README.md)
reports ICA results.

Originally a Neutron module, internalized into nolus-core; see
[ADR 20241025](../../doc/adr/20241025-decouple-and-minting-formula.md).

## Messages

| Message           | Effect                                                                         |
|-------------------|--------------------------------------------------------------------------------|
| `MsgTransfer`     | Standard ICS-20 transfer that, on resolution, sudo-calls the sending contract. |
| `MsgUpdateParams` | Governance: update the underlying IBC-transfer params.                         |

`MsgTransferResponse` returns the `sequence_id` and `channel`, so the caller can
correlate the later acknowledgement/timeout callback with the originating transfer.

## Wiring

`ibc_handlers.go` intercepts the IBC-transfer ack/timeout, escrows relayer fees via
[`x/feerefunder`](../feerefunder/README.md), and routes failures to
[`x/contractmanager`](../contractmanager/README.md).
