package keeper

import (
	"fmt"

	"github.com/Nolus-Protocol/nolus-core/x/tax/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

var HUNDRED_DEC = sdk.NewDec(100)

// DeductTaxDecorator deducts tax by a given fee rate from the standard collected fee.
// The tax is sent to a treasury account
// Call next AnteHandler if tax successfully sent to treasury or no fee provided
// CONTRACT: Tx must implement FeeTx interface to use DeductTaxDecorator.
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

	// If fees are not specified we call the next AnteHandler
	txFees := feeTx.GetFee()
	if txFees.Empty() {
		return next(ctx, tx, simulate)
	}

	// Ensures the module treasury address has been set
	treasuryAddr, err := sdk.AccAddressFromBech32(dtd.tk.ContractAddress(ctx))
	if err != nil {
		return ctx, sdkerrors.Wrap(sdkerrors.ErrUnknownAddress, fmt.Sprintf("invalid treasury smart contract address: %s", err.Error()))
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

	baseDenom := dtd.tk.BaseDenom(ctx)
	if baseDenom != feeCoin.Denom {
		return ctx, sdkerrors.Wrap(types.ErrInvalidFeeDenom, txFees[0].Denom)
	}

	if err = deductTax(ctx, dtd.tk, dtd.bk, feeCoin, treasuryAddr); err != nil {
		return ctx, err
	}

	events := sdk.Events{sdk.NewEvent(sdk.EventTypeTx,
		sdk.NewAttribute(sdk.AttributeKeyFee, feeTx.GetFee().String()),
	)}

	ctx.EventManager().EmitEvents(events)

	return next(ctx, tx, simulate)
}

func deductTax(ctx sdk.Context, taxKeeper Keeper, bankKeeper types.BankKeeper, feeCoin sdk.Coin, treasuryAddr sdk.AccAddress) error {
	feeRate := sdk.NewDec(int64(taxKeeper.FeeRate(ctx)))
	// if feeRate is 0 - we won't deduct any tax
	if feeRate.IsZero() {
		return nil
	}

	tax := sdk.NewCoin(feeCoin.Denom, feeRate.MulInt(feeCoin.Amount).Quo(HUNDRED_DEC).TruncateInt())
	// There are cases where the tax calculation could result in a number between 0 and 1.
	// In those cases, the tax will be 0, since the lowest registered unit we have is 1unls
	// **Note - this case probably won't be reached in reality, because we enforce minimum fees(500 currently). So the feeAmount is always expected to be > 500.
	if tax.IsZero() {
		return nil
	}

	if !tax.IsValid() {
		return types.ErrInvalidTax
	}

	ctx.Logger().Info(fmt.Sprintf("Deducted tax: %s, final fee: %s", tax, feeCoin.Sub(tax)))

	// Send tax from fee collector to the treasury smart contract address
	err := bankKeeper.SendCoinsFromModuleToAccount(ctx, authtypes.FeeCollectorName, treasuryAddr, sdk.Coins{tax})
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInsufficientFunds, err.Error())
	}

	return nil
}
