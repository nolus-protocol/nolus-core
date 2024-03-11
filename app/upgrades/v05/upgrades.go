package v05

import (
	"fmt"

	"github.com/Nolus-Protocol/nolus-core/app/keepers"
	"github.com/Nolus-Protocol/nolus-core/app/params"
	minttypes "github.com/Nolus-Protocol/nolus-core/x/mint/types"
	taxtypes "github.com/Nolus-Protocol/nolus-core/x/tax/types"

	storetypes "cosmossdk.io/store/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	icaMigrations "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/controller/migrations/v6"
	exported "github.com/cosmos/ibc-go/v8/modules/core/exported"
	ibctmmigrations "github.com/cosmos/ibc-go/v8/modules/light-clients/07-tendermint/migrations"

	contractmanagerkeeper "github.com/neutron-org/neutron/v2/x/contractmanager/keeper"
	contractmanagertypes "github.com/neutron-org/neutron/v2/x/contractmanager/types"
	feerefundertypes "github.com/neutron-org/neutron/v2/x/feerefunder/types"
	icqtypes "github.com/neutron-org/neutron/v2/x/interchainqueries/types"
	interchaintxstypes "github.com/neutron-org/neutron/v2/x/interchaintxs/types"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
)

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

		ctx.Logger().Info("Migrating feerefunder module parameters...")
		if err := migrateFeeRefunderParams(ctx, *keepers.ParamsKeeper, keepers.GetKey(feerefundertypes.StoreKey), codec); err != nil {
			return nil, err
		}

		ctx.Logger().Info("Migrating interchainqueries module parameters...")
		if err := migrateInterchainQueriesParams(ctx, *keepers.ParamsKeeper, keepers.GetKey(icqtypes.StoreKey), codec); err != nil {
			return nil, err
		}

		ctx.Logger().Info("Migrating interchaintxs module parameters...")
		if err := setInterchainTxsParams(ctx, *keepers.ParamsKeeper, keepers.GetKey(interchaintxstypes.StoreKey), keepers.GetKey(wasmtypes.StoreKey), codec); err != nil {
			return nil, err
		}

		ctx.Logger().Info("Migrating mint module parameters...")
		if err := migrateMintParams(ctx, *keepers.ParamsKeeper, keepers.GetKey(minttypes.StoreKey), codec); err != nil {
			return nil, err
		}

		ctx.Logger().Info("Migrating tax module parameters...")
		if err := migrateTaxParams(ctx, *keepers.ParamsKeeper, keepers.GetKey(taxtypes.StoreKey), codec); err != nil {
			return nil, err
		}

		ctx.Logger().Info("Setting sudo callback limit...")
		err = setContractManagerParams(ctx, *keepers.ContractManagerKeeper)
		if err != nil {
			return nil, err
		}

		ctx.Logger().Info(`
$$\   $$\           $$\                                       $$$$$$\     $$$$$$$\      $$$$$$\        
$$$\  $$ |          $$ |                                     $$$ __$$\    $$  ____|    $$$ __$$\       
$$$$\ $$ | $$$$$$\  $$ |$$\   $$\  $$$$$$$\       $$\    $$\ $$$$\ $$ |   $$ |         $$$$\ $$ |      
$$ $$\$$ |$$  __$$\ $$ |$$ |  $$ |$$  _____|      \$$\  $$  |$$\$$\$$ |   $$$$$$$\     $$\$$\$$ |      
$$ \$$$$ |$$ /  $$ |$$ |$$ |  $$ |\$$$$$$\         \$$\$$  / $$ \$$$$ |   \_____$$\    $$ \$$$$ |      
$$ |\$$$ |$$ |  $$ |$$ |$$ |  $$ | \____$$\         \$$$  /  $$ |\$$$ |   $$\   $$ |   $$ |\$$$ |      
$$ | \$$ |\$$$$$$  |$$ |\$$$$$$  |$$$$$$$  |         \$  /   \$$$$$$  /$$\\$$$$$$  |$$\\$$$$$$  /      
\__|  \__| \______/ \__| \______/ \_______/           \_/     \______/ \__|\______/ \__|\______/       
																											   
		
$$$$$$$$\            $$\ $$\ $$\           $$\        $$\     
\__$$  __|           \__|$$ |\__|          $$ |       $$ |    
   $$ |$$\  $$\  $$\ $$\ $$ |$$\  $$$$$$\  $$$$$$$\ $$$$$$\   
   $$ |$$ | $$ | $$ |$$ |$$ |$$ |$$  __$$\ $$  __$$\\_$$  _|  
   $$ |$$ | $$ | $$ |$$ |$$ |$$ |$$ /  $$ |$$ |  $$ | $$ |    
   $$ |$$ | $$ | $$ |$$ |$$ |$$ |$$ |  $$ |$$ |  $$ | $$ |$$\ 
   $$ |\$$$$$\$$$$  |$$ |$$ |$$ |\$$$$$$$ |$$ |  $$ | \$$$$  |
   \__| \_____\____/ \__|\__|\__| \____$$ |\__|  \__|  \____/ 
								 $$\   $$ |                   
								 \$$$$$$  |                   
								  \______/                    
														 
  $$\ $$\   $$\       $$$$$$$$\ $$\       
  $$ \$$ \  $$ |      $$  _____|$$ |      
$$$$$$$$$$\ $$ |      $$ |      $$ |      
\_$$  $$   |$$ |      $$$$$\    $$ |      
$$$$$$$$$$\ $$ |      $$  __|   $$ |      
\_$$  $$  _|$$ |      $$ |      $$ |      
  $$ |$$ |  $$$$$$$$\ $$ |      $$$$$$$$\ 
  \__|\__|  \________|\__|      \________|
`)

		return newVersionMap, nil
	}
}

