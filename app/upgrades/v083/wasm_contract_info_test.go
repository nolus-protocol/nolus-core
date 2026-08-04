package v083_test

import (
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/require"

	storetypes "cosmossdk.io/store/types"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/gogoproto/proto"

	v083 "github.com/Nolus-Protocol/nolus-core/app/upgrades/v083"
)

// The leaser contract as queried from mainnet while the chain ran wasmd v0.61.1.
const (
	leaserAddr    = "nolus1wn625s4jcmvk0szpl85rj5azkfc6suyvf75q6vrddscjdphtve8s5gg42f"
	leaserAdmin   = "nolus1gurgpv8savnfw66lckwzn4zk7fp394lpe667dhu7aw48u40lj6jsqxf8nd"
	leaserIBC2Prt = "wasm21wn625s4jcmvk0szpl85rj5azkfc6suyvf75q6vrddscjdphtve8s5gg42f"
)

func mainnetLeaser() wasmtypes.ContractInfo {
	return wasmtypes.ContractInfo{
		CodeID:     906,
		Creator:    leaserAdmin,
		Admin:      leaserAdmin,
		Label:      "leaser",
		Created:    &wasmtypes.AbsoluteTxPosition{BlockHeight: 0, TxIndex: 5830874},
		IBCPortID:  "",
		IBC2PortID: leaserIBC2Prt,
	}
}

// encodeLegacy encodes a ContractInfo the way wasmd v0.61.1 did, with ibc2_port_id
// under field 7. Field order matches gogoproto's ascending output.
func encodeLegacy(t *testing.T, info wasmtypes.ContractInfo) []byte {
	t.Helper()
	require.Nil(t, info.Extension, "v0.61.1 stored extension at field 8; fixtures never set it")

	var bz []byte
	appendTag := func(field, wireType int) {
		bz = binary.AppendUvarint(bz, uint64(field)<<3|uint64(wireType))
	}
	appendString := func(field int, s string) {
		if s == "" {
			return
		}
		appendTag(field, 2)
		bz = binary.AppendUvarint(bz, uint64(len(s)))
		bz = append(bz, s...)
	}

	if info.CodeID != 0 {
		appendTag(1, 0)
		bz = binary.AppendUvarint(bz, info.CodeID)
	}
	appendString(2, info.Creator)
	appendString(3, info.Admin)
	appendString(4, info.Label)
	if info.Created != nil {
		nested, err := proto.Marshal(info.Created)
		require.NoError(t, err)
		appendTag(5, 2)
		bz = binary.AppendUvarint(bz, uint64(len(nested)))
		bz = append(bz, nested...)
	}
	appendString(6, info.IBCPortID)
	appendString(legacyIBC2PortIDField, info.IBC2PortID)

	return bz
}

const (
	legacyIBC2PortIDField = 7
	ibc2PortIDField       = 8
)

func TestMigrateContractInfo_MainnetLeaser(t *testing.T) {
	expected := mainnetLeaser()
	legacy := encodeLegacy(t, expected)

	var beforeMigration wasmtypes.ContractInfo
	require.Error(t, proto.Unmarshal(legacy, &beforeMigration),
		"the v0.61.1 encoding must be undecodable by v0.61.8, otherwise there is nothing to migrate")

	migrated := migrateOne(t, legacy)

	var got wasmtypes.ContractInfo
	require.NoError(t, proto.Unmarshal(migrated, &got))
	require.Equal(t, expected, got)
	require.Equal(t, leaserIBC2Prt, got.IBC2PortID)
	require.Nil(t, got.Extension)

	canonical, err := proto.Marshal(&expected)
	require.NoError(t, err)
	require.Equal(t, canonical, migrated,
		"migrated bytes must be identical to what v0.61.8 writes natively")
}

func TestMigrateContractInfo_LeavesUntouchedRecordsAlone(t *testing.T) {
	// A contract last written before the chain ran v0.61.1 carries no field 7.
	preUpgrade := wasmtypes.ContractInfo{
		CodeID:  12,
		Creator: leaserAdmin,
		Label:   "lpp",
		Created: &wasmtypes.AbsoluteTxPosition{BlockHeight: 4200000, TxIndex: 3},
	}
	legacy := encodeLegacy(t, preUpgrade)

	ctx, storeKey, cdc := setupWasmStore(t)
	contractStore := ctx.KVStore(storeKey)
	key := wasmtypes.GetContractAddressKey([]byte("pre-upgrade-contract"))
	contractStore.Set(key, legacy)

	rewritten, err := v083.MigrateContractInfoIBC2PortID(ctx, storeKey, cdc)
	require.NoError(t, err)
	require.Equal(t, 0, rewritten)
	require.Equal(t, legacy, contractStore.Get(key), "record must be byte-for-byte unchanged")

	var got wasmtypes.ContractInfo
	require.NoError(t, cdc.Unmarshal(contractStore.Get(key), &got))
	require.Equal(t, preUpgrade, got)
}

func TestMigrateContractInfo_AlreadyMigratedIsNoOp(t *testing.T) {
	info := mainnetLeaser()
	upstream, err := proto.Marshal(&info)
	require.NoError(t, err)

	ctx, storeKey, cdc := setupWasmStore(t)
	contractStore := ctx.KVStore(storeKey)
	key := wasmtypes.GetContractAddressKey([]byte("already-migrated"))
	contractStore.Set(key, upstream)

	rewritten, err := v083.MigrateContractInfoIBC2PortID(ctx, storeKey, cdc)
	require.NoError(t, err)
	require.Equal(t, 0, rewritten)
	require.Equal(t, upstream, contractStore.Get(key))
}

