package types

import (
	"strconv"
	"strings"

	"cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	"github.com/Nolus-Protocol/nolus-core/app/params"
)

var baseAssetTicker = strings.ToUpper(params.HumanCoinUnit)

// OracleData is the struct we use to unmarshal the oracle's response for prices.
type OracleData struct {
	Prices []Price `json:"prices"`
}

// Price is inner the struct we use to unmarshal the oracle's response for prices.
type Price struct {
	Amount      PriceFeed `json:"amount"`
	AmountQuote PriceFeed `json:"amount_quote"`
}

type PriceFeed struct {
	Amount string `json:"amount"`
	Ticker string `json:"ticker"`
}

func (prices OracleData) CalculateValueInBaseAsset(ticker, stableTicker string, amount float64, requiredFees sdkmath.Int) (float64, float64, error) {
	var err error
	baseAssetAmountAsInt := 0
	baseAssetQuoteAmountAsInt := 0

	for _, price := range prices.Prices {
		if price.Amount.Ticker == baseAssetTicker {
			baseAssetAmountAsInt, err = strconv.Atoi(price.Amount.Amount)
			if err != nil {
				return 0, 0, err
			}

			baseAssetQuoteAmountAsInt, err = strconv.Atoi(price.AmountQuote.Amount)
			if err != nil {
				return 0, 0, err
			}
		}
	}

	if baseAssetAmountAsInt == 0 || baseAssetQuoteAmountAsInt == 0 {
		return 0, 0, errors.Wrapf(ErrNoPrices, "no prices found for nls")
	}

	if ticker == stableTicker {
		return calculateForStableToken(baseAssetAmountAsInt, baseAssetQuoteAmountAsInt, amount, requiredFees)
	} else {
		return calculateForVolatileToken(baseAssetAmountAsInt, baseAssetQuoteAmountAsInt, prices, ticker, amount, requiredFees)
	}
}

func calculateForStableToken(baseAssetAmountAsInt, baseAssetQuoteAmountAsInt int, amount float64, requiredFees sdkmath.Int) (float64, float64, error) {
	// {"amount":{"amount":"200000000","ticker":"NLS"},"amount_quote":{"amount":"12386383","ticker":"USDC"}}}

	// fee amount in stables * (price of 1 stable token in base asset)
	// 2 000 000uusdc        *  	16,146763749
	fullFeeAmountInBaseAsset := amount * (float64(baseAssetAmountAsInt) / float64(baseAssetQuoteAmountAsInt))

	// requiredFees is always in base asset, we calculate the value of the minimum required base asset in the paid denom
	// requiredFees * (price of 1 unit of base asset in usdc) * (price of 1 uusdc in unit of denom)
	// 2500unls        *   			0,027027027        *   			5,010046396
	requiredFeesInStable := float64(requiredFees.Int64()) * (float64(baseAssetQuoteAmountAsInt) / float64(baseAssetAmountAsInt))

	return fullFeeAmountInBaseAsset, requiredFeesInStable, nil
}

func calculateForVolatileToken(baseAssetAmountAsInt, baseAssetQuoteAmountAsInt int, prices OracleData, ticker string, amount float64, requiredFees sdkmath.Int) (float64, float64, error) {
	var err error
	denomAmountAsInt := 0
	denomQuoteAmountAsInt := 0

	for _, price := range prices.Prices {
		if price.Amount.Ticker == ticker {
			denomAmountAsInt, err = strconv.Atoi(price.Amount.Amount)
			if err != nil {
				return 0, 0, err
			}

			denomQuoteAmountAsInt, err = strconv.Atoi(price.AmountQuote.Amount)
			if err != nil {
				return 0, 0, err
			}
		}
	}

	if denomAmountAsInt == 0 || denomQuoteAmountAsInt == 0 {
		return 0, 0, errors.Wrapf(ErrNoPrices, "no prices found for %s", ticker)
	}

	// fee amount * (price of 1 unit of denom in usdc) * (price of 1 uusdc in unit of base asset)
	// 200uosmo        *   			0.6491066072588383usdc 			 	*  21.45123159028491unls
	fullFeeAmountInBaseAsset := amount * (float64(denomQuoteAmountAsInt) / float64(denomAmountAsInt)) * (float64(baseAssetAmountAsInt) / float64(baseAssetQuoteAmountAsInt))

	// {"amount":{"amount":"20000000","ticker":"OSMO"},"amount_quote":{"amount":"3991979","ticker":"USDC"}}
	// {"amount":{"amount":"1000000000000000000","ticker":"NLS"},"amount_quote":{"amount":"27027027027027027","ticker":"USDC"}}

	// requiredFees is always in base asset, we calculate the value of the minimum required base asset in the paid denom
	// requiredFees * (price of 1 base asset in uusdc) * (price of 1 uusdc in unit ofdenom)
	// 2500unls        *   			0,027027027        *   			5,010046396
	requiredFeesInPaidDenom := float64(requiredFees.Int64()) * (float64(baseAssetQuoteAmountAsInt) / float64(baseAssetAmountAsInt)) * (float64(denomAmountAsInt) / float64(denomQuoteAmountAsInt))

	return fullFeeAmountInBaseAsset, requiredFeesInPaidDenom, nil
}
