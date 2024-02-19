package keeper

import (
	"encoding/json"
	"math"

	"cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	"github.com/Nolus-Protocol/nolus-core/x/tax/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// CustomTxFeeChecker reuses the default fee logic, but we will add the ability to pay fees in other denoms
// defined as a module parameter. The exact price will be calculated in base asset(defined
// in the min-gas-prices of the validators' config).
func (k Keeper) CustomTxFeeChecker(ctx sdk.Context, tx sdk.Tx) (sdk.Coins, int64, error) {
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return nil, 0, errors.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
	}

	feeCoins := feeTx.GetFee()
	gas := feeTx.GetGas()

	// Ensure that the provided fees meet a minimum threshold for the validator,
	// if this is a CheckTx. This is only for local mempool purposes, and thus
	// is only ran on check tx.
	if ctx.IsCheckTx() {
		minGasPrices := ctx.MinGasPrices()
		if !minGasPrices.IsZero() {
			requiredFees := make(sdk.Coins, len(minGasPrices))

			// Determine the required fees by multiplying each required minimum gas
			// price by the gas limit, where fee = ceil(minGasPrice * gasLimit).
			glDec := sdkmath.LegacyNewDec(int64(gas))
			for i, gp := range minGasPrices {
				fee := gp.Amount.Mul(glDec)
				requiredFees[i] = sdk.NewCoin(gp.Denom, fee.Ceil().RoundInt())
			}

			// Base denom is module param, should be "unls"
			baseDenom := k.GetParams(ctx).BaseDenom
			minimumFeeRequired := sdk.NewCoin(baseDenom, minGasPrices[0].Amount.Mul(glDec).Ceil().RoundInt())

			// if there are no fees paid in the base asset
			if ok, _ := feeCoins.Find(baseDenom); !ok {
				// Get Fee Param for select dex based on the feeCoins provided
				feeParam, err := getFeeParamBasedOnDenom(k.GetParams(ctx).FeeParams, feeCoins)
				if err != nil {
					return nil, 0, errors.Wrapf(sdkerrors.ErrInvalidRequest, err.Error())
				}

				// get the oracle address
				oracleAddress, err := sdk.AccAddressFromBech32(feeParam.OracleAddress)
				if err != nil {
					return nil, 0, errors.Wrapf(sdkerrors.ErrInvalidAddress, "failed to convert treasury, bech32 to AccAddress: %s: %s", feeParam.OracleAddress, err.Error())
				}

				// query the oracle for all available prices from this dex
				pricesBytes, err := k.wasmKeeper.QuerySmart(ctx, oracleAddress, []byte(`{"prices":{}}`))
				if err != nil {
					return nil, 0, errors.Wrapf(sdkerrors.ErrInvalidRequest, "failed to query oracle: %s", err.Error())
				}

				// unmarshal pricesBytes in an appropriate struct
				var prices types.OracleData
				err = json.Unmarshal(pricesBytes, &prices)
				if err != nil {
					return nil, 0, errors.Wrapf(sdkerrors.ErrJSONUnmarshal, "failed to unmarshal oracle data: %s", err.Error())
				}

				// go through every fee provided
				for _, fee := range feeCoins {
					denomTicker, err := isValidFeeDenom(fee.Denom, *feeParam)
					if err != nil {
						return nil, 0, err
					}

					if len(prices.Prices) == 0 {
						return nil, 0, errors.Wrapf(types.ErrNoPrices, "no prices found for oracle: %s", feeParam.OracleAddress)
					}

					// AmountQuote.Ticker should be the same for every price in the prices array fetched from an oracle so we just get the first price and use the stableTicker.
					// Each oracle could have different stableTicker.
					stableTicker := prices.Prices[0].AmountQuote.Ticker
					providedFeeAmountInBaseAsset, requiredFeesInPaidDenom, err := prices.CalculateValueInBaseAsset(denomTicker, stableTicker, fee.Amount.ToLegacyDec().MustFloat64(), requiredFees[0].Amount)
					if err != nil {
						return nil, 0, errors.Wrapf(sdkerrors.ErrInvalidRequest, "failed to calculate fee denom(%s) price in base asset: %s", fee.Denom, err.Error())
					}

					// if the fee calculated in nls is greater than the required fee in nls, then fee is valid
					if providedFeeAmountInBaseAsset > minimumFeeRequired.Amount.ToLegacyDec().MustFloat64() {
						priority := getTxPriority(feeCoins, int64(gas))
						return feeCoins, priority, nil
					} else {
						return nil, 0, errors.Wrapf(sdkerrors.ErrInsufficientFee, "insufficient fees; got: %f%s required: %f%s", fee.Amount.ToLegacyDec().MustFloat64(), fee.Denom, requiredFeesInPaidDenom, fee.Denom)
					}
				}
			}
			if !feeCoins.IsAnyGTE(requiredFees) {
				return nil, 0, errors.Wrapf(sdkerrors.ErrInsufficientFee, "insufficient fees; got: %s required: %s", feeCoins, requiredFees)
			}
		}
	}

	priority := getTxPriority(feeCoins, int64(gas))
	return feeCoins, priority, nil
}

