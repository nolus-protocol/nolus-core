package mint

import (
	"errors"
	"fmt"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/Nolus-Protocol/nolus-core/x/mint/keeper"
	"github.com/Nolus-Protocol/nolus-core/x/mint/types"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	normInitialTotal     = types.CalcTokensByIntegral(types.NormOffset)
	nanoSecondsInMonth   = sdk.NewDec(time.Hour.Nanoseconds() * 24 * 30)
	nanoSecondsInFormula = types.MonthsInFormula.Mul(nanoSecondsInMonth)
	twelveMonths         = sdk.MustNewDecFromStr("12.0")

	errTimeInFutureBeforeTimePassed = errors.New("time in future can not be before passed time")
)

func calcFunctionIncrement(nanoSecondsPassed sdkmath.Uint) sdkmath.LegacyDec {
	return types.NormMonthsRange.Mul(calcFixedIncrement(nanoSecondsPassed))
}

func calcFixedIncrement(nanoSecondsPassed sdkmath.Uint) sdkmath.LegacyDec {
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
	if minter.TotalMinted.GTE(types.MintingCap) {
		return sdkmath.ZeroUint()
	}

	if minter.PrevBlockTimestamp.IsZero() {
		// we do not know how much time has passed since the previous block, thus nothing will be mined
		minter.PrevBlockTimestamp = blockTime
		return sdkmath.ZeroUint()
	}

	nsecPassed := calcTimeDifference(blockTime, minter.PrevBlockTimestamp, maxMintableSeconds)
	if minter.NormTimePassed.LT(types.MonthsInFormula) {
		// First 96 months follow the minting formula
		// As the integral starts from NormOffset (ie > 0), previous total needs to be incremented by predetermined amount
		previousTotal := minter.TotalMinted.Add(normInitialTotal)
		newNormTime := minter.NormTimePassed.Add(calcFunctionIncrement(nsecPassed))
		nextTotal := types.CalcTokensByIntegral(newNormTime)

		delta := nextTotal.Sub(previousTotal)

		return updateMinter(minter, blockTime, newNormTime, delta)
	} else {
		// After reaching 96 normalized time, mint fixed amount of tokens per month until we reach the minting cap
		normIncrement := calcFixedIncrement(nsecPassed)
		delta := sdk.NewUint((normIncrement.Mul(types.DecFromUint(types.FixedMintedAmount))).TruncateInt().Uint64())

		if minter.TotalMinted.Add(delta).GT(types.MintingCap) {
			// Trim off excess tokens if the cap is reached
			delta = types.MintingCap.Sub(minter.TotalMinted)
		}

		return updateMinter(minter, blockTime, minter.NormTimePassed.Add(normIncrement), delta)
	}
}

func updateMinter(minter *types.Minter, blockTime sdkmath.Uint, newNormTime sdkmath.LegacyDec, newlyMinted sdkmath.Uint) sdkmath.Uint {
	if newlyMinted.LT(sdkmath.ZeroUint()) {
		// Sanity check, should not happen. However, if this were to happen,
		// do not update the minter state (primary the previous block timestamp)
		// and wait for a new block which should increase the minted amount
		return sdkmath.ZeroUint()
	}
	minter.NormTimePassed = newNormTime
	minter.PrevBlockTimestamp = blockTime
	minter.TotalMinted = minter.TotalMinted.Add(newlyMinted)
	return newlyMinted
}

// Returns the amount of tokens that should be minted by the integral formula
// for the period between normTimePassed and the timeInFuture.
func predictMintedByIntegral(totalMinted sdkmath.Uint, normTimePassed, timeAhead sdkmath.LegacyDec) (sdkmath.Uint, error) {
	timeAheadNs := timeAhead.Mul(nanoSecondsInMonth).TruncateInt()
	normTimeInFuture := normTimePassed.Add(calcFunctionIncrement(sdkmath.Uint(timeAheadNs)))
	if normTimePassed.GT(normTimeInFuture) {
		return sdkmath.ZeroUint(), errTimeInFutureBeforeTimePassed
	}

	if normTimePassed.GTE(types.MonthsInFormula) {
		return sdkmath.ZeroUint(), nil
	}

	// integral minting is caped to the 96th month
	if normTimeInFuture.GT(types.MonthsInFormula) {
		normTimeInFuture = types.MonthsInFormula
	}

	return types.CalcTokensByIntegral(normTimeInFuture).Sub(normInitialTotal).Sub(totalMinted), nil
}

