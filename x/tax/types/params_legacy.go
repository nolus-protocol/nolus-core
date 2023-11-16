/*
NOTE: Usage of x/params to manage parameters is deprecated in favor of x/gov
controlled execution of MsgUpdateParams messages. These types remains solely
for migration purposes and will be removed in a future release.
*/
package types

import (
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Parameter store keys.
var (
	KeyFeeRate         = []byte("FeeRate")
	KeyBaseDenom       = []byte("BaseDenom")
	KeyContractAddress = []byte("ContractAddress")
)

// ParamTable for tax module.
//
// NOTE: Deprecated.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// Implements params.ParamSet
//
// NOTE: Deprecated.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyFeeRate, &p.FeeRate, validateFeeRate),
		paramtypes.NewParamSetPair(KeyContractAddress, &p.ContractAddress, validateContractAddress),
		paramtypes.NewParamSetPair(KeyBaseDenom, &p.BaseDenom, validateBaseDenom),
	}
}
