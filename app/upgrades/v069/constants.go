package v069

import (
	store "cosmossdk.io/store/types"
	"github.com/Nolus-Protocol/nolus-core/app/upgrades"
)

const (
	// UpgradeName defines the on-chain upgrades name.
	UpgradeName = "v0.6.9"
)

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: store.StoreUpgrades{
		Added: []string{},
	},
}
