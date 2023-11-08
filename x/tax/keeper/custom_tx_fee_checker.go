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

const baseAssetTicker = "NLS"

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
			minimumFeeRequired := sdk.NewCoin(minGasPrices[0].Denom, minGasPrices[0].Amount.Mul(glDec).Ceil().RoundInt())

			// if there are no fees paid in the base asset
			if ok, _ := feeCoins.Find(minimumFeeRequired.Denom); !ok {

				// Get FeeParams from tax keeper
				feeParams := k.GetParams(ctx).FeeParams

				var err error
				var correctFeeParam *types.FeeParam
				// check if there is an accepted_denom in feeParams matching any of the paid feeCoins
				for _, feeParam := range feeParams {
					correctFeeParam = findDenom(*feeParam, feeCoins)
					// if there is a match then we ensure the feeParam with correct oracle and profit
					// smart contrat addresses will be used. This is in case of multiple supported DEXes.
					if correctFeeParam != nil {
						break
					}
				}

				if !isFeeParamValid(correctFeeParam) {
					return nil, 0, errors.Wrapf(types.ErrInvalidFeeParam, "oracle address or profit address is not set")
				}

				// get the oracle address
				oracleAddress, err := sdk.AccAddressFromBech32(correctFeeParam.OracleAddress)
				if err != nil {
					return nil, 0, errors.Wrapf(sdkerrors.ErrInvalidAddress, "failed to convert treasury, bech32 to AccAddress: %s: %s", correctFeeParam.OracleAddress, err.Error())
				}

				// query the oracle for all available prices from this dex
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
				// go through every fee provided
				for _, fee := range feeCoins {
					currentFeeAmountInNLS, err := calculateValueInBaseAsset(fee.Denom, fee.Amount.ToLegacyDec().MustFloat64(), prices, *correctFeeParam)
					if err != nil {
						return nil, 0, errors.Wrapf(sdkerrors.ErrInvalidRequest, "failed to calculate fee denom(%s) price in base asset: %s", fee.Denom, err.Error())
					}

					// if the fee calculated in nls is greater than the required fee in nls, then fee is valid
					if currentFeeAmountInNLS > minimumFeeRequired.Amount.ToLegacyDec().MustFloat64() {
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
			// fees should be sorted, so on the first match, we conclude that this is the current dex's fee param
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

func calculateValueInBaseAsset(denom string, amount float64, prices OracleData, feeParam types.FeeParam) (float64, error) {
	ticker := ""
	for _, acceptedDenoms := range feeParam.AcceptedDenoms {
		if acceptedDenoms.Denom == denom {
			ticker = acceptedDenoms.Ticker
			break
		}
	}

	if ticker == "" {
		return 0, errors.Wrapf(types.ErrInvalidFeeDenom, "denom(%s) is not allowed", denom)
	}

	var err error
	denomAmountAsInt := 0
	denomQuoteAmountAsInt := 0
	baseAssetAmountAsInt := 0
	baseAssetQuoteAmountAsInt := 0
	for _, price := range prices.Prices {
		if price.Amount.Ticker == baseAssetTicker {
			baseAssetAmountAsInt, err = strconv.Atoi(price.Amount.Amount)
			if err != nil {
				return 0, err
			}

			baseAssetQuoteAmountAsInt, err = strconv.Atoi(price.AmountQuote.Amount)
			if err != nil {
				return 0, err
			}
		}

		if price.Amount.Ticker == ticker {
			denomAmountAsInt, err = strconv.Atoi(price.Amount.Amount)
			if err != nil {
				return 0, err
			}

			denomQuoteAmountAsInt, err = strconv.Atoi(price.AmountQuote.Amount)
			if err != nil {
				return 0, err
			}
		}
	}

	if denomAmountAsInt == 0 || denomQuoteAmountAsInt == 0 || baseAssetAmountAsInt == 0 || baseAssetQuoteAmountAsInt == 0 {
		return 0, errors.Wrapf(types.ErrInvalidFeeDenom, "no prices found for nls or %s", denom)
	}

	// TODO: check float max decimals ?
	fullFeeAmountInBaseAsset := amount * (float64(denomQuoteAmountAsInt) / float64(denomAmountAsInt)) * (float64(baseAssetAmountAsInt) / float64(baseAssetQuoteAmountAsInt))

	return fullFeeAmountInBaseAsset, nil
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
