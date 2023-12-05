package v05

import (
	"fmt"

	"github.com/Nolus-Protocol/nolus-core/app/keepers"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	icaMigrations "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/controller/migrations/v6"
	exported "github.com/cosmos/ibc-go/v7/modules/core/exported"
	ibctmmigrations "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint/migrations"
	interchaintxstypes "github.com/neutron-org/neutron/x/interchaintxs/types"
)

func setInitialMinCommissionRate(ctx sdk.Context, keepers *keepers.AppKeepers) error {
	minRate := sdk.NewDecWithPrec(5, 2)
	minMaxRate := sdk.NewDecWithPrec(1, 1)

	stakingParams := keepers.StakingKeeper.GetParams(ctx)
	stakingParams.MinCommissionRate = minRate
	if err := keepers.StakingKeeper.SetParams(ctx, stakingParams); err != nil {
		return fmt.Errorf("failed to set MinCommissionRate to 5%%: %w", err)
	}

	// Force update validator commission & max rate if it is lower than the minRate & minMaxRate respectively
	validators := keepers.StakingKeeper.GetAllValidators(ctx)
	for _, v := range validators {
		valUpdated := false
		if v.Commission.Rate.LT(minRate) {
			v.Commission.Rate = minRate
			valUpdated = true
		}
		if v.Commission.MaxRate.LT(minMaxRate) {
			v.Commission.MaxRate = minMaxRate
			valUpdated = true
		}
		if valUpdated {
			v.Commission.UpdateTime = ctx.BlockHeader().Time
			// call the before-modification hook since we're about to update the commission
			if err := keepers.StakingKeeper.Hooks().BeforeValidatorModified(ctx, v.GetOperator()); err != nil {
				return fmt.Errorf("BeforeValidatorModified failed with: %w", err)
			}
			keepers.StakingKeeper.SetValidator(ctx, v)
		}
	}

	return nil
}

func setMinInitialDepositRatio(ctx sdk.Context, keepers *keepers.AppKeepers) error {
	govParams := keepers.GovKeeper.GetParams(ctx)
	govParams.MinInitialDepositRatio = "0.25"
	return keepers.GovKeeper.SetParams(ctx, govParams)
}

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	keepers *keepers.AppKeepers,
	codec codec.Codec,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		ctx.Logger().Info("running upgrade handler")

		// ibc v4-to-v5
		// -- nothing --

		// ibc v5-to-v6
		ctx.Logger().Info("migrating ics27 channel capability")
		err := icaMigrations.MigrateICS27ChannelCapability(
			ctx,
			codec,
			keepers.GetKey(capabilitytypes.StoreKey),
			keepers.CapabilityKeeper,
			interchaintxstypes.ModuleName,
		)
		if err != nil {
			return nil, err
		}

		// ibc v6-to-v7
		// (optional) prune expired tendermint consensus states to save storage space
		ctx.Logger().Info("pruning expired tendermint consensus states for ibc clients")
		_, err = ibctmmigrations.PruneExpiredConsensusStates(ctx, codec, keepers.IBCKeeper.ClientKeeper)
		if err != nil {
			return nil, err
		}

		// ibc v7-to-v7.1
		// explicitly update the IBC 02-client params, adding the localhost client type
		ctx.Logger().Info("adding localhost client to IBC params")
		params := keepers.IBCKeeper.ClientKeeper.GetParams(ctx)
		params.AllowedClients = append(params.AllowedClients, exported.Localhost)
		keepers.IBCKeeper.ClientKeeper.SetParams(ctx, params)

		// sdk v45-to-v46
		// -- nothing --

		// sdk v46-to-v47
		// initialize param subspaces for params migration
		baseAppLegacySS := getLegacySubspaces(keepers.ParamsKeeper)
		// Migrate Tendermint consensus parameters from x/params module to a dedicated x/consensus module.
		ctx.Logger().Info("migrating tendermint x/consensus params")
		baseapp.MigrateParams(ctx, baseAppLegacySS, keepers.ConsensusParamsKeeper)

		ctx.Logger().Info("running module manager migrations")

		ctx.Logger().Info(fmt.Sprintf("[MM] pre migrate version map: %v", fromVM))
		newVersionMap, err := mm.RunMigrations(ctx, configurator, fromVM)
		if err != nil {
			return nil, err
		}
		ctx.Logger().Info(fmt.Sprintf("[MM] post migrate version map: %v", newVersionMap))

		ctx.Logger().Info("setting x/staking min commission rate to 5%")
		if err = setInitialMinCommissionRate(ctx, keepers); err != nil {
			return nil, err
		}

		ctx.Logger().Info("setting x/gov min initial deposit ratio to 25%")
		if err = setMinInitialDepositRatio(ctx, keepers); err != nil {
			return nil, err
		}

		return newVersionMap, nil
	}
}
