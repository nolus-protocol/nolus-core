# x/tax

Nolus-custom fee module. It does two things at the ante-handler level:

1. **Custom fee checker** (`CustomTxFeeChecker`) — lets transactions pay fees in
   approved foreign denoms, not only the base denom (`unls`). Accepted denoms and
   their minimum prices are governance parameters. See
   [ADR 20250115](../../doc/adr/20250115-accept-fees-in-foreign-denoms.md).
2. **Tax deduction** (`DeductTaxDecorator`) — takes `fee_rate` percent of the
   collected transaction fee and routes it: base-denom fees go to the
   `treasury_address`, foreign-denom fees go to the matching DEX `profit_address`.
   The remainder stays as the validator fee.

## Params

| Field              | Type             | Meaning                                                        |
|--------------------|------------------|----------------------------------------------------------------|
| `fee_rate`         | `int32`          | Percent of the tx fee taken as tax (`0` disables deduction).   |
| `base_denom`       | `string`         | Native fee denom (`unls`).                                     |
| `treasury_address` | `string`         | Receives base-denom tax.                                       |
| `dex_fee_params`   | `[]DexFeeParams` | Per-DEX profit address + accepted foreign denoms / min prices. |

## Messages

| Message           | Authority   | Effect           |
|-------------------|-------------|------------------|
| `MsgUpdateParams` | governance  | Update `Params`. |

## Wiring

Both decorators are installed in the ante chain (`app/ante.go`);
`TaxKeeper.CustomTxFeeChecker` is passed to the SDK `DeductFeeDecorator` in
`app/app.go`. `typesv2` holds the current params type; `migrations/` upgrades the
v1beta1 params to v2.
