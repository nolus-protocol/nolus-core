package simapp

import (
	"path"
	"runtime"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/log"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/libs/json"
	tmrand "github.com/cometbft/cometbft/libs/rand"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	cmttypes "github.com/cometbft/cometbft/types"
	cometbfttypes "github.com/cometbft/cometbft/types"

	pruningtypes "cosmossdk.io/store/pruning/types"
	db "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/testutil/mock"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	"github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/Nolus-Protocol/nolus-core/app"
)

func RootDir() string {
	_, b, _, _ := runtime.Caller(0) //nolint:dogsled
	d := path.Join(path.Dir(b), "..")
	return d
}

func testHomeDir(chainID string) string {
	projectRoot := RootDir()
	return path.Join(projectRoot, ".testchains", chainID)
}

// New creates application instance with in-memory database and disabled logging.
func New(t *testing.T, withDefaultGenesisState bool) *app.App {
	database := db.NewMemDB()
	encoding := app.MakeEncodingConfig(app.ModuleBasics)
	chainID := "chain-" + tmrand.NewRand().Str(6)

	a := app.New(
		log.NewNopLogger(),
		database,
		nil,
		true,
		sims.EmptyAppOptions{},
		baseapp.SetChainID(chainID),
	)

	// InitChain updates deliverState which is required when app.NewContext is called
	genState := []byte("{}")
	if withDefaultGenesisState {
		privVal := mock.NewPV()
		pubKey, err := privVal.GetPubKey()
		require.NoError(t, err)

		// create validator set with single validator
		validator := cmttypes.NewValidator(pubKey, 1)
		valSet := cmttypes.NewValidatorSet([]*cmttypes.Validator{validator})

		// generate genesis account
		senderPrivKey := secp256k1.GenPrivKey()

		acc := authtypes.NewBaseAccount(senderPrivKey.PubKey().Address().Bytes(), senderPrivKey.PubKey(), 0, 0)
		balance := banktypes.Balance{
			Address: acc.GetAddress().String(),
			Coins:   sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(100000000000000))),
		}

		genState := app.NewDefaultGenesisState(encoding)

		nolusApp := SetupWithGenesisValSet(t, a, genState, valSet, []authtypes.GenesisAccount{acc}, chainID, balance)

		return nolusApp
	}
	_, err := a.InitChain(&abci.RequestInitChain{
		ConsensusParams: defaultConsensusParams,
		AppStateBytes:   genState,
		ChainId:         chainID,
	})
	require.NoError(t, err)

	return a
}

// SetupWithGenesisValSet initializes a new GaiaApp with a validator set and genesis accounts
// that also act as delegators. For simplicity, each validator is bonded with a delegation
// of one consensus engine unit in the default token of the GaiaApp from first genesis
// account. A Nop logger is set in GaiaApp.
func SetupWithGenesisValSet(t *testing.T, nolusApp *app.App, genesisState app.GenesisState, valSet *cmttypes.ValidatorSet, genAccs []authtypes.GenesisAccount, chainID string, balances ...banktypes.Balance) *app.App {
	t.Helper()

	genesisState = genesisStateWithValSet(t, nolusApp, genesisState, valSet, genAccs, balances...)

	stateBytes, err := json.MarshalIndent(genesisState, "", " ")
	require.NoError(t, err)

	// init chain will set the validator set and initialize the genesis accounts
	_, err = nolusApp.InitChain(
		&abci.RequestInitChain{
			Validators:      []abci.ValidatorUpdate{},
			ConsensusParams: defaultConsensusParams,
			AppStateBytes:   stateBytes,
			ChainId:         chainID,
		},
	)
	require.NoError(t, err)

	_, err = nolusApp.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height:             nolusApp.LastBlockHeight() + 1,
		Hash:               nolusApp.LastCommitID().Hash,
		NextValidatorsHash: valSet.Hash(),
		Time:               time.Now(),
	})
	require.NoError(t, err)

	// commit genesis changes
	// _, err = nolusApp.Commit()
	// require.NoError(t, err)

	return nolusApp
}

