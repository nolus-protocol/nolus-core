package bindings

import (
	"encoding/json"

	contractmanagertypes "github.com/Nolus-Protocol/nolus-core/x/contractmanager/types"
	feerefundertypes "github.com/Nolus-Protocol/nolus-core/x/feerefunder/types"

	"github.com/cosmos/cosmos-sdk/types/query"
	//nolint:staticcheck
)

// NolusQuery contains nolus custom queries.
type NolusQuery struct {
	// Interchain account address for specified ConnectionID and OwnerAddress
	InterchainAccountAddress *QueryInterchainAccountAddressRequest `json:"interchain_account_address,omitempty"`
	// MinIbcFee
	MinIbcFee *QueryMinIbcFeeRequest `json:"min_ibc_fee,omitempty"`
	// Contractmanager queries
	// Query all failures for address
	Failures *Failures `json:"failures,omitempty"`
}

/* Requests */

type QueryInterchainAccountAddressRequest struct {
	// owner_address is the owner of the interchain account on the controller chain
	OwnerAddress string `json:"owner_address,omitempty"`
	// interchain_account_id is an identifier of your interchain account from which you want to execute msgs
	InterchainAccountID string `json:"interchain_account_id,omitempty"`
	// connection_id is an IBC connection identifier between Nolus and remote chain
	ConnectionID string `json:"connection_id,omitempty"`
}

type QueryMinIbcFeeRequest struct{}

type QueryMinIbcFeeResponse struct {
	MinFee feerefundertypes.Fee `json:"min_fee"`
}

// Query response for an interchain account address.
type QueryInterchainAccountAddressResponse struct {
	// The corresponding interchain account address on the host chain
	InterchainAccountAddress string `json:"interchain_account_address,omitempty"`
}

type StorageValue struct {
	StoragePrefix string `json:"storage_prefix,omitempty"`
	Key           []byte `json:"key"`
	Value         []byte `json:"value"`
}

func (sv StorageValue) MarshalJSON() ([]byte, error) {
	type AliasSV StorageValue

	a := struct {
		AliasSV
	}{
		AliasSV: (AliasSV)(sv),
	}

	// We want Key and Value be as empty arrays in Json ('[]'), not 'null'
	// It's easier to work with on smart-contracts side
	if a.Key == nil {
		a.Key = make([]byte, 0)
	}
	if a.Value == nil {
		a.Value = make([]byte, 0)
	}

	return json.Marshal(a)
}

type Failures struct {
	Address    string             `json:"address"`
	Pagination *query.PageRequest `json:"pagination,omitempty"`
}

type FailuresResponse struct {
	Failures []contractmanagertypes.Failure `json:"failures"`
}
