package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"gitlab-nomo.credissimo.net/nomo/cosmzone/x/tax/types"
)

// DeductFeeDecorator deducts fees from the first signer of the tx
// If the first signer does not have the funds to pay for the fees, return with InsufficientFunds error
// Call next AnteHandler if fees successfully deducted
// CONTRACT: Tx must implement FeeTx interface to use DeductFeeDecorator
type DeductFeeDecorator struct {
	ak         types.AccountKeeper
	tk         Keeper
	bankKeeper types.BankKeeper
}

func NewDeductFeeDecorator(ak types.AccountKeeper, bk types.BankKeeper, tk Keeper) DeductFeeDecorator {
	return DeductFeeDecorator{
		ak:         ak,
		tk:         tk,
		bankKeeper: bk,
	}
}

func (dfd DeductFeeDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return ctx, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
	}

	if addr := dfd.ak.GetModuleAddress(authtypes.FeeCollectorName); addr == nil {
		panic(fmt.Sprintf("%s module account has not been set", authtypes.FeeCollectorName))
	}

	treasuryAddr, err := sdk.AccAddressFromBech32(dfd.tk.ContractAddress(ctx))
	if err != nil {
		return ctx, sdkerrors.Wrap(sdkerrors.ErrUnknownAddress, fmt.Sprintf("Invalid Treasury Smart Contract Address [ %s ]", err.Error()))
	}

	feePayer := feeTx.FeePayer()
	feeCoins := feeTx.GetFee()
	feeRate := sdk.NewDec(int64(dfd.tk.FeeRate(ctx)))

	deductFeesFrom := feePayer

	deductFeesFromAcc := dfd.ak.GetAccount(ctx, deductFeesFrom)
	if deductFeesFromAcc == nil {
		return ctx, sdkerrors.Wrapf(sdkerrors.ErrUnknownAddress, "fee payer address: %s does not exist", deductFeesFrom)
	}

	taxFees, afterTax, err := ApplyFee(feeRate, feeCoins)
	ctx.Logger().Info(fmt.Sprintf("DeductFees: tax: %s, fee: %s", taxFees, afterTax))
	if err != nil {
		return ctx, err
	}

	// deduct the fees
	if !feeTx.GetFee().IsZero() {
		err = DeductFees(ctx, dfd.bankKeeper, deductFeesFromAcc, treasuryAddr, taxFees, afterTax)
		if err != nil {
			return ctx, err
		}
	}

	//TODO set taxfee event
	events := sdk.Events{sdk.NewEvent(sdk.EventTypeTx,
		sdk.NewAttribute(sdk.AttributeKeyFee, feeTx.GetFee().String()),
	)}
	ctx.EventManager().EmitEvents(events)

	return next(ctx, tx, simulate)
}

// DeductFees deducts fees and tax from the given account.
func DeductFees(ctx sdk.Context, bankKeeper types.BankKeeper, acc authtypes.AccountI, treasuryAddr sdk.AccAddress, taxFees sdk.Coins, afterTax sdk.Coins) error {
	if !afterTax.IsValid() {
		return sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee, "invalid fee amount: %s", afterTax)
	}

	if !taxFees.IsValid() {
		return sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee, "invalid tax fee amount: %s", taxFees)
	}

	// Send taxFees to the treasury smart contract
	err := bankKeeper.SendCoins(ctx, acc.GetAddress(), treasuryAddr, taxFees)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInsufficientFunds, err.Error())
	}

	err = bankKeeper.SendCoinsFromAccountToModule(ctx, acc.GetAddress(), authtypes.FeeCollectorName, afterTax)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInsufficientFunds, err.Error())
	}

	return nil
}
