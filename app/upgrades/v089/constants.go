package v089

import (
	store "cosmossdk.io/store/types"

	"github.com/Nolus-Protocol/nolus-core/app/upgrades"
)

const UpgradeName = "v0.8.9"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: store.StoreUpgrades{
		Added:   []string{},
		Deleted: []string{},
	},
}
