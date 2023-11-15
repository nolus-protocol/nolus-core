package keeper_test

import (
	"errors"
	"testing"

	keepertest "github.com/Nolus-Protocol/nolus-core/testutil/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

// The bytes below represent this string: {"prices":[{"amount":{"amount":"20000000","ticker":"OSMO"},"amount_quote":{"amount":"4248067","ticker":"USDC"}},{"amount":{"amount":"2000000000","ticker":"NLS"},"amount_quote":{"amount":"10452150388158391","ticker":"USDC"}}]}
var queryPricesResponseBytes = []byte{123, 34, 112, 114, 105, 99, 101, 115, 34, 58, 91, 123, 34, 97, 109, 111, 117, 110, 116, 34, 58, 123, 34, 97, 109, 111, 117, 110, 116, 34, 58, 34, 50, 48, 48, 48, 48, 48, 48, 48, 34, 44, 34, 116, 105, 99, 107, 101, 114, 34, 58, 34, 79, 83, 77, 79, 34, 125, 44, 34, 97, 109, 111, 117, 110, 116, 95, 113, 117, 111, 116, 101, 34, 58, 123, 34, 97, 109, 111, 117, 110, 116, 34, 58, 34, 52, 50, 52, 56, 48, 54, 55, 34, 44, 34, 116, 105, 99, 107, 101, 114, 34, 58, 34, 85, 83, 68, 67, 34, 125, 125, 44, 123, 34, 97, 109, 111, 117, 110, 116, 34, 58, 123, 34, 97, 109, 111, 117, 110, 116, 34, 58, 34, 50, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 34, 44, 34, 116, 105, 99, 107, 101, 114, 34, 58, 34, 78, 76, 83, 34, 125, 44, 34, 97, 109, 111, 117, 110, 116, 95, 113, 117, 111, 116, 101, 34, 58, 123, 34, 97, 109, 111, 117, 110, 116, 34, 58, 34, 49, 48, 52, 53, 50, 49, 53, 48, 51, 56, 56, 49, 53, 56, 51, 57, 49, 34, 44, 34, 116, 105, 99, 107, 101, 114, 34, 58, 34, 85, 83, 68, 67, 34, 125, 125, 93, 125}

// Successfully pay fees in ibc/C4C... which represents OSMO. Minimum gas prices set to unls.
func TestCustomTxFeeCheckerSuccessfulInOsmo(t *testing.T) {
	taxKeeper, ctx, mockWasmKeeper := keepertest.TaxKeeper(t, true, sdk.DecCoins{sdk.NewDecCoin("unls", sdk.NewInt(1))})
	// create a new CustomTxFeeChecker
	feeTx := MockFeeTx{
		Msgs: []sdk.Msg{},
		Gas:  100000,
		Fee:  sdk.Coins{sdk.NewInt64Coin("ibc/C4CFF46FD6DE35CA4CF4CE031E643C8FDC9BA4B99AE598E9B0ED98FE3A2319F9y", 1000000000)},
	}

	oracleAddress, err := sdk.AccAddressFromBech32(taxKeeper.GetParams(ctx).FeeParams[0].OracleAddress)
	require.NoError(t, err)

	mockWasmKeeper.EXPECT().QuerySmart(ctx, oracleAddress, []byte(`{"prices":{}}`)).Return(queryPricesResponseBytes, nil)

	feeCoins, priority, err := taxKeeper.CustomTxFeeChecker(ctx, feeTx)
	require.NoError(t, err)
	require.Equal(t, priority, int64(10000))
	require.Equal(t, sdk.NewCoins(sdk.NewInt64Coin("ibc/C4CFF46FD6DE35CA4CF4CE031E643C8FDC9BA4B99AE598E9B0ED98FE3A2319F9y", 1000000000)), feeCoins)
}

// Successfully pay fees in unls which represents NLS. Minimum gas prices set to unls.
func TestCustomTxFeeCheckerSuccessfulInNLS(t *testing.T) {
	taxKeeper, ctx, _ := keepertest.TaxKeeper(t, true, sdk.DecCoins{sdk.NewDecCoin("unls", sdk.NewInt(1))})
	// create a new CustomTxFeeChecker
	feeTx := MockFeeTx{
		Msgs: []sdk.Msg{},
		Gas:  100000,
		Fee:  sdk.Coins{sdk.NewInt64Coin("unls", 1000000000)},
	}

	feeCoins, priority, err := taxKeeper.CustomTxFeeChecker(ctx, feeTx)
	require.NoError(t, err)
	require.Equal(t, priority, int64(10000))
	require.Equal(t, sdk.NewCoins(sdk.NewInt64Coin("unls", 1000000000)), feeCoins)
}

// Fail to pay fees in unsupported denom.
func TestCustomTxFeeCheckerSuccessfulInUnsupportedDenom(t *testing.T) {
	taxKeeper, ctx, _ := keepertest.TaxKeeper(t, true, sdk.DecCoins{sdk.NewDecCoin("unls", sdk.NewInt(1))})
	// create a new CustomTxFeeChecker
	feeTx := MockFeeTx{
		Msgs: []sdk.Msg{},
		Gas:  100000,
		Fee:  sdk.Coins{sdk.NewInt64Coin("unsupported", 1000000000)},
	}

	_, _, err := taxKeeper.CustomTxFeeChecker(ctx, feeTx)
	require.Error(t, err)
}

// Successfully pay fees in ibc/C4C... which represents OSMO. Minimum gas prices set to unls.
func TestCustomTxFeeCheckerWithWrongOracleAddr(t *testing.T) {
	taxKeeper, ctx, _ := keepertest.TaxKeeper(t, true, sdk.DecCoins{sdk.NewDecCoin("unls", sdk.NewInt(1))})
	// create a new CustomTxFeeChecker
	feeTx := MockFeeTx{
		Msgs: []sdk.Msg{},
		Gas:  100000,
		Fee:  sdk.Coins{sdk.NewInt64Coin("ibc/C4CFF46FD6DE35CA4CF4CE031E643C8FDC9BA4B99AE598E9B0ED98FE3A2319F9y", 1000000000)},
	}

	wrongParams := taxKeeper.GetParams(ctx)
	wrongParams.FeeParams[0].OracleAddress = "wrong"
	taxKeeper.SetParams(ctx, wrongParams)

	_, priority, err := taxKeeper.CustomTxFeeChecker(ctx, feeTx)
	require.Error(t, err)
	require.Equal(t, priority, int64(0))
}

// Successfully pay fees in ibc/C4C... which represents OSMO. Minimum gas prices set to unls.
func TestCustomTxFeeCheckerPricesQueryReturnsError(t *testing.T) {
	taxKeeper, ctx, mockWasmKeeper := keepertest.TaxKeeper(t, true, sdk.DecCoins{sdk.NewDecCoin("unls", sdk.NewInt(1))})
	// create a new CustomTxFeeChecker
	feeTx := MockFeeTx{
		Msgs: []sdk.Msg{},
		Gas:  100000,
		Fee:  sdk.Coins{sdk.NewInt64Coin("ibc/C4CFF46FD6DE35CA4CF4CE031E643C8FDC9BA4B99AE598E9B0ED98FE3A2319F9y", 1000000000)},
	}

	oracleAddress, err := sdk.AccAddressFromBech32(taxKeeper.GetParams(ctx).FeeParams[0].OracleAddress)
	require.NoError(t, err)

	mockWasmKeeper.EXPECT().QuerySmart(ctx, oracleAddress, []byte(`{"prices":{}}`)).Return([]byte{}, errors.New("badQuery"))

	_, _, err = taxKeeper.CustomTxFeeChecker(ctx, feeTx)
	require.Error(t, err)
}

// Successfully pay fees in ibc/C4C... which represents OSMO. Minimum gas prices set to unls.
func TestCustomTxFeeCheckerPriceQueryReturnsNoPrices(t *testing.T) {
	taxKeeper, ctx, mockWasmKeeper := keepertest.TaxKeeper(t, true, sdk.DecCoins{sdk.NewDecCoin("unls", sdk.NewInt(1))})
	// create a new CustomTxFeeChecker
	feeTx := MockFeeTx{
		Msgs: []sdk.Msg{},
		Gas:  100000,
		Fee:  sdk.Coins{sdk.NewInt64Coin("ibc/C4CFF46FD6DE35CA4CF4CE031E643C8FDC9BA4B99AE598E9B0ED98FE3A2319F9y", 1000000000)},
	}

	oracleAddress, err := sdk.AccAddressFromBech32(taxKeeper.GetParams(ctx).FeeParams[0].OracleAddress)
	require.NoError(t, err)

	mockWasmKeeper.EXPECT().QuerySmart(ctx, oracleAddress, []byte(`{"prices":{}}`)).Return([]byte{}, nil)

	_, _, err = taxKeeper.CustomTxFeeChecker(ctx, feeTx)
	require.Error(t, err)
}

// Successfully pay fees in ibc/C4C... which represents OSMO. Minimum gas prices set to unls.
func TestCustomTxFeeCheckerPriceQueryReturnsPricesOnlyForOsmo(t *testing.T) {
	taxKeeper, ctx, mockWasmKeeper := keepertest.TaxKeeper(t, true, sdk.DecCoins{sdk.NewDecCoin("unls", sdk.NewInt(1))})
	// create a new CustomTxFeeChecker
	feeTx := MockFeeTx{
		Msgs: []sdk.Msg{},
		Gas:  100000,
		Fee:  sdk.Coins{sdk.NewInt64Coin("ibc/C4CFF46FD6DE35CA4CF4CE031E643C8FDC9BA4B99AE598E9B0ED98FE3A2319F9y", 1000000000)},
	}

	byteOsmoPrices := []byte(`{"prices":[{"amount":{"amount":"20000000","ticker":"OSMO"},"amount_quote":{"amount":"4248067","ticker":"USDC"}}]}`)

	oracleAddress, err := sdk.AccAddressFromBech32(taxKeeper.GetParams(ctx).FeeParams[0].OracleAddress)
	require.NoError(t, err)

	mockWasmKeeper.EXPECT().QuerySmart(ctx, oracleAddress, []byte(`{"prices":{}}`)).Return(byteOsmoPrices, nil)

	_, _, err = taxKeeper.CustomTxFeeChecker(ctx, feeTx)
	require.Error(t, err)
}

type MockFeeTx struct {
	Msgs []sdk.Msg
	Gas  uint64
	Fee  sdk.Coins
}

func (m MockFeeTx) GetMsgs() []sdk.Msg {
	return m.Msgs
}

func (m MockFeeTx) ValidateBasic() error {
	// Implement your basic validation logic here or return nil if not needed for the test.
	return nil
}

func (m MockFeeTx) GetGas() uint64 {
	return m.Gas
}

func (m MockFeeTx) GetFee() sdk.Coins {
	return m.Fee
}

func (m MockFeeTx) FeePayer() sdk.AccAddress {
	return sdk.AccAddress{}
}

func (m MockFeeTx) FeeGranter() sdk.AccAddress {
	return sdk.AccAddress{}
}
