package keeper

import (
	"encoding/json"
	"math"
	"strconv"

	sdkmath "cosmossdk.io/math"
	"github.com/Nolus-Protocol/nolus-core/x/tax/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// TODO: test && check all calculations and make sure they are correct
type Price struct {
	Amount struct {
		Amount string `json:"amount"`
		Ticker string `json:"ticker"`
	} `json:"amount"`
	AmountQuote struct {
		Amount string `json:"amount"`
		Ticker string `json:"ticker"`
	} `json:"amount_quote"`
}

type OracleData struct {
	Prices []Price `json:"prices"`
}

// CustomTxFeeChecker reuses the default fee logic, but we will add the ability to pay fees in other denoms
// defined as a module parameter. The exact price will be calculated in usd representing the minimum value of base asset(defined
// in the min-gas-prices of the validators' config).
func (k Keeper) CustomTxFeeChecker(ctx sdk.Context, tx sdk.Tx) (sdk.Coins, int64, error) {
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return nil, 0, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
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

			// Base asset is always the first value defined in min-gas-prices config (should be unls)
			// TODO: what do we do if a malicious validator changes his base asset to something different than unls?
			baseFeeRequired := sdk.NewCoin(minGasPrices[0].Denom, minGasPrices[0].Amount.Mul(glDec).Ceil().RoundInt())

			// if there are no fees paid in the base asset
			if ok, _ := feeCoins.Find(baseFeeRequired.Denom); !ok {

				// Get FeeParams from tax keeper
				feeParams := k.GetParams(ctx).FeeParams

				var correctFeeParam types.FeeParam
			outerLoop:
				// check if there is a accepted_denom in feeParams matching any of the paid feeCoins
				for _, feeParam := range feeParams {
					for _, denom := range feeParam.AcceptedDenoms {
						if ok, _ := feeCoins.Find(denom); ok {
							// fees should be sorted, so on the first match, we conclude that this is the current dex's fee param
							// We need this check since denoms for the same token but from different dexes are different (because the channel differs)
							//
							// * Examples:
							// Osmo from dex1 would have denom ibc/72...
							// Osmo from dex2 would have denom ibc/2a...
							correctFeeParam = *feeParam
							break outerLoop
						}
					}
				}

				// get the oracle address
				oracleAddress, err := sdk.AccAddressFromBech32(correctFeeParam.OracleAddress)
				if err != nil {
					return nil, 0, errors.Wrapf(sdkerrors.ErrInvalidAddress, "failed to convert treasury, bech32 to AccAddress: %s: %s", correctFeeParam.OracleAddress, err.Error())
				}

				// query the oracle for all available prices
				pricesBytes, err := k.wasmKeeper.QuerySmart(ctx, oracleAddress, []byte(`{"prices":{}}`))
				if err != nil {
					return nil, 0, errors.Wrapf(sdkerrors.ErrInvalidRequest, "failed to query oracle: %s", err.Error())
				}

				// unmarshal pricesBytes in an appropriate struct
				var prices OracleData
				err = json.Unmarshal(pricesBytes, &prices)
				if err != nil {
					return nil, 0, errors.Wrapf(sdkerrors.ErrJSONUnmarshal, "failed to unmarshal oracle data: %s", err.Error())
				}

				// Calculate required fee in usdc
				requiredFeeAmountInUsdc, err := calculateuDenomInUSDC(baseFeeRequired.Denom, baseFeeRequired.Amount.ToLegacyDec().MustFloat64(), prices)
				if err != nil {
					return nil, 0, errors.Wrapf(sdkerrors.ErrInvalidRequest, "failed to calculate base denom(%s) price in usdc: %s", baseFeeRequired.Denom, err.Error())
				}

				// go through every fee provided
				for _, fee := range feeCoins {
					currentFeeAmountInUsdc, err := calculateuDenomInUSDC(fee.Denom, fee.Amount.ToLegacyDec().MustFloat64(), prices)
					if err != nil {
						return nil, 0, errors.Wrapf(sdkerrors.ErrInvalidRequest, "failed to calculate fee denom(%s) price in usdc: %s", fee.Denom, err.Error())
					}

					// if the fee calculated in usdc is greater than the required fee in usdc, then fee is valid
					if currentFeeAmountInUsdc > requiredFeeAmountInUsdc {
						priority := getTxPriority(feeCoins, int64(gas))
						return feeCoins, priority, nil
					}
				}
			}
			if !feeCoins.IsAnyGTE(requiredFees) {
				return nil, 0, sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee, "insufficient fees; got: %s required: %s", feeCoins, requiredFees)
			}
		}
	}

	priority := getTxPriority(feeCoins, int64(gas))
	return feeCoins, priority, nil
}

func calculateuDenomInUSDC(denom string, amount float64, prices OracleData) (float64, error) {
	// divisonZeroes := 1_000_000

	for _, price := range prices.Prices {
		if price.Amount.Ticker == denom {
			amountAsInt, err := strconv.Atoi(price.Amount.Amount)
			if err != nil {
				return 0, err
			}

			quoteAmountAsInt, err := strconv.Atoi(price.AmountQuote.Amount)
			if err != nil {
				return 0, err
			}

			// For these denoms, the price of the oracle can be calculated for the smallest unit of the token
			// // TODO:  || denom == "WBTC"
			// if denom == "WETH" || denom == "EVMOS" || denom == "INJ" {
			// 	fullFeeAmountInUsdc := amount * (float64(quoteAmountAsInt) / float64(amountAsInt))
			// 	return fullFeeAmountInUsdc, nil
			// }

			// get the price of 1 token in usdc
			// TODO: check float max zeroes ?
			fullFeeAmountInUsdc := amount * (float64(quoteAmountAsInt) / float64(amountAsInt))

			// // Get the price of 1 uDenom in usdc. We divide based on what asset we are working with.
			// uTokenPriceInUSDC := TokenInUSDC / float64(divisonZeroes)
			return fullFeeAmountInUsdc, nil
		}
	}

	return 0, errors.Wrapf(types.ErrInvalidFeeDenom, "unsupported denom for paying fees: %s", denom)
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
