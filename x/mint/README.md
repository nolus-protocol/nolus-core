# x/mint

Nolus-custom minting module. Replaces the Cosmos SDK `x/mint` with a time-based
inflation schedule rather than a bonded-ratio target.

New `unls` is minted in `BeginBlock` proportional to the time elapsed since the
previous block, following a fixed minting formula (effective from month 17). The
module was internalized — and the formula upgraded — as part of decoupling from
external forks; see [ADR 20241025](../../doc/adr/20241025-decouple-and-minting-formula.md).

## State

- **`Minter`** — running mint state: `norm_time_passed`, `total_minted`,
  `prev_block_timestamp`, `annual_inflation`.
- **`Params`** — `mint_denom`, `max_mintable_nanoseconds` (caps the minted amount
  for an abnormally long inter-block gap, e.g. after a halt).

## Messages

| Message          | Authority   | Effect                       |
|------------------|-------------|------------------------------|
| `MsgUpdateParams` | governance | Update `Params`.             |

## Wiring

`BeginBlocker` (`abci.go`) computes and mints the per-block amount and emits the
`minted_tokens` OpenTelemetry counter. Params are governance-controlled.
