package v068

import (
	store "cosmossdk.io/store/types"
	"github.com/Nolus-Protocol/nolus-core/app/upgrades"

	icqtypes "github.com/neutron-org/neutron/v4/x/interchainqueries/types"
)

const (
	// UpgradeName defines the on-chain upgrades name.
	UpgradeName = "v0.6.8"
)

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: store.StoreUpgrades{
		Added:   []string{},
		Deleted: []string{icqtypes.ModuleName},
	},
}
