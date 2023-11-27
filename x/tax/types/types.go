package types

import (
	"strconv"
	"strings"

	"cosmossdk.io/errors"
	"github.com/Nolus-Protocol/nolus-core/app/params"
)

var baseAssetTicker = strings.ToUpper(params.HumanCoinUnit)

// OracleData is the struct we use to unmarshal the oracle's response for prices.
type OracleData struct {
	Prices []Price `json:"prices"`
}

// Price is inner the struct we use to unmarshal the oracle's response for prices.
type Price struct {
	Amount      `json:"amount"`
	AmountQuote `json:"amount_quote"`
}

type Amount struct {
	Amount string `json:"amount"`
	Ticker string `json:"ticker"`
}

type AmountQuote struct {
	Amount string `json:"amount"`
	Ticker string `json:"ticker"`
}

func (prices OracleData) CalculateValueInBaseAsset(ticker string, amount float64) (float64, error) {
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
		return 0, errors.Wrapf(ErrInvalidFeeDenom, "no prices found for nls or %s", ticker)
	}

	// fee amount * (price of 1 unit of denom in usdc) * (price of 1 uusdc in smallest unit of base asset)
	// 200uosmo        *   			0.6491066072588383usdc 			 	*  21.45123159028491unls
	fullFeeAmountInBaseAsset := amount * (float64(denomQuoteAmountAsInt) / float64(denomAmountAsInt)) * (float64(baseAssetAmountAsInt) / float64(baseAssetQuoteAmountAsInt))

	return fullFeeAmountInBaseAsset, nil
}
