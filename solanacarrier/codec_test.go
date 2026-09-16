package solanacarrier_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/api/amino"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/gogoproto/proto"
	protov2 "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"

	"github.com/Nolus-Protocol/nolus-core/solanacarrier"
)

// The legacy amino codec and the aminojson sign-mode handler take the name from
// two separate places. If they drift, a wallet signs over one name while the
// chain verifies the other and the signature silently fails to match.
func TestAminoNameMatchesTheProtoOption(t *testing.T) {
	desc, err := proto.HybridResolver.FindDescriptorByName("nolus.solanacarrier.v1.MsgUpdateParams")
	require.NoError(t, err)

	msgDesc, ok := desc.(protoreflect.MessageDescriptor)
	require.True(t, ok, "MsgUpdateParams must resolve to a message descriptor")

	opts, ok := msgDesc.Options().(*descriptorpb.MessageOptions)
	require.True(t, ok, "message options must be present")
	require.True(t, protov2.HasExtension(opts, amino.E_Name), "MsgUpdateParams must carry an amino.name option")

	require.Equal(t,
		solanacarrier.AminoNameMsgUpdateParams,
		protov2.GetExtension(opts, amino.E_Name).(string),
	)
}

// RegisterLegacyAminoCodec must actually bind the message to that name, not just
// agree with the proto file about what it should be.
func TestLegacyAminoCodecRegistersMsgUpdateParamsUnderTheAminoName(t *testing.T) {
	cdc := codec.NewLegacyAmino()
	solanacarrier.RegisterLegacyAminoCodec(cdc)

	bz, err := cdc.MarshalJSON(&solanacarrier.MsgUpdateParams{
		Authority: "nolus10d07y265gmmuvt4z0w9aw880jnsr700js7zslc",
		Params:    solanacarrier.DefaultParams(),
	})
	require.NoError(t, err)
	require.Contains(t, string(bz), `"type":"`+solanacarrier.AminoNameMsgUpdateParams+`"`)
}
