package v05

import (
	"github.com/Nolus-Protocol/nolus-core/app/upgrades"

	store "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	consensusparamstypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
)

const (
	// UpgradeName defines the on-chain upgrades name.
	UpgradeName = "v1.0.0"
)

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: store.StoreUpgrades{
		Added: []string{
			consensusparamstypes.ModuleName,
			crisistypes.ModuleName,
			authz.ModuleName,
		},
	},
}
