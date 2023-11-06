package keeper_test

import (
	"testing"

	keepertest "github.com/Nolus-Protocol/nolus-core/testutil/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

var testAddressFrom = "nolus1932u"
var testAddressTo = "2"

// create a test function for CustomTxFeeChecker
func TestCustomTxFeeChecker(t *testing.T) {
	taxKeeper, ctx, mockWasmKeeper := keepertest.TaxKeeper(t, true, sdk.DecCoins{sdk.NewDecCoin("unls", sdk.NewInt(1))})
	// create a new CustomTxFeeChecker
	feeTx := MockFeeTx{
		Msgs: []sdk.Msg{},
		Gas:  100000,
		Fee:  sdk.Coins{sdk.NewInt64Coin("uosmo", 1000000000)},
	}

	oracleAddress, err := sdk.AccAddressFromBech32(taxKeeper.GetParams(ctx).FeeParams[0].OracleAddress)
	require.NoError(t, err)

	mockWasmKeeper.EXPECT().QuerySmart(ctx, oracleAddress, []byte(`{"prices":{}}`)).Return([]byte("1"), nil)

	feeCoins, priority, err := taxKeeper.CustomTxFeeChecker(ctx, feeTx)
	require.NoError(t, err)
	require.Equal(t, priority, int64(1))
	require.Equal(t, sdk.NewCoins(sdk.NewInt64Coin("unls", 1000000000)), feeCoins)

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
