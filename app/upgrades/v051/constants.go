package v051

import (
	"github.com/Nolus-Protocol/nolus-core/app/upgrades"
	store "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
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
