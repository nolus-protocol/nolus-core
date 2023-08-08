package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

// AccountKeeper defines the expected account keeper used for simulations (noalias).
type AccountKeeper interface {
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) types.AccountI
	// Methods imported from account should be defined here
	GetModuleAddress(moduleName string) sdk.AccAddress
}

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	IsSendEnabledCoins(ctx sdk.Context, coins ...sdk.Coin) error
	SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	// Methods imported from bank should be defined here
}
