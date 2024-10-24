package wasmbinding

import (
	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkquery "github.com/cosmos/cosmos-sdk/types/query"

	"github.com/Nolus-Protocol/nolus-core/wasmbinding/bindings"

	contractmanagertypes "github.com/Nolus-Protocol/nolus-core/x/contractmanager/types"
	icatypes "github.com/Nolus-Protocol/nolus-core/x/interchaintxs/types"
)

func (qp *QueryPlugin) GetInterchainAccountAddress(ctx sdk.Context, req *bindings.QueryInterchainAccountAddressRequest) (*bindings.QueryInterchainAccountAddressResponse, error) {
	grpcReq := icatypes.QueryInterchainAccountAddressRequest{
		OwnerAddress:        req.OwnerAddress,
		InterchainAccountId: req.InterchainAccountID,
		ConnectionId:        req.ConnectionID,
	}

	grpcResp, err := qp.icaControllerKeeper.InterchainAccountAddress(ctx, &grpcReq)
	if err != nil {
		return nil, err
	}

	return &bindings.QueryInterchainAccountAddressResponse{InterchainAccountAddress: grpcResp.GetInterchainAccountAddress()}, nil
}

func (qp *QueryPlugin) GetMinIbcFee(ctx sdk.Context, _ *bindings.QueryMinIbcFeeRequest) (*bindings.QueryMinIbcFeeResponse, error) {
	fee := qp.feeRefunderKeeper.GetMinFee(ctx)
	return &bindings.QueryMinIbcFeeResponse{MinFee: fee}, nil
}

func (qp *QueryPlugin) GetFailures(ctx sdk.Context, address string, pagination *sdkquery.PageRequest) (*bindings.FailuresResponse, error) {
	res, err := qp.contractmanagerKeeper.AddressFailures(ctx, &contractmanagertypes.QueryFailuresByAddressRequest{
		Address:    address,
		Pagination: pagination,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get failures for address: %s", address)
	}

	return &bindings.FailuresResponse{Failures: res.Failures}, nil
}
