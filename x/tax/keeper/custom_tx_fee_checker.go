package keeper

import (
	"math"

	"cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"

	types "github.com/Nolus-Protocol/nolus-core/x/tax/typesv2"
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

			// if there are no fees provided
			if feeCoins.Len() == 0 {
				return nil, 0, errors.Wrapf(sdkerrors.ErrInsufficientFee, "insufficient fees; got: %s required: %s", feeCoins, requiredFees)
			}

			params, err := k.GetParams(ctx)
			if err != nil {
				return nil, 0, errors.Wrap(sdkerrors.ErrNotFound, err.Error())
			}

			// if there are no fees paid in the base asset
			if ok, _ := feeCoins.Find(params.BaseDenom); !ok {
				// Get Fee Param for select dex based on the feeCoins provided
				feeParam, err := getFeeParamBasedOnDenom(params.DexFeeParams, feeCoins)
				if err != nil {
					return nil, 0, errors.Wrapf(sdkerrors.ErrInvalidRequest, err.Error()) //nolint:govet
				}

				// go through every fee provided
				for _, fee := range feeCoins {
					minGasPrice, err := denomMinPrice(fee.Denom, *feeParam)
					if err != nil {
						return nil, 0, err
					}

					// price by the gas limit, where fee = ceil(minGasPrice * gasLimit).
					gasLimitDec := sdkmath.LegacyNewDec(int64(gas))
					minimumFeeRequiredInPaidDenom := gasLimitDec.Mul(minGasPrice).RoundInt()
					// if the fee provided is greater than the minimum fee required in the paid denom, then it is accepted
					if fee.Amount.GT(minimumFeeRequiredInPaidDenom) {
						priority := getTxPriority(feeCoins, int64(gas))
						return feeCoins, priority, nil
					} else {
						return nil, 0, errors.Wrapf(sdkerrors.ErrInsufficientFee, "insufficient fees; got: %d%s required: %d%s", fee.Amount.ToLegacyDec().RoundInt().Int64(), fee.Denom, minimumFeeRequiredInPaidDenom.Int64(), fee.Denom)
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

func getFeeParamBasedOnDenom(feeParams []*types.DexFeeParams, feeCoins sdk.Coins) (*types.DexFeeParams, error) {
	var correctFeeParam *types.DexFeeParams
	// check if there is an accepted_denom in feeParams matching any of the paid fees' denom
	for _, feeParam := range feeParams {
		correctFeeParam = findDenom(*feeParam, feeCoins)
		// if there is a match then we ensure this feeParam with correct profit
		// smart contrat addresses will be used. This is in case of multiple supported DEXes.
		if correctFeeParam != nil {
			return correctFeeParam, nil
		}
	}

	return nil, errors.Wrapf(types.ErrInvalidFeeDenom, "no fee param found for denoms: %s", feeCoins)
}

func validateFeeParam(profitAddress string, pair *types.DenomPrice) bool {
	if profitAddress == "" {
		return false
	}

	minPrice, err := sdkmath.LegacyNewDecFromStr(pair.MinPrice)
	if err != nil {
		return false
	}

	if pair.Denom == "" || minPrice.IsZero() {
		return false
	}

	return true
}

func findDenom(feeParams types.DexFeeParams, feeCoins sdk.Coins) *types.DexFeeParams {
	for _, denom := range feeParams.AcceptedDenomsMinPrices {
		if ok, _ := feeCoins.Find(denom.Denom); ok {
			// fees should be sorted(biggest to smallest), so on the first match, we conclude that this is the current dex's fee param
			// We need this check since denoms for the same token but from different dexes are different (because the channel differs)
			//
			// * Examples:
			// Osmo from dex1 would have denom ibc/72...
			// Osmo from dex2 would have denom ibc/2a...
			if validateFeeParam(feeParams.ProfitAddress, denom) {
				return &feeParams
			}
		}
	}
	return nil
}

func denomMinPrice(denom string, feeParam types.DexFeeParams) (sdkmath.LegacyDec, error) {
	for _, denomMinPrice := range feeParam.AcceptedDenomsMinPrices {
		if denomMinPrice.Denom == denom {
			minPrice, err := sdkmath.LegacyNewDecFromStr(denomMinPrice.MinPrice)
			if err != nil {
				return sdkmath.LegacyDec{}, errors.Wrapf(types.ErrInvalidFeeDenom, "minPrice(%s) is not a valid decimal", denomMinPrice.MinPrice)
			}
			return minPrice, nil
		}
	}

	return sdkmath.LegacyDec{}, errors.Wrapf(types.ErrInvalidFeeDenom, "denom(%s) is not allowed", denom)
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