func genesisStateWithValSet(t *testing.T,
	nolusApp *app.App, genesisState app.GenesisState,
	valSet *cmttypes.ValidatorSet, genAccs []authtypes.GenesisAccount,
	balances ...banktypes.Balance,
) app.GenesisState {
	t.Helper()
	// set genesis accounts
	authGenesis := authtypes.NewGenesisState(authtypes.DefaultParams(), genAccs)
	genesisState[authtypes.ModuleName] = nolusApp.AppCodec().MustMarshalJSON(authGenesis)

	validators := make([]stakingtypes.Validator, 0, len(valSet.Validators))
	delegations := make([]stakingtypes.Delegation, 0, len(valSet.Validators))

	bondAmt := sdk.DefaultPowerReduction

	for _, val := range valSet.Validators {
		pk, err := cryptocodec.FromCmtPubKeyInterface(val.PubKey)
		require.NoError(t, err)

		pkAny, err := codectypes.NewAnyWithValue(pk)
		require.NoError(t, err)

		validator := stakingtypes.Validator{
			OperatorAddress:   sdk.ValAddress(val.Address).String(),
			ConsensusPubkey:   pkAny,
			Jailed:            false,
			Status:            stakingtypes.Bonded,
			Tokens:            bondAmt,
			DelegatorShares:   sdkmath.LegacyOneDec(),
			Description:       stakingtypes.Description{},
			UnbondingHeight:   int64(0),
			UnbondingTime:     time.Unix(0, 0).UTC(),
			Commission:        stakingtypes.NewCommission(sdkmath.LegacyZeroDec(), sdkmath.LegacyZeroDec(), sdkmath.LegacyZeroDec()),
			MinSelfDelegation: sdkmath.ZeroInt(),
		}
		validators = append(validators, validator)
		delegations = append(delegations, stakingtypes.NewDelegation(genAccs[0].GetAddress().String(), sdk.ValAddress(val.Address).String(), sdkmath.LegacyOneDec()))
	}

	// set validators and delegations
	stakingGenesis := stakingtypes.NewGenesisState(stakingtypes.DefaultParams(), validators, delegations)
	genesisState[stakingtypes.ModuleName] = nolusApp.AppCodec().MustMarshalJSON(stakingGenesis)

	totalSupply := sdk.NewCoins()
	for _, b := range balances {
		// add genesis acc tokens to total supply
		totalSupply = totalSupply.Add(b.Coins...)
	}

	for range delegations {
		// add delegated tokens to total supply
		totalSupply = totalSupply.Add(sdk.NewCoin(sdk.DefaultBondDenom, bondAmt))
	}

	// add bonded amount to bonded pool module account
	balances = append(balances, banktypes.Balance{
		Address: authtypes.NewModuleAddress(stakingtypes.BondedPoolName).String(),
		Coins:   sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, bondAmt)},
	})

	// update total supply
	bankGenesis := banktypes.NewGenesisState(banktypes.DefaultGenesisState().Params, balances, totalSupply, []banktypes.Metadata{}, []banktypes.SendEnabled{})
	genesisState[banktypes.ModuleName] = nolusApp.AppCodec().MustMarshalJSON(bankGenesis)

	return genesisState
}

func TestSetup(t *testing.T) (*app.App, error) {
	nolusApp := New(t, true)
	return nolusApp, nil
}

var defaultConsensusParams = &cmtproto.ConsensusParams{
	Block: &cmtproto.BlockParams{
		MaxBytes: 200000,
		MaxGas:   2000000,
	},
	Evidence: &cmtproto.EvidenceParams{
		MaxAgeNumBlocks: 302400,
		MaxAgeDuration:  504 * time.Hour, // 3 weeks is the max duration
		MaxBytes:        10000,
	},
	Validator: &cmtproto.ValidatorParams{
		PubKeyTypes: []string{
			cometbfttypes.ABCIPubKeyTypeEd25519,
		},
	},
}

// NewAppConstructor returns a new simapp AppConstructor.
func NewAppConstructor() network.AppConstructor {
	return func(val network.ValidatorI) servertypes.Application {
		return app.New(val.GetCtx().Logger, db.NewMemDB(), nil, true,
			sims.EmptyAppOptions{},
			baseapp.SetPruning(pruningtypes.NewPruningOptionsFromString(val.GetAppConfig().Pruning)),
			baseapp.SetMinGasPrices(val.GetAppConfig().MinGasPrices),
		)
	}
}
