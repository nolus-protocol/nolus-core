package v065

import (
	"github.com/Nolus-Protocol/nolus-core/app/upgrades"
	store "github.com/cosmos/cosmos-sdk/store/v2/types"
)

const (
	// UpgradeName defines the on-chain upgrades name.
	UpgradeName = "v0.6.5"
)

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: store.StoreUpgrades{
		Added: []string{},
	},
}
