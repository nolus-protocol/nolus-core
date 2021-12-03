package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	autht "github.com/cosmos/cosmos-sdk/x/auth/types"
	types2 "gitlab-nomo.credissimo.net/nomo/cosmzone/x/suspend/types"
)

// AccountKeeper defines the contract needed for AccountKeeper related APIs.
// Interface provides support to use non-sdk AccountKeeper for AnteHandler's decorators.
type AccountKeeper interface {
	GetParams(ctx sdk.Context) (params autht.Params)
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) autht.AccountI
	SetAccount(ctx sdk.Context, acc autht.AccountI)
	GetModuleAddress(moduleName string) sdk.AccAddress
}

type SuspendKeeper interface {
	IsNodeSuspend(ctx sdk.Context) (suspend types2.MsgChangeSuspend)
}
