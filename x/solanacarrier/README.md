# x/solanacarrier

Parameters for sign mode 193, `SIGN_MODE_SOLANA_TX_CARRIER`: a Solana wallet signs a canonical Solana legacy message that travels in the Cosmos transaction's non-critical extension options (`/nolus.solanacarrier.v1.SolanaCarrier`). The sign-mode handler lives in `solanacarrier/`; this module stores the parameters it verifies a carrier against and lets governance change them.

## Params

| Field | Meaning |
|---|---|
| `fee_payer` | 32-byte account key every carrier places at account index 0. It must be off the ed25519 curve, so no one can sign for it and a carrier can never become a broadcastable Solana transaction. |
| `allowed_programs` | Programs a carrier may invoke besides its one SPL Memo instruction, each with `program_id` (32 bytes) and `max_instructions` (1 to 64). |

The SPL Memo program is not part of the list and cannot be added to it: every carrier holds exactly one Memo instruction whose data is the transaction's amino JSON, byte for byte. An empty `allowed_programs` admits that Memo instruction alone.

`Params.Validate` refuses a `program_id` that is not 32 bytes, the same program listed twice, the Memo program, `max_instructions` of 0 or above 64, and more than 16 entries. Genesis, `MsgUpdateParams` and the v2 migration all go through it.

Defaults (fresh chains, the v0.8.6 handler, and the v1 to v2 migration in v0.8.7):

| Program | base58 | base64 (as in params JSON) | max |
|---|---|---|---|
| ComputeBudget | `ComputeBudget111111111111111111111111111111` | `AwZGb+UhFzL/7K26csOb57yM5bvF9xJrLEObOkAAAAA=` | 4 |
| Lighthouse | `L2TExMFKdjpN9kozasaurPirfHy9P8sbXoAN1qA3S95` | `BN+teWL/sd2SXQqftebQDOYZW6i7OpH9B++YYMXpe7g=` | 4 |

## Verification

For each carrier signature the handler reads the params once, then checks in order: the fee payer at index 0, the signer among the required signers, every non-Memo instruction against `allowed_programs` (an unlisted program fails with `instruction <i> invokes a program outside the allowlist`), each listed program's instruction count against its `max_instructions` (checked in list order; the error names the program in base58 and the cap), and exactly one Memo bound to the amino JSON.

## Changing the allowlist

Submit a `MsgUpdateParams` gov proposal. `x/solanacarrier/testdata/update-params-proposal.json` is a ready template carrying today's defaults; `TestUpdateParamsProposalExampleMatchesTheDefaults` regenerates it from the Go constants (`go test ./x/solanacarrier/ -run ProposalExample -update`). Set `deposit`, `title` and `summary` before submitting.

- **The proposal replaces the whole params.** Supply the fee payer and every program that should stay allowed. Omitting `allowed_programs` empties the list, leaving carriers with the Memo instruction alone.
- **Concurrent proposals: the last one executed wins.** Each carries a full list, so a later proposal undoes an earlier one it does not repeat.
- **Program ids are base64 in JSON.** Solana tools print base58. Convert with:

  ```bash
  python3 -c 'import base64,sys;A="123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz";n=0
  for c in sys.argv[1]: n=n*58+A.index(c)
  print(base64.b64encode(n.to_bytes(32,"big")).decode())' L2TExMFKdjpN9kozasaurPirfHy9P8sbXoAN1qA3S95
  ```

- **A removal does not evict the mempool.** CheckTx re-checks after a block skip signature verification, so a carrier accepted into the mempool before a program was removed stays there and fails when it is delivered.
- **Gas.** The params read is metered: each entry adds 38 bytes to the stored params, about 114 gas per carrier signature (about 230 at the two defaults, about 1.8k at 16 entries).

## Store and upgrades

Params live under key `0x01`. Consensus version 2 added `allowed_programs`; the v0.8.7 upgrade migrates version 1 params by keeping the stored fee payer and writing the list frozen in `migrations/v2` (equal to today's defaults, pinned by a test). Stored params missing at migration time fail the upgrade rather than being replaced by defaults.

Exporting state from before the v0.8.7 upgrade with a v0.8.7 binary yields params with an empty `allowed_programs`, since version 1 params decode without the field; a chain started from such an export admits the Memo instruction alone until governance sets the list.
