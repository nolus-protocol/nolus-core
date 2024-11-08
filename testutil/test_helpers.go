package testutil

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"cosmossdk.io/log"
	"cosmossdk.io/math"

	icacontrollerkeeper "github.com/cosmos/ibc-go/v9/modules/apps/27-interchain-accounts/controller/keeper"
	icacontrollertypes "github.com/cosmos/ibc-go/v9/modules/apps/27-interchain-accounts/controller/types"
	icatypes "github.com/cosmos/ibc-go/v9/modules/apps/27-interchain-accounts/types"
	"github.com/cosmos/ibc-go/v9/modules/apps/transfer/types"
	ibctesting "github.com/cosmos/ibc-go/v9/testing"

	"github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	tmrand "github.com/cometbft/cometbft/libs/rand"

	db2 "github.com/cosmos/cosmos-db"

	"github.com/cosmos/cosmos-sdk/baseapp"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simcli "github.com/cosmos/cosmos-sdk/x/simulation/client/cli"
	channeltypes "github.com/cosmos/ibc-go/v9/modules/core/04-channel/types"
	"github.com/stretchr/testify/suite"

	"github.com/Nolus-Protocol/nolus-core/app"
	"github.com/Nolus-Protocol/nolus-core/app/params"
)

var (
	// TestOwnerAddress defines a reusable bech32 address for testing purposes.
	TestOwnerAddress = "nolus1ghd753shjuwexxywmgs4xz7x2q732vcnkm6h2pyv9s6ah3hylvrq8welhp"

	TestInterchainID = "owner_id"

	Connection = "connection-1"

	// TestVersion defines a reusable interchainaccounts version string for testing purposes.
	TestVersion = string(icatypes.ModuleCdc.MustMarshalJSON(&icatypes.Metadata{
		Version:                icatypes.Version,
		ControllerConnectionId: Connection,
		HostConnectionId:       Connection,
		Encoding:               icatypes.EncodingProtobuf,
		TxType:                 icatypes.TxTypeSDKMultiMsg,
	}))
)

func init() {
	ibctesting.DefaultTestingAppInit = SetupTestingApp
	params.GetDefaultConfig()
	// Disable cache since enabled cache triggers test errors when `AccAddress.String()`
	// gets called before setting nolus bech32 prefix
	sdk.SetAddrCacheEnabled(false)
}

type IBCConnectionTestSuite struct {
	suite.Suite
	Coordinator *ibctesting.Coordinator

	// testing chains used for convenience and readability
	ChainA *ibctesting.TestChain
	ChainB *ibctesting.TestChain

	// Path         *ibctesting.Path
	TransferPath *ibctesting.Path
}

func (suite *IBCConnectionTestSuite) SetupTest() {
	// we need to redefine this variable to make tests work cause we use unls as default bond denom in nolus
	sdk.DefaultBondDenom = params.DefaultBondDenom

	suite.Coordinator = ibctesting.NewCoordinator(suite.T(), 2) // initialize 2 test chains
	suite.ChainA = suite.Coordinator.GetChain(ibctesting.GetChainID(1))
	suite.ChainB = suite.Coordinator.GetChain(ibctesting.GetChainID(2))

	suite.ConfigureTransferChannel()
}

func (suite *IBCConnectionTestSuite) ConfigureTransferChannel() {
	suite.TransferPath = NewTransferPath(suite.ChainA, suite.ChainB)
	suite.Coordinator.SetupConnections(suite.TransferPath)
	err := SetupTransferPath(suite.TransferPath)
	suite.Require().NoError(err)
}

func (suite *IBCConnectionTestSuite) GetNolusZoneApp(chain *ibctesting.TestChain) *app.App {
	testApp, ok := chain.App.(*app.App)
	if !ok {
		panic("not NolusZone app")
	}

	return testApp
}

