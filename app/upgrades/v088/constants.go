package v088

import (
	store "cosmossdk.io/store/types"

	"github.com/Nolus-Protocol/nolus-core/app/upgrades"
)

const UpgradeName = "v0.8.8"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: store.StoreUpgrades{
		Added:   []string{},
		Deleted: []string{},
	},
}
