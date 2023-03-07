package v0

import (
	"github.com/Nolus-Protocol/nolus-core/app/upgrades"

	store "github.com/cosmos/cosmos-sdk/store/types"
	interchainqueries "github.com/neutron-org/neutron/x/interchainqueries/types"
	interchaintxs "github.com/neutron-org/neutron/x/interchaintxs/types"
)

// TODO Start using this method to upgrade the app after export app state is fixed.
const (
	// UpgradeName defines the on-chain upgrade name.
	UpgradeName = "v0.2.1"
)

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: store.StoreUpgrades{
		Added: []string{
			interchainqueries.ModuleName,
			interchaintxs.ModuleName,
		},
	},
}
