package types

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// AccountKeeper defines the expected account keeper used for simulations (noalias).
type AccountKeeper interface {
	GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
	// Methods imported from account should be defined here
	GetModuleAddress(moduleName string) sdk.AccAddress
}

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	IsSendEnabledCoins(ctx context.Context, coins ...sdk.Coin) error
	SpendableCoins(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	SendCoins(ctx context.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	// Methods imported from bank should be defined here
}

// WasmKeeper defines the expected interface needed to retrieve price from oracle instance of smart contract.
type WasmKeeper interface {
	QuerySmart(ctx context.Context, contractAddr sdk.AccAddress, req []byte) ([]byte, error)
}
