package app

import (
	"testing"

	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/Nolus-Protocol/nolus-core/solanacarrier"
)

// The carrier rides in the transaction as an Any, so the tx decoder must be able to
// resolve its type URL through the proto registry; a Go type registered without its
// file descriptor is rejected at CheckTx with "does not have a Descriptor() method".
func TestSolanaCarrierExtensionDecodesThroughTxConfig(t *testing.T) {
	testApp, _ := CreateTestApp(true, t.TempDir())
	txConfig := testApp.GetTxConfig()

	_, err := proto.HybridResolver.FindDescriptorByName("nolus.solanacarrier.v1.SolanaCarrier")
	require.NoError(t, err, "carrier descriptor must be registered")

	carrier := &solanacarrier.SolanaCarrier{Message: []byte{0x01, 0x00, 0x00, 0x02}}
	carrierAny, err := codectypes.NewAnyWithValue(carrier)
	require.NoError(t, err)
	require.Equal(t, solanacarrier.SolanaCarrierTypeURL, carrierAny.TypeUrl)

	builder := txConfig.NewTxBuilder()
	require.NoError(t, builder.SetMsgs(&banktypes.MsgSend{
		FromAddress: sdk.AccAddress(make([]byte, 20)).String(),
		ToAddress:   sdk.AccAddress(make([]byte, 20)).String(),
		Amount:      sdk.NewCoins(sdk.NewInt64Coin("unls", 1)),
	}))
	extBuilder, ok := builder.(interface{ SetNonCriticalExtensionOptions(...*codectypes.Any) })
	require.True(t, ok, "tx builder must expose non-critical extension options")
	extBuilder.SetNonCriticalExtensionOptions(carrierAny)

	txBytes, err := txConfig.TxEncoder()(builder.GetTx())
	require.NoError(t, err)

	decoded, err := txConfig.TxDecoder()(txBytes)
	require.NoError(t, err)

	extTx, ok := decoded.(interface{ GetNonCriticalExtensionOptions() []*codectypes.Any })
	require.True(t, ok)
	exts := extTx.GetNonCriticalExtensionOptions()
	require.Len(t, exts, 1)
	require.Equal(t, solanacarrier.SolanaCarrierTypeURL, exts[0].TypeUrl)

	var decodedCarrier solanacarrier.SolanaCarrier
	require.NoError(t, decodedCarrier.Unmarshal(exts[0].Value))
	require.Equal(t, carrier.Message, decodedCarrier.Message)
}
