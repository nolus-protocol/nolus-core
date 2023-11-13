package keepers

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	authztkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	icacontrollertypes "github.com/cosmos/ibc-go/v4/modules/apps/27-interchain-accounts/controller/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v4/modules/apps/transfer/types"
	ibchost "github.com/cosmos/ibc-go/v4/modules/core/24-host"

	minttypes "github.com/Nolus-Protocol/nolus-core/x/mint/types"
	taxmoduletypes "github.com/Nolus-Protocol/nolus-core/x/tax/types"
	vestingstypes "github.com/Nolus-Protocol/nolus-core/x/vestings/types"

	"github.com/CosmWasm/wasmd/x/wasm"

	contractmanagermoduletypes "github.com/neutron-org/neutron/x/contractmanager/types"
	feerefundertypes "github.com/neutron-org/neutron/x/feerefunder/types"
	interchainqueriestypes "github.com/neutron-org/neutron/x/interchainqueries/types"
	interchaintxstypes "github.com/neutron-org/neutron/x/interchaintxs/types"
)

func (appKeepers *AppKeepers) GenerateKeys() {
	// Define what keys will be used in the cosmos-sdk key/value store.
	// Cosmos-SDK modules each have a "key" that allows the application to reference what they've stored on the chain.
	appKeepers.keys = sdk.NewKVStoreKeys(
		authtypes.StoreKey, authztkeeper.StoreKey, banktypes.StoreKey, stakingtypes.StoreKey,
		minttypes.StoreKey, distrtypes.StoreKey, slashingtypes.StoreKey,
		govtypes.StoreKey, paramstypes.StoreKey, ibchost.StoreKey,
		upgradetypes.StoreKey, evidencetypes.StoreKey, ibctransfertypes.StoreKey,
		taxmoduletypes.StoreKey, vestingstypes.StoreKey, icacontrollertypes.StoreKey, capabilitytypes.StoreKey,
		interchainqueriestypes.StoreKey, contractmanagermoduletypes.StoreKey, interchaintxstypes.StoreKey,
		wasm.StoreKey, feerefundertypes.StoreKey,
	)

	// Define transient store keys
	appKeepers.tkeys = sdk.NewTransientStoreKeys(paramstypes.TStoreKey)

	// MemKeys are for information that is stored only in RAM.
	appKeepers.memKeys = sdk.NewMemoryStoreKeys(capabilitytypes.MemStoreKey, feerefundertypes.MemStoreKey)
}

func (appKeepers *AppKeepers) GetKVStoreKey() map[string]*sdk.KVStoreKey {
	return appKeepers.keys
}

func (appKeepers *AppKeepers) GetTransientStoreKey() map[string]*sdk.TransientStoreKey {
	return appKeepers.tkeys
}

func (appKeepers *AppKeepers) GetMemoryStoreKey() map[string]*sdk.MemoryStoreKey {
	return appKeepers.memKeys
}

// GetKey returns the KVStoreKey for the provided store key.
//
// NOTE: This is solely to be used for testing purposes.
func (appKeepers *AppKeepers) GetKey(storeKey string) *sdk.KVStoreKey {
	return appKeepers.keys[storeKey]
}

// GetTKey returns the TransientStoreKey for the provided store key.
//
// NOTE: This is solely to be used for testing purposes.
func (appKeepers *AppKeepers) GetTKey(storeKey string) *sdk.TransientStoreKey {
	return appKeepers.tkeys[storeKey]
}

// GetMemKey returns the MemStoreKey for the provided mem key.
//
// NOTE: This is solely used for testing purposes.
func (appKeepers *AppKeepers) GetMemKey(storeKey string) *sdk.MemoryStoreKey {
	return appKeepers.memKeys[storeKey]
}
