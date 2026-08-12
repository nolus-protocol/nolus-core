package v084

import (
	"context"
	"fmt"

	"github.com/Nolus-Protocol/nolus-core/app/keepers"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
)

// CreateUpgradeHandler for v0.8.4 carries no state migration: the release is a
// security patch to the ibc-go Solana light client (relayer-allowlist gating of
// MsgSubmitMisbehaviour, the SolanaHeader client-type guard, and header slot and
// timestamp bounds), all of which live in the binary and take effect the moment
// the upgraded node runs. The handler exists so the software-upgrade plan named
// "v0.8.4" resolves and the chain resumes on the patched binary.
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
