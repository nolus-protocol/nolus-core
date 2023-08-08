package simapp

import (
	"time"

	"github.com/cometbft/cometbft/libs/json"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	"github.com/cosmos/cosmos-sdk/testutil/sims"

	"github.com/Nolus-Protocol/nolus-core/app"
	tmdb "github.com/cometbft/cometbft-db"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/libs/log"
	tenderminttypes "github.com/cometbft/cometbft/proto/tendermint/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	cometbfttypes "github.com/cometbft/cometbft/types"
	pruningtypes "github.com/cosmos/cosmos-sdk/store/pruning/types"
)

// New creates application instance with in-memory database and disabled logging.
func New(dir string, withDefaultGenesisState bool) *app.App {
	db := tmdb.NewMemDB()
	logger := log.NewNopLogger()

	encoding := app.MakeEncodingConfig(app.ModuleBasics)

	a := app.New(logger, db, nil, true, map[int64]bool{}, dir, 0, encoding,
		sims.EmptyAppOptions{})
	// InitChain updates deliverState which is required when app.NewContext is called
	genState := []byte("{}")
	if withDefaultGenesisState {
		genStateObj := NewDefaultGenesisState(encoding.Marshaler)
		state, err := json.MarshalIndent(genStateObj, "", " ")
		if err != nil {
			panic(err)
		}
		genState = state
	}
	a.InitChain(abci.RequestInitChain{
		ConsensusParams: defaultConsensusParams,
		AppStateBytes:   genState,
	})
	return a
}

// TODO: Improve tests that use this function in genesis_test.go files of modules and modify this function's return.
func TestSetup() (*app.App, error) {
	nolusApp := New(app.DefaultNodeHome, true)
	return nolusApp, nil
}

// NewDefaultGenesisState generates the default state for the application.
func NewDefaultGenesisState(cdc codec.JSONCodec) app.GenesisState {
	return app.ModuleBasics.DefaultGenesis(cdc)
}

var defaultConsensusParams = &tenderminttypes.ConsensusParams{
	Block: &tenderminttypes.BlockParams{
		MaxBytes: 200000,
		MaxGas:   2000000,
	},
	Evidence: &tmproto.EvidenceParams{
		MaxAgeNumBlocks: 302400,
		MaxAgeDuration:  504 * time.Hour, // 3 weeks is the max duration
		MaxBytes:        10000,
	},
	Validator: &tmproto.ValidatorParams{
		PubKeyTypes: []string{
			cometbfttypes.ABCIPubKeyTypeEd25519,
		},
	},
}

// NewAppConstructor returns a new simapp AppConstructor.
func NewAppConstructor() network.AppConstructor {
	encoding := app.MakeEncodingConfig(app.ModuleBasics)

	return func(val network.ValidatorI) servertypes.Application {
		return app.New(val.GetCtx().Logger, tmdb.NewMemDB(), nil, true, map[int64]bool{}, val.GetCtx().Config.RootDir, 0, encoding,
			sims.EmptyAppOptions{},
			baseapp.SetPruning(pruningtypes.NewPruningOptionsFromString(val.GetAppConfig().Pruning)),
			baseapp.SetMinGasPrices(val.GetAppConfig().MinGasPrices),
		)
	}
}
