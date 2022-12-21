package keeper

import (
	"github.com/Nolus-Protocol/nolus-core/x/tax/types"
)

var _ types.QueryServer = Keeper{}
