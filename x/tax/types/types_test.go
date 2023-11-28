package types_test

import (
	"testing"

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
		name       string
		oracleData types.OracleData
		expResult  float64
		expError   bool
	}{
		{
			"successful price calculation",
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
			false,
		},
		{
			"successful price calculation realistic prices",
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
			false,
		},
		{
			"wrong price calculation due to missing base asset prices",
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
			true,
		},
		{
			"wrong price calculation due to malformed base prices",
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
			true,
		},
		{
			"wrong price calculation due to missing prices ",
			types.OracleData{},
			float64(0),
			true,
		},
	}
	for _, tc := range testCases {
		s.Run(tc.name, func() {
			amountInBaseAsset, err := tc.oracleData.CalculateValueInBaseAsset("OSMO", 100)
			if tc.expError {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
			}
			s.Require().Equal(tc.expResult, amountInBaseAsset)
		})
	}
}

func TestTypesTestSuite(t *testing.T) {
	suite.Run(t, new(TypesTestSuite))
}

func TestWrongPriceCalculationDueToMissingPrices(t *testing.T) {
	oracleData := types.OracleData{}

	_, err := oracleData.CalculateValueInBaseAsset("OSMO", 100)
	require.Error(t, err)
}
