package keeper

// import (
// 	"github.com/Nolus-Protocol/nolus-core/x/mint/types"
// 	abci "github.com/cometbft/cometbft/abci/types"

// 	"github.com/cosmos/cosmos-sdk/codec"
// 	sdk "github.com/cosmos/cosmos-sdk/types"
// 	"cosmossdk.io/errors"
// )

// refactor: querier is deprecated - https://github.com/cosmos/cosmos-sdk/blob/release/v0.47.x/UPGRADING.md#appmodule-interface
// // NewQuerier returns a minting Querier handler.
// func NewQuerier(k Keeper, legacyQuerierCdc *codec.LegacyAmino) sdk.Querier {
// 	return func(ctx sdk.Context, path []string, _ abci.RequestQuery) ([]byte, error) {
// 		switch path[0] {
// 		case types.QueryParameters:
// 			return queryParams(ctx, k, legacyQuerierCdc)

// 		case types.QueryMintState:
// 			return queryMintState(ctx, k, legacyQuerierCdc)

// 		default:
// 			return nil, errors.Wrapf(sdkerrors.ErrUnknownRequest, "unknown query path: %s", path[0])
// 		}
// 	}
// }

// func queryParams(ctx sdk.Context, k Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
// 	params := k.GetParams(ctx)

// 	res, err := codec.MarshalJSONIndent(legacyQuerierCdc, params)
// 	if err != nil {
// 		return nil, errors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
// 	}

// 	return res, nil
// }

// func queryMintState(ctx sdk.Context, k Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
// 	minter := k.GetMinter(ctx)

// 	minterState := types.QueryMintStateResponse{NormTimePassed: minter.NormTimePassed, TotalMinted: minter.TotalMinted}
// 	res, err := codec.MarshalJSONIndent(legacyQuerierCdc, minterState)
// 	if err != nil {
// 		return nil, errors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
// 	}

// 	return res, nil
// }
