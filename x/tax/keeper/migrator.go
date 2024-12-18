package keeper

import (
	"github.com/Nolus-Protocol/nolus-core/x/tax/exported"
	v4 "github.com/Nolus-Protocol/nolus-core/x/tax/migrations/v4"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Migrator is a struct for handling in-place state migrations.
type Migrator struct {
	keeper         Keeper
	legacySubspace exported.Subspace
}

func NewMigrator(k Keeper, ss exported.Subspace) Migrator {
	return Migrator{
		keeper:         k,
		legacySubspace: ss,
	}
}

// Migrate3to4 migrates the x/tax module state from the consensus version 3 to
// version 4. Specifically, it takes the parameters that are currently stored
// and managed by the x/params modules and stores them directly into the x/tax
// module state.
func (m Migrator) Migrate3to4(ctx sdk.Context) error {
	return v4.Migrate(ctx, m.keeper.storeService.OpenKVStore(ctx), m.keeper.cdc)
}
