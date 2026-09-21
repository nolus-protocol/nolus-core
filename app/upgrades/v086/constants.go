package v086

import (
	store "cosmossdk.io/store/types"

	"github.com/Nolus-Protocol/nolus-core/app/upgrades"
	solanacarriertypes "github.com/Nolus-Protocol/nolus-core/solanacarrier"
)

const (
	// UpgradeName defines the on-chain upgrades name.
	UpgradeName = "v0.8.6"
)

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: store.StoreUpgrades{
		Added:   []string{solanacarriertypes.StoreKey},
		Deleted: []string{},
	},
}
