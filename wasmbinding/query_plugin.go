package wasmbinding

import (
	contractmanagerkeeper "github.com/neutron-org/neutron/x/contractmanager/keeper"
	feerefunderkeeper "github.com/neutron-org/neutron/x/feerefunder/keeper"
	icqkeeper "github.com/neutron-org/neutron/x/interchainqueries/keeper"
	icacontrollerkeeper "github.com/neutron-org/neutron/x/interchaintxs/keeper"
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
