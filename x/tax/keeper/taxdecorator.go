package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"gitlab-nomo.credissimo.net/nomo/cosmzone/x/tax/types"
)

var HUNDRED_DEC = sdk.NewDec(100)

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

func (dfd DeductTaxDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
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

	feeCoins := feeTx.GetFee()
	feeRate := sdk.NewDec(int64(dfd.tk.FeeRate(ctx)))

	taxFees, remainingFees, err := ApplyTax(feeRate, feeCoins)
	if err != nil {
		return ctx, err
	}
	ctx.Logger().Info(fmt.Sprintf("DeductTaxes: tax: %s, fee: %s", taxFees, remainingFees))

	// Send taxFees from fee collector to the treasury smart contract
	err = dfd.bk.SendCoins(ctx, dfd.ak.GetModuleAddress(authtypes.FeeCollectorName), treasuryAddr, taxFees)
	if err != nil {
		return ctx, sdkerrors.Wrapf(sdkerrors.ErrInsufficientFunds, err.Error())
	}

	events := sdk.Events{sdk.NewEvent(sdk.EventTypeTx,
		sdk.NewAttribute(sdk.AttributeKeyFee, feeTx.GetFee().String()),
	)}
	ctx.EventManager().EmitEvents(events)

	return next(ctx, tx, simulate)
}

func ApplyTax(feeRate sdk.Dec, feeCoins sdk.Coins) (sdk.Coins, sdk.Coins, error) {
	taxFees := sdk.Coins{}

	if feeRate.IsZero() || feeCoins.Empty() {
		return taxFees, feeCoins, nil
	}

	// we will deduct the fee from every denomination send
	for _, fee := range feeCoins {
		taxFee := sdk.NewCoin(fee.Denom, feeRate.MulInt(fee.Amount).Quo(HUNDRED_DEC).TruncateInt())
		taxFees = taxFees.Add(taxFee)
	}

	remainingFees, neg := feeCoins.SafeSub(taxFees)
	if neg {
		return nil, nil, sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee, "ApplyFee: insufficient fees; got: %s required: %s", feeCoins, taxFees)
	}

	return taxFees, remainingFees, nil
}
