package keeper

import (
	types "github.com/Nolus-Protocol/nolus-core/x/tax/typesv2"
)

var _ types.QueryServer = Keeper{}
