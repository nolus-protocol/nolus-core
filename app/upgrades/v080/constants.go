package v072

import (
	store "cosmossdk.io/store/types"
	"github.com/Nolus-Protocol/nolus-core/app/upgrades"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
)

const (
	// UpgradeName defines the on-chain upgrades name.
	UpgradeName = "v0.8.0"
)

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: store.StoreUpgrades{
		Added:   []string{},
		Deleted: []string{crisistypes.ModuleName, capabilitytypes.ModuleName},
	},
}
