package v085

import (
	"context"
	"fmt"

	"github.com/Nolus-Protocol/nolus-core/app/keepers"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
)

// CreateUpgradeHandler for v0.8.5 carries no state migration: the release adds
// the SOLANA_OFFCHAIN and SOLANA_TX_CARRIER custom sign modes, which live in the
// binary and are wired at app construction. The handler exists so the
// software-upgrade plan named "v0.8.5" resolves and so its recorded done-height
// is the deterministic activation signal for those sign modes (the ante gate in
// SolanaSignModeGateDecorator rejects them until this upgrade has been applied,
// which keeps an early-swapped binary in consensus with un-upgraded nodes).
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	_ *keepers.AppKeepers,
	_ codec.Codec,
) upgradetypes.UpgradeHandler {
	return func(c context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx := sdk.UnwrapSDKContext(c)

		ctx.Logger().Info("Starting module migrations...")
		vm, err := mm.RunMigrations(ctx, configurator, vm) //nolint:contextcheck
		if err != nil {
			return vm, err
		}

		ctx.Logger().Info(fmt.Sprintf("Migration {%s} applied", UpgradeName))
		return vm, nil
	}
}
