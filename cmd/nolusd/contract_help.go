package main

import (
	"github.com/spf13/cobra"
)

// contractHelpCmd prints a flat reference of every Nolus smart-contract
// operation (query + execute) with JSON templates and example invocations.
//
// Placeholders like <admin>, <leaser>, <lease>, <lpp>, <oracle>, <treasury>
// are intentional — addresses differ per network/protocol/user and must be
// resolved at runtime via the discovery queries shown at the top of the
// output. Shapes are kept in sync with @nolus/nolusjs src/contracts/messages.
func contractHelpCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "contract-help",
		Short: "Reference for Nolus smart-contract queries and executes",
		Long:  contractHelpText,
		Run: func(cmd *cobra.Command, _ []string) {
			cmd.Print(contractHelpText)
		},
	}
}

const contractHelpText = `Nolus smart-contract operations reference.

All operations target CosmWasm contracts via the standard wasm subcommands:
  Query:    nolusd q wasm contract-state smart <contract> '<json>' --output json
  Execute:  nolusd tx wasm execute <contract> '<json>' \
              --from <key> --gas auto --gas-adjustment 1.3 \
              --fees <fee><denom> --chain-id <chain-id> [--amount <coin>]

Placeholders below (<admin>, <leaser>, <lease>, <lpp>, <oracle>, <treasury>,
<ticker>, <bech32>, <name>, <micro-units>) MUST be resolved at runtime.
Markers: [Q] read-only query   [X] execute (state-mutating tx)

================================================================================
ADDRESS DISCOVERY  (start here)
================================================================================

The Admin contract is the entry point. From it you can resolve every other
contract address. Known admin addresses by chain-id:

  pirin-1     (mainnet)  nolus1gurgpv8savnfw66lckwzn4zk7fp394lpe667dhu7aw48u40lj6jsqxf8nd
  rila-3      (testnet)  nolus17p9rzwnnfxcjp32un9ug7yhhzgtkhvl9jfksztgw5uh69wac2pgsmc5xhq
  vitosha-8   (testnet)  nolus150ggq0zu22wf7jwehces77j9m7zxarnkd0v0j60cwancqh5cke8sx7x6p7

Substitute the one matching your --chain-id for every <admin> placeholder
below. For any other chain, obtain the admin address out-of-band.

[Q] List all registered protocol identifiers (e.g. OSMOSIS-OSMOSIS-USDC_NOBLE).
  nolusd q wasm contract-state smart <admin> '{"protocols":{}}' --output json

[Q] Get contract addresses (leaser, lpp, oracle, profit, ...) for one protocol.
  nolusd q wasm contract-state smart <admin> '{"protocol":"<name>"}' --output json

[Q] Get platform-wide contract addresses (timealarms, treasury).
  nolusd q wasm contract-state smart <admin> '{"platform":{}}' --output json

[Q] List the lease (margin-position) addresses owned by a wallet.
  nolusd q wasm contract-state smart <leaser> '{"leases":{"owner":"<bech32>"}}' --output json

================================================================================
ADMIN
================================================================================
Governs storage and migration of all smart contracts on the chain.

[Q] protocols  —  registered protocol identifiers
  nolusd q wasm contract-state smart <admin> '{"protocols":{}}'

[Q] protocol  —  per-protocol contract addresses; <name> from {"protocols":{}}
  nolusd q wasm contract-state smart <admin> '{"protocol":"<name>"}'

[Q] platform  —  platform-wide contracts (timealarms, treasury)
  nolusd q wasm contract-state smart <admin> '{"platform":{}}'

================================================================================
LEASER     (one per protocol)
================================================================================
Customer-facing registry of Lease (margin-position) contracts.

[Q] config  —  liability thresholds, margin rate, min amounts, IBC channels
  nolusd q wasm contract-state smart <leaser> '{"config":{}}'

[Q] leases  —  active lease addresses owned by <bech32>
  nolusd q wasm contract-state smart <leaser> '{"leases":{"owner":"<bech32>"}}'

[Q] quote  —  estimate borrow amount, total position, interest rate.
  Amounts in micro-units. max_ltd is in permilles (1500 = 150% = 2.5x); omit
  for the contract default.
  nolusd q wasm contract-state smart <leaser> \
    '{"quote":{"lease_asset":"<ticker>","downpayment":{"ticker":"<ticker>","amount":"<micro-units>"}}}'
  nolusd q wasm contract-state smart <leaser> \
    '{"quote":{"lease_asset":"<ticker>","downpayment":{"ticker":"<ticker>","amount":"<micro-units>"},"max_ltd":<permilles>}}'

[X] open_lease  —  open a leveraged position. Collateral via --amount as the
  IBC-denom form of <ticker>. max_ltd is optional.
  nolusd tx wasm execute <leaser> \
    '{"open_lease":{"currency":"<ticker>"}}' \
    --amount <micro-units><ibc-denom> --from <key> ...
  nolusd tx wasm execute <leaser> \
    '{"open_lease":{"currency":"<ticker>","max_ltd":<permilles>}}' \
    --amount <micro-units><ibc-denom> --from <key> ...

================================================================================
LEASE     (one per active margin position; resolve via leaser '{"leases":...}')
================================================================================

[Q] state  —  current position: principal, interest, due, LTV, SL/TP, status.
  due_projection_secs (optional) projects accrued interest N seconds ahead.
  nolusd q wasm contract-state smart <lease> '{"state":{}}'
  nolusd q wasm contract-state smart <lease> '{"state":{"due_projection_secs":<seconds>}}'

[X] repay  —  repay outstanding debt (anyone may repay). Funds via --amount.
  nolusd tx wasm execute <lease> '{"repay":[]}' \
    --amount <micro-units><ibc-denom> --from <key> ...

[X] close_position (full)  —  market-close the entire position to LPN.
  nolusd tx wasm execute <lease> '{"close_position":{"full_close":{}}}' --from <key> ...

[X] close_position (partial)  —  market-close part of the position.
  amount.ticker is the lease asset.
  nolusd tx wasm execute <lease> \
    '{"close_position":{"partial_close":{"amount":{"ticker":"<ticker>","amount":"<micro-units>"}}}}' \
    --from <key> ...

[X] change_close_policy  —  set/clear stop-loss and/or take-profit (LTV
  thresholds, permilles). Use {"set":<n>} to set, "reset" to clear; omit a
  field to leave it unchanged.
  nolusd tx wasm execute <lease> \
    '{"change_close_policy":{"stop_loss":{"set":<permilles>},"take_profit":{"set":<permilles>}}}' \
    --from <key> ...
  nolusd tx wasm execute <lease> \
    '{"change_close_policy":{"stop_loss":"reset","take_profit":"reset"}}' \
    --from <key> ...

================================================================================
LPP — Liquidity Provider Pool     (one per LPN currency per protocol)
================================================================================
Single-sided lender pool; issues loans to Lease instances; mints nLPN receipts.

[Q] config  —  base rate, optimal/min utilization, slope (all permilles)
  nolusd q wasm contract-state smart <lpp> '{"config":[]}'

[Q] lpp_balance  —  available, total principal due, total interest due, nLPN
  nolusd q wasm contract-state smart <lpp> '{"lpp_balance":[]}'

[Q] stable_balance  —  pool value priced in the oracle's stable currency
  nolusd q wasm contract-state smart <lpp> '{"stable_balance":{"oracle_addr":"<oracle>"}}'

[Q] deposit_capacity  —  remaining capacity given the min-utilization rule
  nolusd q wasm contract-state smart <lpp> '{"deposit_capacity":[]}'

[Q] lpn  —  this pool's native asset ticker (e.g. USDC_NOBLE)
  nolusd q wasm contract-state smart <lpp> '{"lpn":[]}'

[Q] price  —  nLPN ↔ LPN receipt price
  nolusd q wasm contract-state smart <lpp> '{"price":[]}'

[Q] rewards  —  claimable NLS incentives for a lender wallet
  nolusd q wasm contract-state smart <lpp> '{"rewards":{"address":"<bech32>"}}'

[Q] balance  —  nLPN deposit balance held by a lender wallet
  nolusd q wasm contract-state smart <lpp> '{"balance":{"address":"<bech32>"}}'

[X] deposit  —  deposit LPN into the pool, receive nLPN. Funds via --amount.
  nolusd tx wasm execute <lpp> '{"deposit":[]}' \
    --amount <micro-units><lpn-ibc-denom> --from <key> ...

[X] burn  —  burn nLPN to withdraw LPN.
  nolusd tx wasm execute <lpp> '{"burn":{"amount":{"amount":"<micro-units>"}}}' \
    --from <key> ...

[X] claim_rewards  —  claim accumulated NLS incentives. other_recipient
  optional; omit to claim to --from.
  nolusd tx wasm execute <lpp> '{"claim_rewards":{}}' --from <key> ...
  nolusd tx wasm execute <lpp> \
    '{"claim_rewards":{"other_recipient":"<bech32>"}}' --from <key> ...

================================================================================
ORACLE     (one per protocol)
================================================================================
Provides market prices and swap topology to the protocol.

[Q] base_currency  —  pricing base ticker
  nolusd q wasm contract-state smart <oracle> '{"base_currency":{}}'

[Q] stable_currency  —  ticker used for stable-denominated prices
  nolusd q wasm contract-state smart <oracle> '{"stable_currency":{}}'

[Q] prices  —  all known asset prices (mapping)
  nolusd q wasm contract-state smart <oracle> '{"prices":{}}'

[Q] base_price  —  price of <ticker> in base currency
  nolusd q wasm contract-state smart <oracle> '{"base_price":{"currency":"<ticker>"}}'

[Q] stable_price  —  price of <ticker> in stable currency
  nolusd q wasm contract-state smart <oracle> '{"stable_price":{"currency":"<ticker>"}}'

[Q] currencies  —  all supported currencies (ticker, bank/dex symbols, decimals, group)
  nolusd q wasm contract-state smart <oracle> '{"currencies":{}}'

================================================================================
NOTES
================================================================================
- Micro-units: amounts are in the smallest indivisible unit (e.g. 6 decimals
  for USDC/OSMO/NLS, 8 for ALL_BTC, 9 for ALL_SOL). 15 USDC = "15000000".
- Permilles: rates and ratios are in thousandths. 1500 = 150%, 80 = 8%.
- IBC denoms: collateral and repayments use the IBC-wrapped denom of the
  ticker on Nolus, not the raw ticker. Resolve via Oracle '{"currencies":{}}'.
`
