package app

import (
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmTypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	taxkeeper "gitlab-nomo.credissimo.net/nomo/cosmzone/x/tax/keeper"
	taxtypes "gitlab-nomo.credissimo.net/nomo/cosmzone/x/tax/types"

	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

// HandlerOptions are the options required for constructing a default SDK AnteHandler.
type HandlerOptions struct {
	AccountKeeper     ante.AccountKeeper
	BankKeeper        taxtypes.BankKeeper
	FeegrantKeeper    ante.FeegrantKeeper
	TaxKeeper         taxkeeper.Keeper
	TxCounterStoreKey sdk.StoreKey
	WasmConfig        wasmTypes.WasmConfig
	SignModeHandler   authsigning.SignModeHandler
	SigGasConsumer    func(meter sdk.GasMeter, sig signing.SignatureV2, params types.Params) error
}

func NewAnteHandler(options HandlerOptions) (sdk.AnteHandler, error) {
	if options.AccountKeeper == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "account keeper is required for ante builder")
	}

	if options.BankKeeper == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "bank keeper is required for ante builder")
	}

	if options.SignModeHandler == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "sign mode handler is required for ante builder")
	}

	var sigGasConsumer = options.SigGasConsumer
	if sigGasConsumer == nil {
		sigGasConsumer = ante.DefaultSigVerificationGasConsumer
	}

	mempoolFeeDecorator := taxkeeper.NewMempoolFeeDecorator(options.TaxKeeper)
	deductFeeDecorator := taxkeeper.NewDeductFeeDecorator(options.AccountKeeper, options.BankKeeper, options.TaxKeeper)

	anteDecorators := []sdk.AnteDecorator{
		ante.NewSetUpContextDecorator(), // outermost AnteDecorator. SetUpContext must be called first
		wasmkeeper.NewLimitSimulationGasDecorator(options.WasmConfig.SimulationGasLimit), // after setup context to enforce limits early
		wasmkeeper.NewCountTXDecorator(options.TxCounterStoreKey),
		ante.NewRejectExtensionOptionsDecorator(),
		// ante.NewMempoolFeeDecorator(),
		mempoolFeeDecorator,
		ante.NewValidateBasicDecorator(),
		ante.NewTxTimeoutHeightDecorator(),
		ante.NewValidateMemoDecorator(options.AccountKeeper),
		ante.NewConsumeGasForTxSizeDecorator(options.AccountKeeper),
		// ante.NewDeductFeeDecorator(options.AccountKeeper, options.BankKeeper, options.FeegrantKeeper),
		deductFeeDecorator,
		ante.NewSetPubKeyDecorator(options.AccountKeeper), // SetPubKeyDecorator must be called before all signature verification decorators
		ante.NewValidateSigCountDecorator(options.AccountKeeper),
		ante.NewSigGasConsumeDecorator(options.AccountKeeper, sigGasConsumer),
		ante.NewSigVerificationDecorator(options.AccountKeeper, options.SignModeHandler),
		ante.NewIncrementSequenceDecorator(options.AccountKeeper),
	}

	return sdk.ChainAnteDecorators(anteDecorators...), nil
}
