package v05

import (
	"fmt"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	icacontrollertypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/controller/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	ibcexported "github.com/cosmos/ibc-go/v7/modules/core/exported"

	minttypes "github.com/Nolus-Protocol/nolus-core/x/mint/types"
	taxmoduletypes "github.com/Nolus-Protocol/nolus-core/x/tax/types"
	vestingstypes "github.com/Nolus-Protocol/nolus-core/x/vestings/types"
)

func getLegacySubspaces(paramsKeeper *paramskeeper.Keeper) paramstypes.Subspace {
	// Set param key table for params module migration
	for _, subspace := range paramsKeeper.GetSubspaces() {
		subspace := subspace

		var keyTable paramstypes.KeyTable
		switch subspace.Name() {
		case authtypes.ModuleName:
			keyTable = authtypes.ParamKeyTable() //nolint:staticcheck
		case taxmoduletypes.ModuleName:
			keyTable = taxmoduletypes.ParamKeyTable() //nolint:staticcheck
		case banktypes.ModuleName:
			keyTable = banktypes.ParamKeyTable() //nolint:staticcheck
		case stakingtypes.ModuleName:
			keyTable = stakingtypes.ParamKeyTable()
		case minttypes.ModuleName:
			keyTable = minttypes.ParamKeyTable() //nolint:staticcheck
		case distrtypes.ModuleName:
			keyTable = distrtypes.ParamKeyTable() //nolint:staticcheck
		case slashingtypes.ModuleName:
			keyTable = slashingtypes.ParamKeyTable() //nolint:staticcheck
		case crisistypes.ModuleName:
			keyTable = crisistypes.ParamKeyTable() //nolint:staticcheck
		case govtypes.ModuleName:
			keyTable = govv1.ParamKeyTable() //nolint:staticcheck
		// wasm
		case wasmtypes.ModuleName:
			keyTable = wasmtypes.ParamKeyTable() //nolint:staticcheck
		case icacontrollertypes.SubModuleName:
			keyTable = icacontrollertypes.ParamKeyTable() //nolint:staticcheck
		// these modules have not migrated away from x/params or have no params to migrate
		case ibcexported.ModuleName,
			ibctransfertypes.ModuleName,
			vestingstypes.ModuleName:
			continue
		default:
			panic(fmt.Sprintf("param key table not set for subspace: %s", subspace.Name()))
		}

		if !subspace.HasKeyTable() {
			subspace.WithKeyTable(keyTable)
		}
	}

	return paramsKeeper.Subspace(baseapp.Paramspace).WithKeyTable(paramstypes.ConsensusParamsKeyTable())
}
