package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	types "github.com/Nolus-Protocol/nolus-core/solanacarrier"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) Params(ctx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	params, err := k.GetParams(ctx)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return &types.QueryParamsResponse{Params: params}, nil
}
