# x/feerefunder

Escrows the IBC relayer fees that back asynchronous interchain operations
(ICA / IBC transfers). When a contract sends an interchain packet it locks a
`Fee` up front; once the packet resolves, the matching portion is paid to the
relayer and any unused portion is refunded.

Originally a Neutron module, internalized into nolus-core; see
[ADR 20241025](../../doc/adr/20241025-decouple-and-minting-formula.md).

## Concepts

- **`Fee`** — three coin buckets per packet: `recv_fee`, `ack_fee`, `timeout_fee`.
- **`PacketID`** — `(port_id, channel_id, sequence)` key the locked fee is stored under.

## Keeper API (used by other modules)

| Method                        | Effect                                                       |
|-------------------------------|--------------------------------------------------------------|
| `LockFees`                    | Escrow a payer's `Fee` for a packet before it is sent.       |
| `DistributeAcknowledgementFee`| On ack: pay the relayer, refund the timeout bucket.          |
| `DistributeTimeoutFee`        | On timeout: pay the relayer, refund the ack bucket.          |
| `CheckFees`                   | Validate a `Fee` against module params before locking.       |

## Wiring

Called by [`x/interchaintxs`](../interchaintxs/README.md) and
[`x/transfer`](../transfer/README.md) through their expected-keeper interfaces.
`MsgUpdateParams` is governance-only.
