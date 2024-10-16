package wasmbinding

import (
	contractmanagerkeeper "github.com/Nolus-Protocol/nolus-core/x/contractmanager/keeper"
	feerefunderkeeper "github.com/Nolus-Protocol/nolus-core/x/feerefunder/keeper"
	icqkeeper "github.com/Nolus-Protocol/nolus-core/x/interchainqueries/keeper"
	icacontrollerkeeper "github.com/Nolus-Protocol/nolus-core/x/interchaintxs/keeper"
)

type QueryPlugin struct {
	icaControllerKeeper   *icacontrollerkeeper.Keeper
	icqKeeper             *icqkeeper.Keeper
	feeRefunderKeeper     *feerefunderkeeper.Keeper
	contractmanagerKeeper *contractmanagerkeeper.Keeper
}

// NewQueryPlugin returns a reference to a new QueryPlugin.
func NewQueryPlugin(
	icaControllerKeeper *icacontrollerkeeper.Keeper,
	icqKeeper *icqkeeper.Keeper,
	feeRefunderKeeper *feerefunderkeeper.Keeper,
	contractmanagerKeeper *contractmanagerkeeper.Keeper,
) *QueryPlugin {
	return &QueryPlugin{
		icaControllerKeeper:   icaControllerKeeper,
		icqKeeper:             icqKeeper,
		feeRefunderKeeper:     feeRefunderKeeper,
		contractmanagerKeeper: contractmanagerKeeper,
	}
}
