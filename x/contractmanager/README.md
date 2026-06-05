# x/contractmanager

Hardens the sudo callbacks that CosmWasm contracts receive for IBC events
(acknowledgements and timeouts). Two responsibilities:

1. **Gas-limited sudo** — wraps each sudo call in a configurable gas limit
   (`sudo_call_gas_limit`) so a misbehaving or expensive contract callback cannot
   exhaust block gas or halt packet processing (`ibc_middleware.go`).
2. **Failure recording** — if a sudo callback errors or runs out of gas, the
   failure is persisted as a `Failure` instead of reverting the IBC flow. The
   contract can later query and resubmit it.

Originally a Neutron module, internalized into nolus-core; see
[ADR 20241025](../../doc/adr/20241025-decouple-and-minting-formula.md).

## State

`Failure` — `{ address, id, sudo_payload, error }`, keyed per contract address.

## Queries

| Query             | Returns                                   |
|-------------------|-------------------------------------------|
| `Failures`        | All recorded failures.                    |
| `AddressFailures` | Failures for one contract address.        |
| `AddressFailure`  | A single failure by `(address, id)`.      |

## Params

- `sudo_call_gas_limit` — gas ceiling applied to each contract sudo callback.

## Wiring

Installed as IBC middleware; `MsgUpdateParams` is governance-only. The error path
is invoked by the IBC stack used by
[`x/interchaintxs`](../interchaintxs/README.md) and
[`x/transfer`](../transfer/README.md).
