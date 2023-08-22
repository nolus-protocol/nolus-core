package simulation

// import (
// 	"math/rand"

// 	"github.com/Nolus-Protocol/nolus-core/x/tax/types"
// 	sdk "github.com/cosmos/cosmos-sdk/types"
// 	"github.com/cosmos/cosmos-sdk/types/address"
// 	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
// 	"github.com/cosmos/cosmos-sdk/x/simulation"
// )

// // DONTCOVER

// // Simulation operation weights constants
// const (
// 	DefaultWeightMsgUpdateParams int = 50

// 	OpWeightMsgUpdateParams = "op_weight_msg_update_params" //nolint:gosec
// )

// // ProposalMsgs defines the module weighted proposals' contents
// func ProposalMsgs() []simtypes.WeightedProposalMsg {
// 	return []simtypes.WeightedProposalMsg{
// 		simulation.NewWeightedProposalMsg(
// 			OpWeightMsgUpdateParams,
// 			DefaultWeightMsgUpdateParams,
// 			SimulateMsgUpdateParams,
// 		),
// 	}
// }

// // SimulateMsgUpdateParams returns a random MsgUpdateParams
// func SimulateMsgUpdateParams(r *rand.Rand, _ sdk.Context, _ []simtypes.Account) sdk.Msg {
// 	// use the default gov module account address as authority
// 	var authority sdk.AccAddress = address.Module("gov")

// 	params := GenRandomFeeRate(r)

// 	return &types.MsgUpdateParams{
// 		Authority: authority.String(),
// 		Params:    params,
// 	}
// }