func TestMigrateContractInfo_MixedStore(t *testing.T) {
	ctx, storeKey, cdc := setupWasmStore(t)
	contractStore := ctx.KVStore(storeKey)

	leaser := mainnetLeaser()
	lease := wasmtypes.ContractInfo{
		CodeID:     907,
		Creator:    leaserAddr,
		Admin:      leaserAddr,
		Label:      "lease",
		Created:    &wasmtypes.AbsoluteTxPosition{BlockHeight: 0, TxIndex: 5830900},
		IBC2PortID: "wasm21qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq",
	}
	preUpgrade := wasmtypes.ContractInfo{
		CodeID:  12,
		Creator: leaserAdmin,
		Label:   "oracle",
		Created: &wasmtypes.AbsoluteTxPosition{BlockHeight: 4200000, TxIndex: 3},
	}

	keys := map[string]wasmtypes.ContractInfo{
		"leaser":     leaser,
		"lease":      lease,
		"preUpgrade": preUpgrade,
	}
	for name, info := range keys {
		contractStore.Set(wasmtypes.GetContractAddressKey([]byte(name)), encodeLegacy(t, info))
	}

	rewritten, err := v083.MigrateContractInfoIBC2PortID(ctx, storeKey, cdc)
	require.NoError(t, err)
	require.Equal(t, 2, rewritten, "only the two records carrying an ibc2 port id are rewritten")

	for name, expected := range keys {
		var got wasmtypes.ContractInfo
		require.NoError(t, cdc.Unmarshal(contractStore.Get(wasmtypes.GetContractAddressKey([]byte(name))), &got), name)
		require.Equal(t, expected, got, name)
	}
}

func TestMigrateContractInfo_RejectsNonPortPayloadAtField7(t *testing.T) {
	// A genuine Extension would also sit at field 7. Nolus has never set one, so
	// rather than silently reinterpreting it as a port id the upgrade must abort.
	extension, err := proto.Marshal(&wasmtypes.AbsoluteTxPosition{BlockHeight: 1, TxIndex: 2})
	require.NoError(t, err)

	var bz []byte
	bz = binary.AppendUvarint(bz, 1<<3|0)
	bz = binary.AppendUvarint(bz, 906)
	bz = binary.AppendUvarint(bz, legacyIBC2PortIDField<<3|2)
	bz = binary.AppendUvarint(bz, uint64(len(extension)))
	bz = append(bz, extension...)

	ctx, storeKey, cdc := setupWasmStore(t)
	ctx.KVStore(storeKey).Set(wasmtypes.GetContractAddressKey([]byte("has-extension")), bz)

	_, err = v083.MigrateContractInfoIBC2PortID(ctx, storeKey, cdc)
	require.ErrorContains(t, err, "not an IBC v2 port id")
}

func TestMigrateContractInfo_RejectsRecordCarryingBothFields(t *testing.T) {
	// v0.61.1 stored extension at field 8. Nolus never set one, but if a record did
	// carry both, moving field 7 onto field 8 would silently drop one of the values.
	extension, err := proto.Marshal(&wasmtypes.AbsoluteTxPosition{BlockHeight: 1, TxIndex: 2})
	require.NoError(t, err)

	bz := encodeLegacy(t, mainnetLeaser())
	bz = binary.AppendUvarint(bz, ibc2PortIDField<<3|2)
	bz = binary.AppendUvarint(bz, uint64(len(extension)))
	bz = append(bz, extension...)

	ctx, storeKey, cdc := setupWasmStore(t)
	ctx.KVStore(storeKey).Set(wasmtypes.GetContractAddressKey([]byte("both-fields")), bz)

	_, err = v083.MigrateContractInfoIBC2PortID(ctx, storeKey, cdc)
	require.ErrorContains(t, err, "would collide")
}

func TestMigrateContractInfo_RejectsMalformedRecord(t *testing.T) {
	legacy := encodeLegacy(t, mainnetLeaser())

	ctx, storeKey, cdc := setupWasmStore(t)
	ctx.KVStore(storeKey).Set(wasmtypes.GetContractAddressKey([]byte("truncated")), legacy[:len(legacy)-4])

	_, err := v083.MigrateContractInfoIBC2PortID(ctx, storeKey, cdc)
	require.Error(t, err)
}

func TestMigrateContractInfo_NilStoreKey(t *testing.T) {
	ctx, _, cdc := setupWasmStore(t)
	_, err := v083.MigrateContractInfoIBC2PortID(ctx, nil, cdc)
	require.ErrorContains(t, err, "nil x/wasm store key")
}

func migrateOne(t *testing.T, legacy []byte) []byte {
	t.Helper()

	ctx, storeKey, cdc := setupWasmStore(t)
	contractStore := ctx.KVStore(storeKey)
	key := wasmtypes.GetContractAddressKey([]byte("contract"))
	contractStore.Set(key, legacy)

	rewritten, err := v083.MigrateContractInfoIBC2PortID(ctx, storeKey, cdc)
	require.NoError(t, err)
	require.Equal(t, 1, rewritten)

	return contractStore.Get(key)
}

func setupWasmStore(t *testing.T) (sdk.Context, *storetypes.KVStoreKey, codec.Codec) {
	t.Helper()

	storeKey := storetypes.NewKVStoreKey(wasmtypes.StoreKey)
	tKey := storetypes.NewTransientStoreKey("transient_test")
	ctx := testutil.DefaultContext(storeKey, tKey)
	cdc := moduletestutil.MakeTestEncodingConfig().Codec

	return ctx, storeKey, cdc
}
