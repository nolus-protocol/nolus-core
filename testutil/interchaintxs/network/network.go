package network

import (
	"fmt"
	"testing"
	"time"

	pruningtypes "cosmossdk.io/store/pruning/types"
	db "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/require"

	tmrand "github.com/cometbft/cometbft/libs/rand"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	"github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/Nolus-Protocol/nolus-core/app"
)

type (
	Network = network.Network
	Config  = network.Config
)

// New creates instance with fully configured cosmos network.
// Accepts optional config, that will be used in place of the DefaultConfig() if provided.
func New(t *testing.T, configs ...network.Config) *network.Network {
	if len(configs) > 1 {
		panic("at most one config should be provided")
	}
	var cfg network.Config
	if len(configs) == 0 {
		cfg = DefaultConfig()
	} else {
		cfg = configs[0]
	}
	net, err := network.New(t, t.TempDir(), cfg)
	require.NoError(t, err)
	t.Cleanup(net.Cleanup)
	return net
}

// DefaultConfig will initialize config for the network with custom application,
// genesis and single validator. All other parameters are inherited from cosmos-sdk/testutil/network.DefaultConfig.
func DefaultConfig() network.Config {
	encoding := app.MakeEncodingConfig(app.ModuleBasics)

	chainID := "chain-" + tmrand.NewRand().Str(6)
	return network.Config{
		Codec:             encoding.Marshaler,
		TxConfig:          encoding.TxConfig,
		LegacyAmino:       encoding.Amino,
		InterfaceRegistry: encoding.InterfaceRegistry,
		AccountRetriever:  authtypes.AccountRetriever{},
		AppConstructor: func(val network.ValidatorI) servertypes.Application {
			return app.New(
				val.GetCtx().Logger, db.NewMemDB(), nil, true, map[int64]bool{}, val.GetCtx().Config.RootDir,
				encoding,
				sims.EmptyAppOptions{},
				baseapp.SetPruning(pruningtypes.NewPruningOptionsFromString(val.GetAppConfig().Pruning)),
				baseapp.SetMinGasPrices(val.GetAppConfig().MinGasPrices),
				baseapp.SetChainID(chainID),
			)
		},
		GenesisState:  app.ModuleBasics.DefaultGenesis(encoding.Marshaler),
		TimeoutCommit: 2 * time.Second,
		ChainID:       chainID,
		NumValidators: 1,
		BondDenom:     sdk.DefaultBondDenom,
		MinGasPrices:  fmt.Sprintf("0.000006%s", sdk.DefaultBondDenom),
		AccountTokens: sdk.TokensFromConsensusPower(1000, sdk.DefaultPowerReduction),
		StakingTokens: sdk.TokensFromConsensusPower(500, sdk.DefaultPowerReduction),
		BondedTokens:  sdk.TokensFromConsensusPower(100, sdk.DefaultPowerReduction),

		PruningStrategy: pruningtypes.PruningOptionNothing,
		CleanupDir:      true,
		SigningAlgo:     string(hd.Secp256k1Type),
		KeyringOptions:  []keyring.Option{},
	}
}
