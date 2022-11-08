package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"gitlab-nomo.credissimo.net/nomo/nolus-core/x/tax/types"
)

var HUNDRED_DEC = sdk.NewDec(100)

// DeductTaxDecorator applies tax by a given fee rate on top of the standard Tx fee.
// The additional tax is sent it to a treasury account
// Call next AnteHandler if tax successfully sent to treasury
// CONTRACT: Tx must implement FeeTx interface to use DeductTaxDecorator
type DeductTaxDecorator struct {
	ak types.AccountKeeper
	tk Keeper
	bk types.BankKeeper
}

func NewDeductTaxDecorator(ak types.AccountKeeper, bk types.BankKeeper, tk Keeper) DeductTaxDecorator {
	return DeductTaxDecorator{
		ak: ak,
		tk: tk,
		bk: bk,
	}
}

func (dtd DeductTaxDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return ctx, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
	}

	// Ensures the module treasury address has been set
	treasuryAddr, err := sdk.AccAddressFromBech32(dtd.tk.ContractAddress(ctx))
	if err != nil {
		return ctx, sdkerrors.Wrap(sdkerrors.ErrUnknownAddress, fmt.Sprintf("invalid treasury smart contract address: %s", err.Error()))
	}

	txFees := feeTx.GetFee()
	if txFees.Empty() {
		return ctx, types.ErrFeesNotSet
	}

	// Ensure not more then one denom for paying tx costs
	if len(txFees) > 1 {
		return ctx, types.ErrTooManyFeeCoins
	}

	feeCoin := txFees[0]
	if feeCoin.IsNil() || feeCoin.Amount.IsZero() {
		return ctx, types.ErrAmountNilOrZero
	}

	if err = feeCoin.Validate(); err != nil {
		return ctx, err
	}

	allowedDenoms := dtd.tk.FeeDenoms(ctx)
	if !isAllowed(allowedDenoms, feeCoin.Denom) {
		return ctx, sdkerrors.Wrap(types.ErrInvalidFeeDenom, txFees[0].Denom)
	}

	feeRate := sdk.NewDec(int64(dtd.tk.FeeRate(ctx)))
	if feeRate.IsZero() {
		return ctx, types.ErrAmountNilOrZero
	}

	// calculate the tax
	var taxFees sdk.Coins
	tax := sdk.NewCoin(feeCoin.Denom, feeRate.MulInt(feeCoin.Amount).Quo(HUNDRED_DEC).TruncateInt())
	taxFees = taxFees.Add(tax)

	remainingFees := feeCoin.Sub(tax)
	if remainingFees.IsNegative() {
		return ctx, sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee, "got: %s required: %s", feeCoin, tax)
	}

	ctx.Logger().Info(fmt.Sprintf("DeductTaxes: tax: %s, final fee: %s", tax, remainingFees))

	fmt.Printf("$$$$$$$$$$$$$$tax is: %v, valid: %v, zero: %v \n", tax, tax.IsValid(), tax.IsZero())
	if tax.IsValid() && !tax.IsZero() {
		// Send tax from fee collector to the treasury smart contract
		err = dtd.bk.SendCoinsFromModuleToAccount(ctx, authtypes.FeeCollectorName, treasuryAddr, taxFees)
		if err != nil {
			return ctx, sdkerrors.Wrapf(sdkerrors.ErrInsufficientFunds, err.Error())
		}
	}

	events := sdk.Events{sdk.NewEvent(sdk.EventTypeTx,
		sdk.NewAttribute(sdk.AttributeKeyFee, feeTx.GetFee().String()),
	)}

	ctx.EventManager().EmitEvents(events)

	return next(ctx, tx, simulate)
}

func isAllowed(denoms []string, denom string) bool {
	for _, d := range denoms {
		if d == denom {
			return true
		}
	}
	return false
}
