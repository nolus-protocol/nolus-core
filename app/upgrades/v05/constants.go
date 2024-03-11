package v05

import (
	"github.com/Nolus-Protocol/nolus-core/app/upgrades"

	store "cosmossdk.io/store/types"
	consensusparamstypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
)

const (
	// UpgradeName defines the on-chain upgrades name.
	UpgradeName = "v0.5.0"
)

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: store.StoreUpgrades{
		Added: []string{
			consensusparamstypes.ModuleName,
			crisistypes.ModuleName,
		},
	},
}
