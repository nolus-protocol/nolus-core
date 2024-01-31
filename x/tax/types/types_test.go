package types_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/Nolus-Protocol/nolus-core/x/tax/types"
)

type TypesTestSuite struct {
	suite.Suite
}

func (s *TypesTestSuite) TearDownSuite() {
	s.T().Log("tearing down types test suite")
}

func (s *TypesTestSuite) TestSuccessfulPriceCalculation() {
	testCases := []struct {
		name           string
		requiredFee    int64
		ticker         string
		stableTicker   string
		oracleData     types.OracleData
		expAmount      float64
		expRequiredFee float64
		expError       bool
	}{
		{
			"successful price calculation",
			100,
			"OSMO",
			"",
			types.OracleData{
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
			},
			float64(200),
			float64(50),
			false,
		},
		{
			"successful price calculation realistic prices",
			100,
			"OSMO",
			"",
			types.OracleData{
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
			},
			float64(1392.4136159093453),
			float64(7.181774068956722),
			false,
		},
		{
			"successful price calculation, fees in stable, realistic prices",
			100,
			"USDC",
			"USDC",
			types.OracleData{
				[]types.Price{
					{
						types.Amount{"100000000", "NLS"},
						types.AmountQuote{"4661737", "USDC"},
					},
				},
			},
			float64(2145.123159028491),
			float64(4.661737),
			false,
		},
		{
			"not enough fee amount, realistic prices",
			1000000,
			"OSMO",
			"",
			types.OracleData{
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
			},
			float64(1392.4136159093453),
			float64(71817.74068956722),
			false,
		},
		{
			"wrong price calculation due to missing base asset prices",
			100,
			"OSMO",
			"",
			types.OracleData{
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
			},
			float64(0),
			float64(0),
			true,
		},
		{
			"wrong price calculation due to malformed base prices",
			100,
			"OSMO",
			"",
			types.OracleData{
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
			},
			float64(0),
			float64(0),
			true,
		},
		{
			"wrong price calculation due to missing prices ",
			100,
			"OSMO",
			"",
			types.OracleData{},
			float64(0),
			float64(0),
			true,
		},
	}
	for _, tc := range testCases {
		s.Run(tc.name, func() {
			amountInBaseAsset, requiredFeesInDenom, err := tc.oracleData.CalculateValueInBaseAsset(tc.ticker, tc.stableTicker, 100, sdkmath.NewInt(tc.requiredFee))
			if tc.expError {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
			}
			s.Require().Equal(tc.expAmount, amountInBaseAsset)
			s.Require().Equal(tc.expRequiredFee, requiredFeesInDenom)
		})
	}
}

func TestTypesTestSuite(t *testing.T) {
	suite.Run(t, new(TypesTestSuite))
}

func TestWrongPriceCalculationDueToMissingPrices(t *testing.T) {
	oracleData := types.OracleData{}

	_, _, err := oracleData.CalculateValueInBaseAsset("OSMO", "", 100, sdkmath.NewInt(100))
	require.Error(t, err)
}
