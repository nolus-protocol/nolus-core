package keeper

import (
	"context"

	"cosmossdk.io/errors"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	types "github.com/Nolus-Protocol/nolus-core/solanacarrier"
)

var _ types.MsgServer = msgServer{}

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the x/solanacarrier MsgServer interface.
func NewMsgServerImpl(k Keeper) types.MsgServer {
	return &msgServer{
		Keeper: k,
	}
}

func (ms msgServer) UpdateParams(ctx context.Context, req *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	if err := req.ValidateBasic(); err != nil {
		return nil, err
	}

	if ms.authority != req.Authority {
		return nil, errors.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", ms.authority, req.Authority)
	}

	if err := ms.SetParams(ctx, req.Params); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}
