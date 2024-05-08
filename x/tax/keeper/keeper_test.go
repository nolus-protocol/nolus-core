package keeper_test

import (
	"context"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"

	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"

	"github.com/cosmos/cosmos-sdk/types/tx/signing"

	nolusapp "github.com/Nolus-Protocol/nolus-core/app"
	"github.com/Nolus-Protocol/nolus-core/app/params"
	"github.com/Nolus-Protocol/nolus-core/x/tax/keeper"
	"github.com/Nolus-Protocol/nolus-core/x/tax/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdktestutil "github.com/cosmos/cosmos-sdk/testutil/testdata"

	simulationapp "github.com/Nolus-Protocol/nolus-core/testutil/simapp"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	xauthsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
)

// TestAccount represents a client Account that can be used in unit tests.
type TestAccount struct {
	acc  sdk.AccountI
	priv cryptotypes.PrivKey
}

// KeeperTestSuite is a test suite to be used with ante handler tests.
type KeeperTestSuite struct {
	suite.Suite

	app         *nolusapp.App
	ctx         sdk.Context
	clientCtx   client.Context
	txBuilder   client.TxBuilder
	anteHandler sdk.AnteHandler
	msgServer   types.MsgServer
}

// SetupTest setups a new test, with new app, context, and anteHandler.
func (s *KeeperTestSuite) SetupTest(isCheckTx bool) {
	var err error
	_ = params.GetDefaultConfig()
	s.app, err = simulationapp.TestSetup(s.T())
	s.Require().NoError(err)

	blockTime := time.Now()
	s.ctx = s.app.BaseApp.NewContext(false).WithBlockTime(blockTime)

	// set up TxConfig
	encodingConfig := nolusapp.MakeEncodingConfig()
	s.clientCtx = client.Context{}.WithTxConfig(encodingConfig.TxConfig)
	s.txBuilder = s.clientCtx.TxConfig.NewTxBuilder()
	s.Require().NoError(s.txBuilder.SetMsgs([]sdk.Msg{}...))

	anteHandler, err := ante.NewAnteHandler(
		ante.HandlerOptions{
			AccountKeeper:   s.app.AccountKeeper,
			BankKeeper:      s.app.BankKeeper,
			SignModeHandler: encodingConfig.TxConfig.SignModeHandler(),
			SigGasConsumer:  ante.DefaultSigVerificationGasConsumer,
			TxFeeChecker:    s.app.TaxKeeper.CustomTxFeeChecker,
		},
	)
	s.Require().NoError(err)

	s.anteHandler = anteHandler
	s.msgServer = keeper.NewMsgServerImpl(*s.app.TaxKeeper)
}

// CreateTestAccounts creates accounts.
func (s *KeeperTestSuite) CreateTestAccounts(numAccs int) []TestAccount {
	var accounts []TestAccount
	for i := 0; i < numAccs; i++ {
		priv, _, addr := sdktestutil.KeyTestPubAddr()
		acc := s.app.AccountKeeper.NewAccountWithAddress(s.ctx, addr)
		s.Require().NoError(acc.SetAccountNumber(uint64(i + 1000)))
		s.app.AccountKeeper.SetAccount(s.ctx, acc)
		accounts = append(accounts, TestAccount{acc, priv})
	}

	return accounts
}

// FundAcc funds target address with specified amount.
func (s *KeeperTestSuite) FundAcc(addr sdk.AccAddress, amounts sdk.Coins) {
	err := banktestutil.FundAccount(s.ctx, s.app.BankKeeper, addr, amounts)
	s.Require().NoError(err)
}

// CreateTestTx is a helper function to create a tx given multiple inputs.
func (s *KeeperTestSuite) CreateTestTx(privs []cryptotypes.PrivKey, accNums []uint64, accSeqs []uint64, chainID string) (xauthsigning.Tx, error) {
	defaultSignMode, err := xauthsigning.APISignModeToInternal(s.clientCtx.TxConfig.SignModeHandler().DefaultMode())
	if err != nil {
		return nil, err
	}

	// First round: we gather all the signer infos. We use the "set empty
	// signature" hack to do that.
	var sigsV2 []signing.SignatureV2
	for i, priv := range privs {
		sigV2 := signing.SignatureV2{
			PubKey: priv.PubKey(),
			Data: &signing.SingleSignatureData{
				SignMode:  defaultSignMode,
				Signature: nil,
			},
			Sequence: accSeqs[i],
		}

		sigsV2 = append(sigsV2, sigV2)
	}
	err = s.txBuilder.SetSignatures(sigsV2...)
	if err != nil {
		return nil, err
	}

	// Second round: all signer infos are set, so each signer can sign.
	sigsV2 = []signing.SignatureV2{}
	for i, priv := range privs {
		signerData := xauthsigning.SignerData{
			Address:       sdk.AccAddress(priv.PubKey().Bytes()).String(),
			ChainID:       chainID,
			AccountNumber: accNums[i],
			Sequence:      accSeqs[i],
			PubKey:        priv.PubKey(),
		}
		sigV2, err := tx.SignWithPrivKey(
			context.TODO(), defaultSignMode, signerData,
			s.txBuilder, priv, s.clientCtx.TxConfig, accSeqs[i])
		if err != nil {
			return nil, err
		}

		sigsV2 = append(sigsV2, sigV2)
	}
	err = s.txBuilder.SetSignatures(sigsV2...)
	if err != nil {
		return nil, err
	}

	return s.txBuilder.GetTx(), nil
}

func TestAnteTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
