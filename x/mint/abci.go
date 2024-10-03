package mint

import (
	"context"
	"errors"
	"fmt"
	"time"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Nolus-Protocol/nolus-core/x/mint/keeper"
	"github.com/Nolus-Protocol/nolus-core/x/mint/types"
)

var (
	nanoSecondsInMonth = sdkmath.LegacyNewDec(time.Hour.Nanoseconds() * 24 * 30)
	twelveMonths       = sdkmath.LegacyMustNewDecFromStr("12.0")

	errTimeInFutureBeforeTimePassed = errors.New("time in future can not be before passed time")
	errNegativeBlockTime            = errors.New("block time can not be less then zero")
)

func calcFractionOfMonth(nanoSecondsPassed sdkmath.Uint) sdkmath.LegacyDec {
	return types.DecFromUint(nanoSecondsPassed).Quo(nanoSecondsInMonth)
}

func calcTimeDifference(blockTime, prevBlockTime, maxMintableSeconds sdkmath.Uint) sdkmath.Uint {
	if prevBlockTime.GT(blockTime) {
		panic("new block time cannot be smaller than previous block time")
	}

	nsecBetweenBlocks := blockTime.Sub(prevBlockTime)
	if nsecBetweenBlocks.GT(maxMintableSeconds) {
		nsecBetweenBlocks = maxMintableSeconds
	}

	return nsecBetweenBlocks
}

func calcTokens(blockTime sdkmath.Uint, minter *types.Minter, maxMintableSeconds sdkmath.Uint) sdkmath.Uint {
	if minter.PrevBlockTimestamp.IsZero() {
		// we do not know how much time has passed since the previous block, thus nothing will be mined
		minter.PrevBlockTimestamp = blockTime
		return sdkmath.ZeroUint()
	}

	nsecPassed := calcTimeDifference(blockTime, minter.PrevBlockTimestamp, maxMintableSeconds)
	if minter.NormTimePassed.LT(types.MonthsInFormula) {
		newTime := minter.NormTimePassed.Add(calcFractionOfMonth(nsecPassed))

		// First 120 months follow the minting formula
		calcWithLastNormTimePassed := types.CalcTokensByIntegral(minter.NormTimePassed)
		calcWithNewTimePassed := types.CalcTokensByIntegral(newTime)

		delta := calcWithNewTimePassed.Sub(calcWithLastNormTimePassed)

		return updateMinter(minter, blockTime, newTime, delta)
	} else {
		// After 120 months, we don't mint any more tokens.
		return sdkmath.ZeroUint()
	}
}

func updateMinter(minter *types.Minter, blockTime sdkmath.Uint, newTimePassed sdkmath.LegacyDec, newlyMinted sdkmath.Uint) sdkmath.Uint {
	if newlyMinted.LT(sdkmath.ZeroUint()) {
		// Sanity check, should not happen. However, if this were to happen,
		// do not update the minter state (primary the previous block timestamp)
		// and wait for a new block which should increase the minted amount
		return sdkmath.ZeroUint()
	}
	minter.NormTimePassed = newTimePassed
	minter.PrevBlockTimestamp = blockTime
	minter.TotalMinted = minter.TotalMinted.Add(newlyMinted)
	return newlyMinted
}

// Returns the amount of tokens that should be minted by the integral formula
// for the period between normTimePassed and the timeInFuture.
func predictMintedByIntegral(timePassed, timeAhead sdkmath.LegacyDec) (sdkmath.Uint, error) {
	timeAheadNs := timeAhead.Mul(nanoSecondsInMonth).TruncateInt()
	timeInFuture := timePassed.Add(calcFractionOfMonth(sdkmath.Uint(timeAheadNs)))
	if timePassed.GT(timeInFuture) {
		return sdkmath.ZeroUint(), errTimeInFutureBeforeTimePassed
	}

	if timePassed.GTE(types.MonthsInFormula) {
		return sdkmath.ZeroUint(), nil
	}

	// integral minting is caped to the 120th month
	if timeInFuture.GT(types.MonthsInFormula) {
		timeInFuture = types.MonthsInFormula
	}

	calcTokensTimePassed := types.CalcTokensByIntegral(timePassed)

	return types.CalcTokensByIntegral(timeInFuture).Sub(calcTokensTimePassed), nil
}

// Returns the amount of tokens that should be minted
// between the NormTimePassed and the timeAhead
// timeAhead expects months represented in decimal form.
func predictTotalMinted(timePassed, timeAhead sdkmath.LegacyDec) sdkmath.Uint {
	integralAmount, err := predictMintedByIntegral(timePassed, timeAhead)
	if err != nil {
		return sdkmath.ZeroUint()
	}

	return integralAmount
}

// BeginBlocker mints new tokens for the previous block.
func BeginBlocker(ctx context.Context, k keeper.Keeper) error {
	c := sdk.UnwrapSDKContext(ctx)
	minter := k.GetMinter(ctx)

	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)

	params := k.GetParams(ctx)
	blockTime := c.BlockTime().UnixNano()
	if blockTime < 0 {
		panic(errNegativeBlockTime)
	}

	coinAmount := calcTokens(sdkmath.NewUint(uint64(blockTime)), &minter, params.MaxMintableNanoseconds)
	minter.AnnualInflation = predictTotalMinted(minter.NormTimePassed, twelveMonths)
	c.Logger().Debug(fmt.Sprintf("miner: %v total, %v norm time, %v minted", minter.TotalMinted.String(), minter.NormTimePassed.String(), coinAmount.String()))
	err := k.SetMinter(ctx, minter)
	if err != nil {
		panic(err)
	}

	if coinAmount.GT(sdkmath.ZeroUint()) {
		// mint coins, update supply
		mintedCoins := sdk.NewCoins(sdk.NewCoin(params.MintDenom, sdkmath.NewIntFromBigInt(coinAmount.BigInt())))

		err := k.MintCoins(ctx, mintedCoins)
		if err != nil {
			panic(err)
		}

		// send the minted coins to the fee collector account
		err = k.AddCollectedFees(ctx, mintedCoins)
		if err != nil {
			panic(err)
		}

		defer telemetry.ModuleSetGauge(types.ModuleName, float32(coinAmount.Uint64()), "minted_tokens")
	}

	c.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeMint,
			sdk.NewAttribute(types.AttributeKeyDenom, params.MintDenom),
			sdk.NewAttribute(sdk.AttributeKeyAmount, coinAmount.String()),
		),
	)

	return nil
}
