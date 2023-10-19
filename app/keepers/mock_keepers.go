package keepers

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	feeburnertypes "github.com/neutron-org/neutron/x/feeburner/types"
)

// FeeBurnerKeeperExpectedKeeper is a mock of the FeeBurnerKeeper interface that is going to be used in the interchaintxs module.
// In the interchaintxs  module, the FeeBurnerKeeper is only used to get the treasury address
// from it's params when charging fees for creating interchain account.
type FeeBurnerKeeperExpectedKeeper struct{}

func NewFeeBurnerExpectedKeeper() FeeBurnerKeeperExpectedKeeper {
	return FeeBurnerKeeperExpectedKeeper{}
}

func (s FeeBurnerKeeperExpectedKeeper) GetParams(ctx sdk.Context) feeburnertypes.Params {
	// TODO: ensure correct params
	return feeburnertypes.NewParams("unls", "nolus14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9s0k0puz")
}