func migrateTaxParams(ctx sdk.Context, paramsKeepers paramskeeper.Keeper, storeKey storetypes.StoreKey, codec codec.Codec) error {
	store := ctx.KVStore(storeKey)
	var currParams taxtypes.Params
	subspace, _ := paramsKeepers.GetSubspace(taxtypes.StoreKey)
	subspace.GetParamSet(ctx, &currParams)

	if err := currParams.Validate(); err != nil {
		return err
	}

	bz := codec.MustMarshal(&currParams)
	store.Set(taxtypes.ParamsKey, bz)
	return nil
}

func migrateMintParams(ctx sdk.Context, paramsKeepers paramskeeper.Keeper, storeKey storetypes.StoreKey, codec codec.Codec) error {
	store := ctx.KVStore(storeKey)
	var currParams minttypes.Params
	subspace, _ := paramsKeepers.GetSubspace(minttypes.StoreKey)
	subspace.GetParamSet(ctx, &currParams)

	if err := currParams.Validate(); err != nil {
		return err
	}

	bz := codec.MustMarshal(&currParams)
	store.Set(minttypes.ParamsKey, bz)
	return nil
}

func migrateFeeRefunderParams(ctx sdk.Context, paramsKeepers paramskeeper.Keeper, storeKey storetypes.StoreKey, codec codec.Codec) error {
	store := ctx.KVStore(storeKey)
	var currParams feerefundertypes.Params
	subspace, _ := paramsKeepers.GetSubspace(feerefundertypes.StoreKey)
	subspace.GetParamSet(ctx, &currParams)

	if err := currParams.Validate(); err != nil {
		return err
	}

	bz := codec.MustMarshal(&currParams)
	store.Set(feerefundertypes.ParamsKey, bz)
	return nil
}

func migrateInterchainQueriesParams(ctx sdk.Context, paramsKeepers paramskeeper.Keeper, storeKey storetypes.StoreKey, codec codec.Codec) error {
	store := ctx.KVStore(storeKey)
	var currParams icqtypes.Params
	subspace, _ := paramsKeepers.GetSubspace(icqtypes.StoreKey)
	subspace.GetParamSet(ctx, &currParams)

	currParams.QueryDeposit = sdk.NewCoins(sdk.NewCoin(params.BaseCoinUnit, sdk.NewInt(1_000_000)))

	if err := currParams.Validate(); err != nil {
		return err
	}

	bz := codec.MustMarshal(&currParams)
	store.Set(icqtypes.ParamsKey, bz)
	return nil
}

func setInterchainTxsParams(ctx sdk.Context, paramsKeepers paramskeeper.Keeper, storeKey, wasmStoreKey storetypes.StoreKey, codec codec.Codec) error {
	store := ctx.KVStore(storeKey)
	var currParams interchaintxstypes.Params
	subspace, _ := paramsKeepers.GetSubspace(interchaintxstypes.StoreKey)
	subspace.GetParamSet(ctx, &currParams)
	currParams.RegisterFee = interchaintxstypes.DefaultRegisterFee

	if err := currParams.Validate(); err != nil {
		return err
	}

	bz := codec.MustMarshal(&currParams)
	store.Set(interchaintxstypes.ParamsKey, bz)

	// set an extremely high number for the first code id that will be taxed with fee for opening an ICA
	// these bytes are equal to 1684300900 when parsed with sdk.BigEndianToUint64. It's a number that is unlikely to be reached as a code id
	// if we decide to charge a fee for opening an ICA, we can set this to a lower number in the future
	bytesIcaRegistrationFirstCode := []byte{0, 0, 0, 0, 100, 100, 100, 100}
	store.Set(interchaintxstypes.ICARegistrationFeeFirstCodeID, bytesIcaRegistrationFirstCode)
	return nil
}

func setContractManagerParams(ctx sdk.Context, keeper contractmanagerkeeper.Keeper) error {
	sudoGasLimit := uint64(3_000_000)
	cmParams := contractmanagertypes.Params{
		SudoCallGasLimit: sudoGasLimit,
	}
	return keeper.SetParams(ctx, cmParams)
}

func setInitialMinCommissionRate(ctx sdk.Context, keepers *keepers.AppKeepers) error {
	stakingParams := keepers.StakingKeeper.GetParams(ctx)
	// DefaultMinComission rate is 0%
	stakingParams.MinCommissionRate = stakingtypes.DefaultMinCommissionRate
	if err := keepers.StakingKeeper.SetParams(ctx, stakingParams); err != nil {
		return fmt.Errorf("failed to set MinCommissionRate to 5%%: %w", err)
	}

	return nil
}

func setMinInitialDepositRatio(ctx sdk.Context, keepers *keepers.AppKeepers) error {
	govParams := keepers.GovKeeper.GetParams(ctx)
	govParams.MinInitialDepositRatio = "0.25"
	return keepers.GovKeeper.SetParams(ctx, govParams)
}