// Returns the amount of tokens that should be minted during the fixed amount period
// for the period between NormTimePassed and the timeInFuture.
func predictMintedByFixedAmount(totalMinted sdkmath.Uint, normTimePassed, timeAhead sdkmath.LegacyDec) (sdkmath.Uint, error) {
	timeAheadNs := timeAhead.Mul(nanoSecondsInMonth).TruncateInt()

	normTimeInFuture := normTimePassed.Add(calcFunctionIncrement(sdkmath.Uint(timeAheadNs)))
	if normTimePassed.GT(normTimeInFuture) {
		return sdkmath.ZeroUint(), errTimeInFutureBeforeTimePassed
	}

	normFixedPeriod := normTimeInFuture.Sub(calcFunctionIncrement(sdkmath.Uint(nanoSecondsInFormula.TruncateInt())))
	if normFixedPeriod.LTE(sdk.ZeroDec()) {
		return sdkmath.ZeroUint(), nil
	}

	// convert norm time to non norm time
	fixedPeriod := normFixedPeriod.Sub(types.NormOffset).Quo(types.NormMonthsRange)

	newlyMinted := fixedPeriod.MulInt(sdkmath.Int(types.FixedMintedAmount))
	// Trim off excess tokens if the cap is reached
	if totalMinted.Add(sdkmath.Uint(newlyMinted.TruncateInt())).GT(types.MintingCap) {
		return types.MintingCap.Sub(totalMinted), nil
	}

	return sdkmath.Uint(newlyMinted.TruncateInt()), nil
}

// Returns the amount of tokens that should be minted
// between the NormTimePassed and the timeAhead
// timeAhead expects months represented in decimal form.
func predictTotalMinted(totalMinted sdkmath.Uint, normTimePassed, timeAhead sdkmath.LegacyDec) sdkmath.Uint {
	integralAmount, err := predictMintedByIntegral(totalMinted, normTimePassed, timeAhead)
	if err != nil {
		return sdkmath.ZeroUint()
	}

	fixedAmount, err := predictMintedByFixedAmount(totalMinted, normTimePassed, timeAhead)
	if err != nil {
		return sdkmath.ZeroUint()
	}

	return fixedAmount.Add(integralAmount)
}

// BeginBlocker mints new tokens for the previous block.
func BeginBlocker(ctx sdk.Context, k keeper.Keeper) {
	minter := k.GetMinter(ctx)
	if minter.TotalMinted.GTE(types.MintingCap) {
		return
	}

	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)

	params := k.GetParams(ctx)
	blockTime := ctx.BlockTime().UnixNano()
	coinAmount := calcTokens(sdk.NewUint(uint64(blockTime)), &minter, params.MaxMintableNanoseconds)

	minter.AnnualInflation = predictTotalMinted(minter.TotalMinted, minter.NormTimePassed, twelveMonths)

	ctx.Logger().Debug(fmt.Sprintf("miner: %v total, %v norm time, %v minted", minter.TotalMinted.String(), minter.NormTimePassed.String(), coinAmount.String()))

	k.SetMinter(ctx, minter)
	if coinAmount.GT(sdkmath.ZeroUint()) {
		// mint coins, update supply
		mintedCoins := sdk.NewCoins(sdk.NewCoin(params.MintDenom, sdk.NewIntFromBigInt(coinAmount.BigInt())))

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

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeMint,
			sdk.NewAttribute(types.AttributeKeyDenom, params.MintDenom),
			sdk.NewAttribute(sdk.AttributeKeyAmount, coinAmount.String()),
		),
	)
}
