package wasmbinding

import (
	contractmanagerkeeper "github.com/Nolus-Protocol/nolus-core/x/contractmanager/keeper"
	feerefunderkeeper "github.com/Nolus-Protocol/nolus-core/x/feerefunder/keeper"
	icacontrollerkeeper "github.com/Nolus-Protocol/nolus-core/x/interchaintxs/keeper"
)

type QueryPlugin struct {
	icaControllerKeeper   *icacontrollerkeeper.Keeper
	feeRefunderKeeper     *feerefunderkeeper.Keeper
	contractmanagerKeeper *contractmanagerkeeper.Keeper
}

// NewQueryPlugin returns a reference to a new QueryPlugin.
func NewQueryPlugin(
	icaControllerKeeper *icacontrollerkeeper.Keeper,
	feeRefunderKeeper *feerefunderkeeper.Keeper,
	contractmanagerKeeper *contractmanagerkeeper.Keeper,
) *QueryPlugin {
	return &QueryPlugin{
		icaControllerKeeper:   icaControllerKeeper,
		feeRefunderKeeper:     feeRefunderKeeper,
		contractmanagerKeeper: contractmanagerKeeper,
	}
}
