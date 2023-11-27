package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Nolus-Protocol/nolus-core/x/tax/types"
)

func TestSuccessfulPriceCalculation(t *testing.T) {
	oracleData := types.OracleData{
		[]types.Price{
			{
				types.Amount{"1000", "OSMO"},
				types.AmountQuote{"4000", "USDC"},
			},
			{
				types.Amount{"2000", "NLS"},
				types.AmountQuote{"4000", "USDC"},
			},
		},
	}

	amountInBaseAsset, err := oracleData.CalculateValueInBaseAsset("OSMO", 100)
	require.NoError(t, err)
	require.Equal(t, float64(200), amountInBaseAsset)
}

func TestSuccessfulPriceCalculationRealisticPrices(t *testing.T) {
	oracleData := types.OracleData{
		[]types.Price{
			{
				types.Amount{"500000000000000000", "OSMO"},
				types.AmountQuote{"324553303629419159", "USDC"},
			},
			{
				types.Amount{"100000000", "NLS"},
				types.AmountQuote{"4661737", "USDC"},
			},
		},
	}

	amountInBaseAsset, err := oracleData.CalculateValueInBaseAsset("OSMO", 10000)
	require.NoError(t, err)
	require.Equal(t, float64(139241.3615909345), amountInBaseAsset)
}

func TestWrongPriceCalculationDueToMissingBaseAssetPrices(t *testing.T) {
	oracleData := types.OracleData{
		[]types.Price{
			{
				types.Amount{"1000", "OSMO"},
				types.AmountQuote{"4000", "USDC"},
			},
			{
				types.Amount{"2000", "missing"},
				types.AmountQuote{"4000", "USDC"},
			},
		},
	}

	_, err := oracleData.CalculateValueInBaseAsset("OSMO", 100)
	require.Error(t, err)
}

func TestWrongPriceCalculationDueToMalformedBasePrices(t *testing.T) {
	oracleData := types.OracleData{
		[]types.Price{
			{
				types.Amount{"1000", "OSMO"},
				types.AmountQuote{"4000", "USDC"},
			},
			{
				types.Amount{"20malformed00", "NLS"},
				types.AmountQuote{"4000", "USDC"},
			},
		},
	}

	_, err := oracleData.CalculateValueInBaseAsset("OSMO", 100)
	require.Error(t, err)
}

func TestWrongPriceCalculationDueToMalformedOsmoPrices(t *testing.T) {
	oracleData := types.OracleData{
		[]types.Price{
			{
				types.Amount{"10ss00", "OSMO"},
				types.AmountQuote{"4000", "USDC"},
			},
			{
				types.Amount{"2000", "missing"},
				types.AmountQuote{"4000", "USDC"},
			},
		},
	}

	_, err := oracleData.CalculateValueInBaseAsset("OSMO", 100)
	require.Error(t, err)
}

func TestWrongPriceCalculationDueToMissingPrices(t *testing.T) {
	oracleData := types.OracleData{}

	_, err := oracleData.CalculateValueInBaseAsset("OSMO", 100)
	require.Error(t, err)
}
