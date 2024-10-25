package v067

import (
	store "cosmossdk.io/store/types"
	"github.com/Nolus-Protocol/nolus-core/app/upgrades"

	icqtypes "github.com/neutron-org/neutron/v4/x/interchainqueries/types"
)

const (
	// UpgradeName defines the on-chain upgrades name.
	UpgradeName = "v0.6.7"
)

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: store.StoreUpgrades{
		Added:   []string{},
		Deleted: []string{icqtypes.ModuleName},
	},
}
