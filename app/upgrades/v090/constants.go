package v090

import (
	store "cosmossdk.io/store/types"

	"github.com/Nolus-Protocol/nolus-core/app/upgrades"
)

const UpgradeName = "v0.9.0"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: store.StoreUpgrades{
		Added:   []string{},
		Deleted: []string{},
	},
}