func getFeeParamBasedOnDenom(feeParams []*types.FeeParam, feeCoins sdk.Coins) (*types.FeeParam, error) {
	var correctFeeParam *types.FeeParam
	// check if there is an accepted_denom in feeParams matching any of the feeCoins' denom
	for _, feeParam := range feeParams {
		correctFeeParam = findDenom(*feeParam, feeCoins)
		// if there is a match then we ensure this feeParam with correct oracle and profit
		// smart contrat addresses will be used. This is in case of multiple supported DEXes.
		if isFeeParamValid(correctFeeParam) {
			return correctFeeParam, nil
		}
	}

	return nil, errors.Wrapf(types.ErrInvalidFeeDenom, "no fee param found for denoms: %s", feeCoins)
}

func isFeeParamValid(feeParam *types.FeeParam) bool {
	if feeParam == nil {
		return false
	}
	if feeParam.OracleAddress == "" || feeParam.ProfitAddress == "" {
		return false
	}
	return true
}

func findDenom(feeParam types.FeeParam, feeCoins sdk.Coins) *types.FeeParam {
	for _, denomReadable := range feeParam.AcceptedDenoms {
		if ok, _ := feeCoins.Find(denomReadable.Denom); ok {
			// fees should be sorted(biggest to smallest), so on the first match, we conclude that this is the current dex's fee param
			// We need this check since denoms for the same token but from different dexes are different (because the channel differs)
			//
			// * Examples:
			// Osmo from dex1 would have denom ibc/72...
			// Osmo from dex2 would have denom ibc/2a...
			return &feeParam
		}
	}
	return nil
}

func isValidFeeDenom(denom string, feeParam types.FeeParam) (string, error) {
	ticker := ""
	for _, acceptedDenoms := range feeParam.AcceptedDenoms {
		if acceptedDenoms.Denom == denom {
			ticker = acceptedDenoms.Ticker
			return ticker, nil
		}
	}

	if ticker == "" {
		return "", errors.Wrapf(types.ErrInvalidFeeDenom, "denom(%s) is not allowed", denom)
	}

	return ticker, nil
}

// getTxPriority returns a naive tx priority based on the amount of the smallest denomination of the gas price
// provided in a transaction.
// NOTE: This implementation should be used with a great consideration as it opens potential attack vectors
// where txs with multiple coins could not be prioritize as expected.
func getTxPriority(fee sdk.Coins, gas int64) int64 {
	var priority int64
	for _, c := range fee {
		p := int64(math.MaxInt64)
		gasPrice := c.Amount.QuoRaw(gas)
		if gasPrice.IsInt64() {
			p = gasPrice.Int64()
		}
		if priority == 0 || p < priority {
			priority = p
		}
	}

	return priority
}
