package keeper

import (
	"github.com/Nolus-Protocol/nolus-core/x/interchaintxs/types"
)

var _ types.QueryServer = Keeper{}
