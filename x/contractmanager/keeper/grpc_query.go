package keeper

import (
	"github.com/Nolus-Protocol/nolus-core/x/contractmanager/types"
)

var _ types.QueryServer = Keeper{}
