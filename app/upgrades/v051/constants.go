package v051

import (
	"github.com/Nolus-Protocol/nolus-core/app/upgrades"
	store "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
)

const (
	// UpgradeName defines the on-chain upgrades name.
	UpgradeName = "v0.5.1"
)

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: store.StoreUpgrades{
		Added: []string{
			feegrant.ModuleName,
		},
	},
}
