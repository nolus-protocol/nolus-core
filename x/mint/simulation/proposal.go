package simulation

// DONTCOVER

// refactor: decide if we want to run such simulations

// Simulation operation weights constants
// const (
// 	DefaultWeightMsgUpdateParams int = 100

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

// SimulateMsgUpdateParams returns a random MsgUpdateParams
// func SimulateMsgUpdateParams(r *rand.Rand, _ sdk.Context, _ []simtypes.Account) sdk.Msg {
// 	// use the default gov module account address as authority
// 	var authority sdk.AccAddress = address.Module("gov")

// 	// refactor: align upper and lower ranges when we start running simulationss
// 	params := types.NewParams(
// 		types.DefaultParams().MintDenom,
// 		RandomMaxMintableNanoSeconds(r, 1, 6000000),
// 	)

// 	return types.NewMsgUpdateParams(
// 		params, authority.String(),
// 	)
// }
