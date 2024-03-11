package v051

import (
	store "cosmossdk.io/store/types"
	"cosmossdk.io/x/feegrant"
	"github.com/Nolus-Protocol/nolus-core/app/upgrades"
	icahosttypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/host/types"
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
			icahosttypes.StoreKey,
		},
	},
}