func (suite *IBCConnectionTestSuite) StoreTestCode(ctx sdk.Context, addr sdk.AccAddress, path string) uint64 {
	// wasm file built with https://github.com/neutron-org/neutron-sdk/tree/main/contracts/reflect
	// wasm file built with https://github.com/neutron-org/neutron-dev-contracts/tree/feat/ica-register-fee-update/contracts/neutron_interchain_txs
	wasmCode, err := os.ReadFile(path)
	suite.Require().NoError(err)

	codeID, _, err := keeper.NewDefaultPermissionKeeper(suite.GetNolusZoneApp(suite.ChainA).WasmKeeper).Create(ctx, addr, wasmCode, &wasmtypes.AccessConfig{Permission: wasmtypes.AccessTypeEverybody})
	suite.Require().NoError(err)

	return codeID
}

func (suite *IBCConnectionTestSuite) InstantiateTestContract(ctx sdk.Context, funder sdk.AccAddress, codeID uint64) sdk.AccAddress {
	initMsgBz := []byte("{}")
	contractKeeper := keeper.NewDefaultPermissionKeeper(suite.GetNolusZoneApp(suite.ChainA).WasmKeeper)
	addr, _, err := contractKeeper.Instantiate(ctx, codeID, funder, funder, initMsgBz, "demo contract", nil)
	suite.Require().NoError(err)

	return addr
}

func calculateTrustPeriod(unbondingPeriod time.Duration, defaultTrustPeriodFraction string) (time.Duration, error) {
	trustDec, err := math.LegacyNewDecFromStr(defaultTrustPeriodFraction)
	if err != nil {
		return time.Duration(0), err
	}
	trustPeriod := time.Duration(trustDec.MulInt64(unbondingPeriod.Nanoseconds()).TruncateInt64())

	return trustPeriod, nil
}

func NewICAPath(chainA, chainB *ibctesting.TestChain, icaowner string) *ibctesting.Path {
	path := ibctesting.NewPath(chainA, chainB)
	path.EndpointA.Counterparty = path.EndpointB
	path.EndpointB.Counterparty = path.EndpointA

	path.EndpointA.ChannelConfig.PortID = icatypes.ControllerPortPrefix + icaowner
	path.EndpointB.ChannelConfig.PortID = icatypes.HostPortID
	path.EndpointA.ChannelConfig.Order = channeltypes.UNORDERED
	path.EndpointB.ChannelConfig.Order = channeltypes.UNORDERED
	path.EndpointA.ChannelConfig.Version = TestVersion
	path.EndpointB.ChannelConfig.Version = TestVersion

	unbondingPeriodA := time.Duration(3600000000000)
	unbondingPeriodB := time.Duration(3600000000000)
	trustingPeriodFraction := "0.1"

	path.EndpointA.ClientConfig.(*ibctesting.TendermintConfig).UnbondingPeriod = unbondingPeriodA
	path.EndpointA.ClientConfig.(*ibctesting.TendermintConfig).TrustingPeriod, _ = calculateTrustPeriod(unbondingPeriodA, trustingPeriodFraction)

	path.EndpointB.ClientConfig.(*ibctesting.TendermintConfig).UnbondingPeriod = unbondingPeriodB
	path.EndpointB.ClientConfig.(*ibctesting.TendermintConfig).TrustingPeriod, _ = calculateTrustPeriod(unbondingPeriodB, trustingPeriodFraction)
	return path
}

// SetupICAPath invokes the InterchainAccounts entrypoint and subsequent channel handshake handlers.
func (suite *IBCConnectionTestSuite) SetupICAPath(path *ibctesting.Path, owner string) error {
	if err := RegisterInterchainAccount(path.EndpointA, owner); err != nil {
		return err
	}

	if err := path.EndpointB.ChanOpenTry(); err != nil {
		return err
	}

	if err := path.EndpointA.ChanOpenAck(); err != nil {
		return err
	}

	return path.EndpointB.ChanOpenConfirm()
}

