package lcrepairproof

import (
	"fmt"

	"github.com/Nolus-Protocol/nolus-core/app/keepers"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/ibc-go/v10/modules/core/exported"
	ibctm "github.com/cosmos/ibc-go/v10/modules/light-clients/07-tendermint"
)

// loadTargetClient returns the client the handlers rewrite, after refusing to run anywhere
// but the throwaway localnet. Returning an error here halts the upgrade, which is the
// intended outcome: a chain that reaches this code unexpectedly should stop, not have a
// light client silently rewritten underneath it.
//
// The client is found by the chain it tracks, never by ordinal. Client IDs are assigned in
// creation order, so the same ordinal names a different client on every stack — on mainnet
// Pirin 07-tendermint-0 is the Osmosis client (ibc-go#61) — and the test suites create
// disposable clients that shift the numbering within a single run.
func loadTargetClient(ctx sdk.Context, k *keepers.AppKeepers, upgradeName string) (string, *ibctm.ClientState, error) {
	if ctx.ChainID() != LocalnetChainID {
		return "", nil, fmt.Errorf(
			"%s: refusing to rewrite client state on chain-id %q, this upgrade only runs on %q",
			upgradeName, ctx.ChainID(), LocalnetChainID,
		)
	}

	var (
		clientID string
		target   *ibctm.ClientState
		also     []string
	)

	k.IBCKeeper.ClientKeeper.IterateClientStates(ctx, nil, func(id string, cs exported.ClientState) bool {
		tmCS, ok := cs.(*ibctm.ClientState)
		if !ok || tmCS.ChainId != TargetSolanaChainID {
			return false
		}
		if target != nil {
			also = append(also, id)
			return false
		}
		clientID, target = id, tmCS
		return false
	})

	if target == nil {
		return "", nil, fmt.Errorf("%s: no client tracks %q", upgradeName, TargetSolanaChainID)
	}

	// Picking one of several would rewrite an unrelated client's store on a coin flip, and
	// nothing downstream would say which one moved.
	if len(also) > 0 {
		return "", nil, fmt.Errorf(
			"%s: %d clients track %q (%s, %v), cannot choose between them",
			upgradeName, len(also)+1, TargetSolanaChainID, clientID, also,
		)
	}

	return clientID, target, nil
}
