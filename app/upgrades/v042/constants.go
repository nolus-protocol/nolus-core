package v042

import (
	store "cosmossdk.io/store/types"
	"github.com/Nolus-Protocol/nolus-core/app/upgrades"
	"github.com/cosmos/cosmos-sdk/x/authz"
)

const (
	// UpgradeName defines the on-chain upgrades name.
	UpgradeName = "v0.4.2"
)

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: store.StoreUpgrades{
		Added: []string{
			authz.ModuleName,
		},
	},
}
