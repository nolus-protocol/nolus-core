package keeper_test

import (
	"testing"

	"cosmossdk.io/math"
	keepertest "github.com/Nolus-Protocol/nolus-core/testutil/keeper"
	types "github.com/Nolus-Protocol/nolus-core/x/tax/typesv2"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

var (
	osmoDenom        = "ibc/C4CFF46FD6DE35CA4CF4CE031E643C8FDC9BA4B99AE598E9B0ED98FE3A2319F9y"
	osmoAxlUSDCDenom = "ibc/5DE4FCAF68AE40F81F738C857C0D95F7C1BC47B00FA1026E85C1DD92524D4A11"
	feeAmount        = int64(1_000_000_000)
)

// Successfully pay fees in ibc/C4C... which represents OSMO. Minimum gas prices set to unls.
func TestCustomTxFeeCheckerSuccessfulInOsmo(t *testing.T) {
	taxKeeper, ctx := keepertest.TaxKeeper(t, true, sdk.DecCoins{sdk.NewDecCoin("unls", math.NewInt(1))}, types.DefaultParams())
	// create a new CustomTxFeeChecker
	feeTx := keepertest.MockFeeTx{
		Msgs: []sdk.Msg{},
		Gas:  100000,
		Fee:  sdk.Coins{sdk.NewInt64Coin(osmoDenom, feeAmount)},
	}

	feeCoins, priority, err := taxKeeper.CustomTxFeeChecker(ctx, feeTx)
	require.NoError(t, err)
	require.Equal(t, priority, int64(10000))
	require.Equal(t, sdk.NewCoins(sdk.NewInt64Coin(osmoDenom, feeAmount)), feeCoins)
}

func TestCustomTxFeeCheckerFailDueToZeroPrices(t *testing.T) {
	noPricesParams := types.Params{
		FeeRate:         types.DefaultFeeRate,
		TreasuryAddress: types.DefaultTreasuryAddress,
		BaseDenom:       types.DefaultBaseDenom,
		DexFeeParams: []*types.DexFeeParams{
			{
				ProfitAddress:           types.DefaultProfitAddress,
				AcceptedDenomsMinPrices: []*types.DenomPrice{},
			},
		},
	}
	taxKeeper, ctx := keepertest.TaxKeeper(t, true, sdk.DecCoins{sdk.NewDecCoin("unls", math.NewInt(1))}, noPricesParams)
	// create a new CustomTxFeeChecker
	feeTx := keepertest.MockFeeTx{
		Msgs: []sdk.Msg{},
		Gas:  100000,
		Fee:  sdk.Coins{sdk.NewInt64Coin(osmoDenom, feeAmount)},
	}

	_, _, err := taxKeeper.CustomTxFeeChecker(ctx, feeTx)
	require.Error(t, err)
}

// Successfully pay fees in ibc/5DE... which represents axlUSDC from osmosis. Minimum gas prices set to X unls.
func TestCustomTxFeeCheckerSuccessfulInUsdc(t *testing.T) {
	taxKeeper, ctx := keepertest.TaxKeeper(t, true, sdk.DecCoins{sdk.NewDecCoin("unls", math.NewInt(1))}, types.DefaultParams())
	// create a new CustomTxFeeChecker
	feeTx := keepertest.MockFeeTx{
		Msgs: []sdk.Msg{},
		Gas:  100000,
		Fee:  sdk.Coins{sdk.NewInt64Coin(osmoAxlUSDCDenom, feeAmount)},
	}

	feeCoins, priority, err := taxKeeper.CustomTxFeeChecker(ctx, feeTx)
	require.NoError(t, err)
	require.Equal(t, priority, int64(10000))
	require.Equal(t, sdk.NewCoins(sdk.NewInt64Coin(osmoAxlUSDCDenom, feeAmount)), feeCoins)
}

// Fail to pay fees in ibc/5DE... which represents axlUSDC from osmosis. Minimum gas prices set to unls. High gas -> fee amount not enough.
func TestCustomTxFeeCheckerFailDueToHighGasPayingWithUSDC(t *testing.T) {
	highPricesParams := types.Params{
		FeeRate:         types.DefaultFeeRate,
		TreasuryAddress: types.DefaultTreasuryAddress,
		BaseDenom:       types.DefaultBaseDenom,
		DexFeeParams: []*types.DexFeeParams{
			{
				ProfitAddress: types.DefaultProfitAddress,
				AcceptedDenomsMinPrices: []*types.DenomPrice{
					{
						Denom:    "ibc/C4CFF46FD6DE35CA4CF4CE031E643C8FDC9BA4B99AE598E9B0ED98FE3A2319F9y",
						Ticker:   "OSMO",
						MinPrice: 100.025,
					},
					{
						Denom:    "ibc/5DE4FCAF68AE40F81F738C857C0D95F7C1BC47B00FA1026E85C1DD92524D4A11",
						Ticker:   "USDC",
						MinPrice: 100.030,
					},
				},
			},
		},
	}
	taxKeeper, ctx := keepertest.TaxKeeper(t, true, sdk.DecCoins{sdk.NewDecCoin("unls", math.NewInt(1))}, highPricesParams)
	// create a new CustomTxFeeChecker
	feeTx := keepertest.MockFeeTx{
		Msgs: []sdk.Msg{},
		Gas:  1000000,
		Fee:  sdk.Coins{sdk.NewInt64Coin(osmoAxlUSDCDenom, int64(1))},
	}

	_, _, err := taxKeeper.CustomTxFeeChecker(ctx, feeTx)
	require.Error(t, err)
	require.Equal(t, "insufficient fees; got: 1.000000ibc/5DE4FCAF68AE40F81F738C857C0D95F7C1BC47B00FA1026E85C1DD92524D4A11 required: 100029998.779297ibc/5DE4FCAF68AE40F81F738C857C0D95F7C1BC47B00FA1026E85C1DD92524D4A11: insufficient fee", err.Error())
}

// Fail to pay fees in ibc/C4C... which represents OSMO. Minimum gas prices set to unls. High gas -> fee amount not enough.
func TestCustomTxFeeCheckerFailDueToHighGasPayingWithOSMO(t *testing.T) {
	highPricesParams := types.Params{
		FeeRate:         types.DefaultFeeRate,
		TreasuryAddress: types.DefaultTreasuryAddress,
		BaseDenom:       types.DefaultBaseDenom,
		DexFeeParams: []*types.DexFeeParams{
			{
				ProfitAddress: types.DefaultProfitAddress,
				AcceptedDenomsMinPrices: []*types.DenomPrice{
					{
						Denom:    "ibc/C4CFF46FD6DE35CA4CF4CE031E643C8FDC9BA4B99AE598E9B0ED98FE3A2319F9y",
						Ticker:   "OSMO",
						MinPrice: 100.025,
					},
					{
						Denom:    "ibc/5DE4FCAF68AE40F81F738C857C0D95F7C1BC47B00FA1026E85C1DD92524D4A11",
						Ticker:   "USDC",
						MinPrice: 100.030,
					},
				},
			},
		},
	}
	taxKeeper, ctx := keepertest.TaxKeeper(t, true, sdk.DecCoins{sdk.NewDecCoin("unls", math.NewInt(1))}, highPricesParams)
	// create a new CustomTxFeeChecker
	feeTx := keepertest.MockFeeTx{
		Msgs: []sdk.Msg{},
		Gas:  1000000,
		Fee:  sdk.Coins{sdk.NewInt64Coin(osmoDenom, int64(1))},
	}

	_, _, err := taxKeeper.CustomTxFeeChecker(ctx, feeTx)
	require.Error(t, err)
	require.Equal(t, "insufficient fees; got: 1.000000ibc/C4CFF46FD6DE35CA4CF4CE031E643C8FDC9BA4B99AE598E9B0ED98FE3A2319F9y required: 100025001.525879ibc/C4CFF46FD6DE35CA4CF4CE031E643C8FDC9BA4B99AE598E9B0ED98FE3A2319F9y: insufficient fee", err.Error())
}

// Successfully pay fees in unls which represents NLS. Minimum gas prices set to unls.
func TestCustomTxFeeCheckerSuccessfulInNLS(t *testing.T) {
	taxKeeper, ctx := keepertest.TaxKeeper(t, true, sdk.DecCoins{sdk.NewDecCoin("unls", math.NewInt(1))}, types.DefaultParams())
	// create a new CustomTxFeeChecker
	feeTx := keepertest.MockFeeTx{
		Msgs: []sdk.Msg{},
		Gas:  100000,
		Fee:  sdk.Coins{sdk.NewInt64Coin("unls", feeAmount)},
	}

	feeCoins, priority, err := taxKeeper.CustomTxFeeChecker(ctx, feeTx)
	require.NoError(t, err)
	require.Equal(t, priority, int64(10000))
	require.Equal(t, sdk.NewCoins(sdk.NewInt64Coin("unls", feeAmount)), feeCoins)
}

// Fail to pay fees in unsupported denom.
func TestCustomTxFeeCheckerFailWhenUnsupportedDenom(t *testing.T) {
	taxKeeper, ctx := keepertest.TaxKeeper(t, true, sdk.DecCoins{sdk.NewDecCoin("unls", math.NewInt(1))}, types.DefaultParams())
	// create a new CustomTxFeeChecker
	feeTx := keepertest.MockFeeTx{
		Msgs: []sdk.Msg{},
		Gas:  100000,
		Fee:  sdk.Coins{sdk.NewInt64Coin("unsupported", feeAmount)},
	}

	_, _, err := taxKeeper.CustomTxFeeChecker(ctx, feeTx)
	require.Error(t, err)
}

// Fail tx on zero fees provided.
func TestCustomTxFeeCheckerFailOnZeroFees(t *testing.T) {
	taxKeeper, ctx := keepertest.TaxKeeper(t, true, sdk.DecCoins{sdk.NewDecCoin("unls", math.NewInt(1))}, types.DefaultParams())
	// create a new CustomTxFeeChecker
	feeTx := keepertest.MockFeeTx{
		Msgs: []sdk.Msg{},
		Gas:  100000,
		Fee:  sdk.Coins{sdk.NewInt64Coin("unls", 0)},
	}

	_, _, err := taxKeeper.CustomTxFeeChecker(ctx, feeTx)
	require.Error(t, err)
}

// Fail to pay fees, when empty fees.
func TestCustomTxFeeCheckerFailWhenEmptyFee(t *testing.T) {
	taxKeeper, ctx := keepertest.TaxKeeper(t, true, sdk.DecCoins{sdk.NewDecCoin("unls", math.NewInt(1))}, types.DefaultParams())
	// create a new CustomTxFeeChecker
	feeTx := keepertest.MockFeeTx{}

	_, _, err := taxKeeper.CustomTxFeeChecker(ctx, feeTx)
	require.Error(t, err)
}
