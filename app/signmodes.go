package app

import (
	txsigning "cosmossdk.io/x/tx/signing"
	"cosmossdk.io/x/tx/signing/aminojson"

	"github.com/Nolus-Protocol/nolus-core/eip191"
	"github.com/Nolus-Protocol/nolus-core/solanacarrier"
	"github.com/Nolus-Protocol/nolus-core/solanaoffchain"
)

// CustomSignModeHandlers builds the custom sign-mode handlers wired into the tx
// config over the given amino JSON handler. It is the single source shared by
// the app's TxConfig (app.go) and the client TxConfig (cmd/nolusd/root.go) so
// the two registration sites cannot drift apart.
func CustomSignModeHandlers(aminoHandler *aminojson.SignModeHandler) []txsigning.SignModeHandler {
	eip191Handler := eip191.NewSignModeHandler(eip191.SignModeHandlerOptions{
		AminoJsonSignModeHandler: aminoHandler,
	})
	solanaOffchainHandler := solanaoffchain.NewSignModeHandler(solanaoffchain.SignModeHandlerOptions{
		AminoJsonSignModeHandler: aminoHandler,
	})
	solanaCarrierHandler := solanacarrier.NewSignModeHandler(solanacarrier.SignModeHandlerOptions{
		AminoJsonSignModeHandler: aminoHandler,
	})
	return []txsigning.SignModeHandler{
		*eip191Handler,
		*solanaOffchainHandler,
		*solanaCarrierHandler,
	}
}
