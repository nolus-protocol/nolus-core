// Package lcrepairproof carries two throwaway upgrades used to prove, on a local
// throwaway network, that a governance software-upgrade handler can repair an IBC light
// client that no IBC message can reach. The Solana trusted-relayer client has failure
// modes with no recovery path — a slot frontier pushed beyond what any later header can
// reach, and a MaxClockDrift fixed wrong at creation — and "a chain upgrade handler can
// rewrite the client store" is the documented fallback for both. These upgrades exercise
// that claim against a client bricked by a real relayer message.
//
// Nothing here belongs in a release. Both handlers refuse to run off LocalnetChainID.
package lcrepairproof

import (
	"time"

	store "cosmossdk.io/store/types"
	"github.com/Nolus-Protocol/nolus-core/app/upgrades"
)

const (
	// DriftUpgradeName rewrites MaxClockDrift, which is fixed at client creation and which
	// no IBC message updates.
	DriftUpgradeName = "lc-repair-proof-drift"
	// RepairUpgradeName lowers a slot frontier that a single oversized SolanaHeader pushed
	// beyond the reach of every later header (ibc-go#59), and clears the poisoned consensus
	// state that header wrote.
	RepairUpgradeName = "lc-repair-proof-repair"
)

// LocalnetChainID gates both handlers. Rewriting a light client's store behind IBC's back
// is only ever acceptable on a throwaway network, so the handlers fail the upgrade rather
// than touch a client on any other chain.
const LocalnetChainID = "nolus-local"

// TargetSolanaChainID identifies the client the handlers operate on: the one tracking the
// local Solana validator. Resolving by tracked chain rather than by client ID is what the
// harness's own suites do, because an ordinal names a different client on every stack.
const TargetSolanaChainID = "solana-localnet"

// NewMaxClockDrift is deliberately an odd value no client would be created with, so the
// post-upgrade query cannot be mistaken for the original.
const NewMaxClockDrift = 77 * time.Second

var (
	DriftUpgrade = upgrades.Upgrade{
		UpgradeName:          DriftUpgradeName,
		CreateUpgradeHandler: CreateDriftUpgradeHandler,
		StoreUpgrades:        store.StoreUpgrades{Added: []string{}, Deleted: []string{}},
	}

	RepairUpgrade = upgrades.Upgrade{
		UpgradeName:          RepairUpgradeName,
		CreateUpgradeHandler: CreateRepairUpgradeHandler,
		StoreUpgrades:        store.StoreUpgrades{Added: []string{}, Deleted: []string{}},
	}
)
