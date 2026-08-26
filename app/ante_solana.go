package app

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	txsigning "github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
)

// upgradeDoneHeightGetter reports the height at which a named upgrade was
// applied, or zero if it has not been. *upgradekeeper.Keeper satisfies it.
type upgradeDoneHeightGetter interface {
	GetDoneHeight(ctx context.Context, name string) (int64, error)
}

// SolanaSignModeGateDecorator rejects transactions signed with the Solana custom
// sign modes until the upgrade that introduces them has been applied.
//
// The sign-mode handlers are wired at app construction, not in the upgrade
// handler, so a node running the new binary would otherwise accept a
// SOLANA_OFFCHAIN/SOLANA_TX_CARRIER transaction the instant it starts — before
// the coordinated upgrade height — while un-upgraded nodes reject it, diverging
// the app hash. Gating on the upgrade's recorded done-height keeps every node,
// early-swapped or not, rejecting these modes identically until the upgrade runs.
//
// The gate removes the state-root divergence but leaves a benign residual: a
// mode-192/193 tx force-included before the upgrade height is rejected here by an
// early-swapped binary and at signature verification by an un-upgraded one, whose
// differing result codes hash into LastResultsHash. Operators should not run the
// v0.8.5 binary before the upgrade height (the normal halt-and-swap flow does not),
// which avoids the mixed-binary window entirely.
type SolanaSignModeGateDecorator struct {
	upgradeKeeper upgradeDoneHeightGetter
	upgradeName   string
}

func NewSolanaSignModeGateDecorator(upgradeKeeper upgradeDoneHeightGetter, upgradeName string) SolanaSignModeGateDecorator {
	return SolanaSignModeGateDecorator{upgradeKeeper: upgradeKeeper, upgradeName: upgradeName}
}

func (d SolanaSignModeGateDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	sigTx, ok := tx.(authsigning.SigVerifiableTx)
	if !ok {
		return next(ctx, tx, simulate)
	}

	sigs, err := sigTx.GetSignaturesV2()
	if err != nil {
		return ctx, err
	}

	usesSolanaMode := false
	for _, sig := range sigs {
		if signatureUsesSolanaMode(sig.Data) {
			usesSolanaMode = true
			break
		}
	}
	if !usesSolanaMode {
		return next(ctx, tx, simulate)
	}

	doneHeight, err := d.upgradeKeeper.GetDoneHeight(ctx, d.upgradeName)
	if err != nil {
		return ctx, err
	}
	if doneHeight == 0 {
		return ctx, errorsmod.Wrapf(
			sdkerrors.ErrNotSupported,
			"sign modes SOLANA_OFFCHAIN and SOLANA_TX_CARRIER are not active until the %s upgrade is applied",
			d.upgradeName,
		)
	}

	return next(ctx, tx, simulate)
}

func signatureUsesSolanaMode(data txsigning.SignatureData) bool {
	switch d := data.(type) {
	case *txsigning.SingleSignatureData:
		return isSolanaSignMode(d.SignMode)
	case *txsigning.MultiSignatureData:
		for _, sub := range d.Signatures {
			if signatureUsesSolanaMode(sub) {
				return true
			}
		}
	}
	return false
}

func isSolanaSignMode(mode txsigning.SignMode) bool {
	return mode == txsigning.SignMode_SIGN_MODE_SOLANA_OFFCHAIN ||
		mode == txsigning.SignMode_SIGN_MODE_SOLANA_TX_CARRIER
}