// RegisterInterchainAccount is a helper function for starting the channel handshake.
func RegisterInterchainAccount(endpoint *ibctesting.Endpoint, owner string) error {
	portID, err := icatypes.NewControllerPortID(owner)
	if err != nil {
		return err
	}

	channelSequence := endpoint.Chain.App.GetIBCKeeper().ChannelKeeper.GetNextChannelSequence(endpoint.Chain.GetContext())

	a, ok := endpoint.Chain.App.(*app.App)
	if !ok {
		return fmt.Errorf("not NolusApp")
	}

	icaMsgServer := icacontrollerkeeper.NewMsgServerImpl(a.ICAControllerKeeper)
	_, err = icaMsgServer.RegisterInterchainAccount(endpoint.Chain.GetContext(), &icacontrollertypes.MsgRegisterInterchainAccount{
		Owner:        owner,
		ConnectionId: endpoint.ConnectionID,
		Version:      TestVersion,
		Ordering:     channeltypes.UNORDERED,
	})
	if err != nil {
		return err
	}
	// commit state changes for proof verification
	endpoint.Chain.NextBlock()

	// update port/channel ids
	endpoint.ChannelID = channeltypes.FormatChannelIdentifier(channelSequence)
	endpoint.ChannelConfig.PortID = portID
	endpoint.ChannelConfig.Version = TestVersion

	return nil
}

var tempDir = func() string {
	dir, err := os.MkdirTemp("", "nolusd")
	if err != nil {
		panic("failed to create temp dir: " + err.Error())
	}
	defer os.RemoveAll(dir)

	return dir
}

// SetupTestingApp initializes the IBC-go testing application.
func SetupTestingApp() (ibctesting.TestingApp, map[string]json.RawMessage) {
	db := db2.NewMemDB()
	encConfig := app.MakeEncodingConfig(app.ModuleBasics)
	chainID := "nolus-testapp" + tmrand.NewRand().Str(6)
	testApp := app.New(
		log.NewNopLogger(),
		db,
		nil,
		true,
		map[int64]bool{},
		tempDir(),
		simcli.FlagPeriodValue,
		encConfig,
		simtestutil.EmptyAppOptions{},
		baseapp.SetChainID(chainID),
		baseapp.SetMinGasPrices("0unls"),
	)

	genesisState := app.NewDefaultGenesisState(encConfig)

	return testApp, genesisState
}

func NewTransferPath(chainA, chainB *ibctesting.TestChain) *ibctesting.Path {
	path := ibctesting.NewPath(chainA, chainB)
	path.EndpointA.ChannelConfig.PortID = types.PortID
	path.EndpointB.ChannelConfig.PortID = types.PortID
	path.EndpointA.ChannelConfig.Order = channeltypes.UNORDERED
	path.EndpointB.ChannelConfig.Order = channeltypes.UNORDERED
	path.EndpointA.ChannelConfig.Version = types.Version
	path.EndpointB.ChannelConfig.Version = types.Version

	return path
}

// SetupTransferPath.
func SetupTransferPath(path *ibctesting.Path) error {
	channelSequence := path.EndpointA.Chain.App.GetIBCKeeper().ChannelKeeper.GetNextChannelSequence(path.EndpointA.Chain.GetContext())
	channelSequenceB := path.EndpointB.Chain.App.GetIBCKeeper().ChannelKeeper.GetNextChannelSequence(path.EndpointB.Chain.GetContext())

	// update port/channel ids
	path.EndpointA.ChannelID = channeltypes.FormatChannelIdentifier(channelSequence)
	path.EndpointB.ChannelID = channeltypes.FormatChannelIdentifier(channelSequenceB)

	if err := path.EndpointA.ChanOpenInit(); err != nil {
		return err
	}

	if err := path.EndpointB.ChanOpenTry(); err != nil {
		return err
	}

	if err := path.EndpointA.ChanOpenAck(); err != nil {
		return err
	}

	return path.EndpointB.ChanOpenConfirm()
}
