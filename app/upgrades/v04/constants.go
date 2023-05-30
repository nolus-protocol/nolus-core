package v04

import (
	"github.com/Nolus-Protocol/nolus-core/app/upgrades"
	vestingstypes "github.com/Nolus-Protocol/nolus-core/x/vestings/types"
	store "github.com/cosmos/cosmos-sdk/store/types"
)

const (
	// UpgradeName defines the on-chain upgrades name.
	UpgradeName = "v0.4.0"
)

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: store.StoreUpgrades{
		Added: []string{
			vestingstypes.ModuleName,
		},
	},
}
