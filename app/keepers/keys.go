package keepers

import (
	storetypes "cosmossdk.io/store/types"
	authztkeeper "cosmossdk.io/x/authz/keeper"
	banktypes "cosmossdk.io/x/bank/types"
	consensusparamtypes "cosmossdk.io/x/consensus/types"
	crisistypes "cosmossdk.io/x/crisis/types"
	distrtypes "cosmossdk.io/x/distribution/types"
	evidencetypes "cosmossdk.io/x/evidence/types"
	"cosmossdk.io/x/feegrant"
	govtypes "cosmossdk.io/x/gov/types"
	paramstypes "cosmossdk.io/x/params/types"
	slashingtypes "cosmossdk.io/x/slashing/types"
	stakingtypes "cosmossdk.io/x/staking/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"

	icacontrollertypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/controller/types"
	icahosttypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/host/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	ibcexported "github.com/cosmos/ibc-go/v8/modules/core/exported"

	minttypes "github.com/Nolus-Protocol/nolus-core/x/mint/types"
	taxmoduletypes "github.com/Nolus-Protocol/nolus-core/x/tax/typesv2"
	vestingstypes "github.com/Nolus-Protocol/nolus-core/x/vestings/types"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	contractmanagermoduletypes "github.com/Nolus-Protocol/nolus-core/x/contractmanager/types"
	feerefundertypes "github.com/Nolus-Protocol/nolus-core/x/feerefunder/types"
	interchaintxstypes "github.com/Nolus-Protocol/nolus-core/x/interchaintxs/types"
)

func (appKeepers *AppKeepers) GenerateKeys() {
	// Define what keys will be used in the cosmos-sdk key/value store.
	// Cosmos-SDK modules each have a "key" that allows the application to reference what they've stored on the chain.
	appKeepers.keys = storetypes.NewKVStoreKeys(
		authtypes.StoreKey,
		authztkeeper.StoreKey,
		banktypes.StoreKey,
		stakingtypes.StoreKey,
		crisistypes.StoreKey,
		minttypes.StoreKey,
		distrtypes.StoreKey,
		slashingtypes.StoreKey,
		govtypes.StoreKey,
		paramstypes.StoreKey,
		ibcexported.StoreKey,
		upgradetypes.StoreKey,
		evidencetypes.StoreKey,
		feegrant.StoreKey,
		ibctransfertypes.StoreKey,
		taxmoduletypes.StoreKey,
		vestingstypes.StoreKey,
		icacontrollertypes.StoreKey,
		icahosttypes.StoreKey,
		capabilitytypes.StoreKey,
		contractmanagermoduletypes.StoreKey,
		interchaintxstypes.StoreKey,
		wasmtypes.StoreKey,
		feerefundertypes.StoreKey,
		consensusparamtypes.StoreKey,
	)

	// Define transient store keys
	appKeepers.tkeys = storetypes.NewTransientStoreKeys(paramstypes.TStoreKey)

	// MemKeys are for information that is stored only in RAM.
	appKeepers.memKeys = storetypes.NewMemoryStoreKeys(capabilitytypes.MemStoreKey, feerefundertypes.MemStoreKey)
}

func (appKeepers *AppKeepers) GetKVStoreKeys() map[string]*storetypes.KVStoreKey {
	return appKeepers.keys
}

func (appKeepers *AppKeepers) GetTransientStoreKey() map[string]*storetypes.TransientStoreKey {
	return appKeepers.tkeys
}

func (appKeepers *AppKeepers) GetMemoryStoreKey() map[string]*storetypes.MemoryStoreKey {
	return appKeepers.memKeys
}

// GetKey returns the KVStoreKey for the provided store key.
//
// NOTE: This is solely to be used for testing purposes.
func (appKeepers *AppKeepers) GetKey(storeKey string) *storetypes.KVStoreKey {
	return appKeepers.keys[storeKey]
}

// GetTKey returns the TransientStoreKey for the provided store key.
//
// NOTE: This is solely to be used for testing purposes.
func (appKeepers *AppKeepers) GetTKey(storeKey string) *storetypes.TransientStoreKey {
	return appKeepers.tkeys[storeKey]
}

// GetMemKey returns the MemStoreKey for the provided mem key.
//
// NOTE: This is solely used for testing purposes.
func (appKeepers *AppKeepers) GetMemKey(storeKey string) *storetypes.MemoryStoreKey {
	return appKeepers.memKeys[storeKey]
}
