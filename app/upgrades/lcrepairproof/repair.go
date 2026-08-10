package lcrepairproof

import (
	"context"
	"fmt"

	"github.com/Nolus-Protocol/nolus-core/app/keepers"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	clienttypes "github.com/cosmos/ibc-go/v10/modules/core/02-client/types"
	host "github.com/cosmos/ibc-go/v10/modules/core/24-host"
	"github.com/cosmos/ibc-go/v10/modules/core/exported"
	ibctm "github.com/cosmos/ibc-go/v10/modules/light-clients/07-tendermint"
)

// CreateRepairUpgradeHandler lowers the target client's frontier back onto the highest slot
// that still holds a consensus state below it, and deletes everything above.
//
// A single oversized SolanaHeader moves LatestHeight to a slot no later header can exceed
// and writes a consensus state there, so the client keeps reporting Active and every IBC
// exit closes: MsgRecoverClient refuses a non-Frozen, non-Expired subject, and
// MsgUpgradeClient needs a height above one that is already unreachable.
func CreateRepairUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	k *keepers.AppKeepers,
	cdc codec.Codec,
) upgradetypes.UpgradeHandler {
	return func(c context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx := sdk.UnwrapSDKContext(c)

		clientID, cs, err := loadTargetClient(ctx, k, RepairUpgradeName)
		if err != nil {
			return vm, err
		}

		bricked := cs.LatestHeight
		clientStore := k.IBCKeeper.ClientKeeper.ClientStore(ctx, clientID)

		var stored []exported.Height
		ibctm.IterateConsensusStateAscending(clientStore, func(height exported.Height) bool {
			stored = append(stored, height)
			return false
		})

		var repaired exported.Height
		var poisoned []exported.Height
		for _, height := range stored {
			if height.LT(bricked) {
				repaired = height
				continue
			}
			poisoned = append(poisoned, height)
		}

		if repaired == nil {
			return vm, fmt.Errorf(
				"%s: client %s has no consensus state below its frontier %s to fall back to",
				RepairUpgradeName, clientID, bricked,
			)
		}

		// The frontier must land on a height that still has a consensus state, or status()
		// reports Expired through the "no consensus state at latest height" branch, which sits
		// above the early return that keeps Solana clients Active.
		if _, ok := ibctm.GetConsensusState(clientStore, cdc, repaired); !ok {
			return vm, fmt.Errorf(
				"%s: client %s has an iteration entry at %s with no consensus state behind it",
				RepairUpgradeName, clientID, repaired,
			)
		}

		ctx.Logger().Info(fmt.Sprintf(
			"%s: client %s LatestHeight %s -> %s, dropping %d consensus state(s) at or above the bricked frontier",
			RepairUpgradeName, clientID, bricked, repaired, len(poisoned),
		))

		cs.LatestHeight = clienttypes.NewHeight(repaired.GetRevisionNumber(), repaired.GetRevisionHeight())
		k.IBCKeeper.ClientKeeper.SetClientState(ctx, clientID, cs)

		// Each consensus state owns three metadata keys besides itself. Dropping the state and
		// leaving the iteration key behind would leave the index pointing at nothing, which
		// breaks pruning and the neighbour lookups the update path does on every header.
		for _, height := range poisoned {
			clientStore.Delete(host.ConsensusStateKey(height))
			clientStore.Delete(ibctm.ProcessedTimeKey(height))
			clientStore.Delete(ibctm.ProcessedHeightKey(height))
			clientStore.Delete(ibctm.IterationKey(height))
		}

		vm, err = mm.RunMigrations(ctx, configurator, vm) //nolint:contextcheck
		if err != nil {
			return vm, err
		}

		ctx.Logger().Info(fmt.Sprintf("Migration {%s} applied", RepairUpgradeName))
		return vm, nil
	}
}
