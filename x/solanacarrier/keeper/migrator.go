package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	v2 "github.com/Nolus-Protocol/nolus-core/x/solanacarrier/migrations/v2"
)

type Migrator struct {
	keeper Keeper
}

func NewMigrator(k Keeper) Migrator {
	return Migrator{keeper: k}
}

// Migrate1to2 migrates the x/solanacarrier module state from consensus version
// 1 to version 2, adding the program allowlist to the stored params.
func (m Migrator) Migrate1to2(ctx sdk.Context) error {
	return v2.Migrate(ctx, m.keeper)
}
