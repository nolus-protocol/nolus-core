package solanacarrier

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	sdkcodectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

// AminoNameMsgUpdateParams must stay byte-identical to the amino.name option on
// MsgUpdateParams in tx.proto. The legacy amino codec and the aminojson sign-mode
// handler read the name from those two separate places, so a drift between them
// makes a wallet sign over one name while the chain verifies the other.
const AminoNameMsgUpdateParams = "solanacarrier/MsgUpdateParams"

// RegisterLegacyAminoCodec registers concrete types on the LegacyAmino codec.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(Params{}, "solanacarrier/Params", nil)
	legacy.RegisterAminoMsg(cdc, &MsgUpdateParams{}, AminoNameMsgUpdateParams)
}

// RegisterInterfaces registers the module's interface types.
func RegisterInterfaces(registry sdkcodectypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgUpdateParams{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
