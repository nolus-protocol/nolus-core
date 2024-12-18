package v4

import (
	storetypes "cosmossdk.io/core/store"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	legacytypes "github.com/Nolus-Protocol/nolus-core/x/tax/types"
	types "github.com/Nolus-Protocol/nolus-core/x/tax/typesv2"
)

const (
	ModuleName = "tax"
)

var DexFeeParams = []*types.DexFeeParams{
	{
		ProfitAddress: "nolus1r69jl4n2hp6vd4ex7xx5l9rcq8qcjeh8fefauzgvpnz2e0khqe9qnw25u4",
		AcceptedDenomsMinPrices: []*types.DenomPrice{
			{
				Denom:    "ibc/F5FABF52B54E65064B57BF6DBD8E5FAD22CEE9F4B8A57ADBB20CCD0173AA72A4",
				Ticker:   "USDC_NOBLE",
				MinPrice: 0.025,
			},
			{
				Denom:    "ibc/6CDD4663F2F09CD62285E2D45891FC149A3568E316CE3EBBE201A71A78A69388",
				Ticker:   "ATOM",
				MinPrice: 0.0029,
			},
			{
				Denom:    "ibc/ED07A3391A112B175915CD8FAF43A2DA8E4790EDE12566649D0C2F97716B8518",
				Ticker:   "OSMO",
				MinPrice: 0.044,
			},
		},
	},
	{
		ProfitAddress: "nolus14qnh6egte0yufj2c5cunjxmdu995zkx8x8n24nw8wxx3rl992wpqugtkwh",
		AcceptedDenomsMinPrices: []*types.DenomPrice{
			{
				Denom:    "ibc/18161D8EFBD00FF5B7683EF8E923B8913453567FBE3FB6672D75712B0DEB6682",
				Ticker:   "USDC_NOBLE",
				MinPrice: 0.025,
			},
			{
				Denom:    "ibc/3D6BC6E049CAEB905AC97031A42800588C58FB471EBDC7A3530FFCD0C3DC9E09",
				Ticker:   "NTRN",
				MinPrice: 0.05,
			},
		},
	},
}

// Migrate migrates the x/tax module state from the consensus version 3 to
// version 4. Specifically, it takes the parameters that are currently stored and managed by the x/tax module
// and migrates them to the new v2 format of the x/tax module params.
func Migrate(
	ctx sdk.Context,
	store storetypes.KVStore,
	cdc codec.BinaryCodec,
) error {
	bz, err := store.Get(types.ParamsKey)
	if err != nil {
		return err
	}
	if bz == nil {
		return nil
	}
	var currentParams legacytypes.Params // nolint:staticcheck
	err = cdc.Unmarshal(bz, &currentParams)
	if err != nil {
		return err
	}

	newParams := types.Params{
		FeeRate:         currentParams.FeeRate,
		BaseDenom:       currentParams.BaseDenom,
		TreasuryAddress: currentParams.ContractAddress,
		DexFeeParams:    DexFeeParams,
	}

	// validate and set the new parameters
	if err := newParams.Validate(); err != nil {
		return err
	}

	bz, err = cdc.Marshal(&newParams)
	if err != nil {
		return err
	}

	return store.Set(types.ParamsKey, bz)
}
