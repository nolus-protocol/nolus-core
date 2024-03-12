package wasmbinding

import (
	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkquery "github.com/cosmos/cosmos-sdk/types/query"

	"github.com/Nolus-Protocol/nolus-core/wasmbinding/bindings"

	errorsmod "cosmossdk.io/errors"

	contractmanagertypes "github.com/neutron-org/neutron/v3/x/contractmanager/types"
	"github.com/neutron-org/neutron/v3/x/interchainqueries/types"
	icatypes "github.com/neutron-org/neutron/v3/x/interchaintxs/types"
)

func (qp *QueryPlugin) GetInterchainQueryResult(ctx sdk.Context, queryID uint64) (*bindings.QueryRegisteredQueryResultResponse, error) {
	grpcResp, err := qp.icqKeeper.GetQueryResultByID(ctx, queryID)
	if err != nil {
		return nil, err
	}
	resp := bindings.QueryResult{
		KvResults: make([]*bindings.StorageValue, 0, len(grpcResp.KvResults)),
		Height:    grpcResp.GetHeight(),
		Revision:  grpcResp.GetRevision(),
	}
	for _, grpcKv := range grpcResp.GetKvResults() {
		kv := bindings.StorageValue{
			StoragePrefix: grpcKv.GetStoragePrefix(),
			Key:           grpcKv.GetKey(),
			Value:         grpcKv.GetValue(),
		}
		resp.KvResults = append(resp.KvResults, &kv)
	}

	return &bindings.QueryRegisteredQueryResultResponse{Result: &resp}, nil
}

func (qp *QueryPlugin) GetInterchainAccountAddress(ctx sdk.Context, req *bindings.QueryInterchainAccountAddressRequest) (*bindings.QueryInterchainAccountAddressResponse, error) {
	grpcReq := icatypes.QueryInterchainAccountAddressRequest{
		OwnerAddress:        req.OwnerAddress,
		InterchainAccountId: req.InterchainAccountID,
		ConnectionId:        req.ConnectionID,
	}

	grpcResp, err := qp.icaControllerKeeper.InterchainAccountAddress(sdk.WrapSDKContext(ctx), &grpcReq)
	if err != nil {
		return nil, err
	}

	return &bindings.QueryInterchainAccountAddressResponse{InterchainAccountAddress: grpcResp.GetInterchainAccountAddress()}, nil
}

func (qp *QueryPlugin) GetRegisteredInterchainQueries(ctx sdk.Context, query *bindings.QueryRegisteredQueriesRequest) (*bindings.QueryRegisteredQueriesResponse, error) {
	grpcResp, err := qp.icqKeeper.GetRegisteredQueries(ctx, &types.QueryRegisteredQueriesRequest{
		Owners:       query.Owners,
		ConnectionId: query.ConnectionID,
		Pagination: &sdkquery.PageRequest{
			Key:        query.Pagination.Key,
			Offset:     query.Pagination.Offset,
			Limit:      query.Pagination.Limit,
			CountTotal: query.Pagination.CountTotal,
			Reverse:    query.Pagination.Reverse,
		},
	})
	if err != nil {
		return nil, err
	}

	resp := bindings.QueryRegisteredQueriesResponse{RegisteredQueries: make([]bindings.RegisteredQuery, 0, len(grpcResp.GetRegisteredQueries()))}
	for _, grpcQuery := range grpcResp.GetRegisteredQueries() {
		query := mapGRPCRegisteredQueryToWasmBindings(grpcQuery)
		resp.RegisteredQueries = append(resp.RegisteredQueries, query)
	}
	return &resp, nil
}

func (qp *QueryPlugin) GetRegisteredInterchainQuery(ctx sdk.Context, req *bindings.QueryRegisteredQueryRequest) (*bindings.QueryRegisteredQueryResponse, error) {
	grpcResp, err := qp.icqKeeper.GetQueryByID(ctx, req.QueryID)
	if err != nil {
		return nil, err
	}
	if grpcResp == nil {
		return nil, errorsmod.Wrapf(types.ErrEmptyResult, "interchain query response empty for query id %d", req.QueryID)
	}
	query := mapGRPCRegisteredQueryToWasmBindings(*grpcResp)

	return &bindings.QueryRegisteredQueryResponse{RegisteredQuery: &query}, nil
}

func (qp *QueryPlugin) GetMinIbcFee(ctx sdk.Context, _ *bindings.QueryMinIbcFeeRequest) (*bindings.QueryMinIbcFeeResponse, error) {
	fee := qp.feeRefunderKeeper.GetMinFee(ctx)
	return &bindings.QueryMinIbcFeeResponse{MinFee: fee}, nil
}

func (qp *QueryPlugin) GetFailures(ctx sdk.Context, address string, pagination *sdkquery.PageRequest) (*bindings.FailuresResponse, error) {
	res, err := qp.contractmanagerKeeper.AddressFailures(ctx, &contractmanagertypes.QueryFailuresRequest{
		Address:    address,
		Pagination: pagination,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get failures for address: %s", address)
	}

	return &bindings.FailuresResponse{Failures: res.Failures}, nil
}

func mapGRPCRegisteredQueryToWasmBindings(grpcQuery types.RegisteredQuery) bindings.RegisteredQuery {
	return bindings.RegisteredQuery{
		ID:                              grpcQuery.GetId(),
		Owner:                           grpcQuery.GetOwner(),
		Keys:                            grpcQuery.GetKeys(),
		TransactionsFilter:              grpcQuery.GetTransactionsFilter(),
		QueryType:                       grpcQuery.GetQueryType(),
		ConnectionID:                    grpcQuery.GetConnectionId(),
		UpdatePeriod:                    grpcQuery.GetUpdatePeriod(),
		LastSubmittedResultLocalHeight:  grpcQuery.GetLastSubmittedResultLocalHeight(),
		LastSubmittedResultRemoteHeight: grpcQuery.GetLastSubmittedResultRemoteHeight(),
		Deposit:                         grpcQuery.GetDeposit(),
		SubmitTimeout:                   grpcQuery.GetSubmitTimeout(),
		RegisteredAtHeight:              grpcQuery.GetRegisteredAtHeight(),
	}
}
