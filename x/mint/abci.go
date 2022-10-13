package mint

import (
	"fmt"
	"time"

	"gitlab-nomo.credissimo.net/nomo/cosmzone/custom/util"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gitlab-nomo.credissimo.net/nomo/cosmzone/x/mint/keeper"
	"gitlab-nomo.credissimo.net/nomo/cosmzone/x/mint/types"
)

// Minting formula f(x)=-4.33275 x^3 + 944.61206 x^2 - 88567.25194 x + 3.86335×10^6 integrated over 0.47 to 96
// afterwards minting 55000 tokens each month until reaching the minting cap of 80*10^6 tokens
var (
	quadCoef           = sdk.MustNewDecFromStr("-1.08319")
	cubeCoef           = sdk.MustNewDecFromStr("314.871")
	squareCoef         = sdk.MustNewDecFromStr("-44283.6")
	coef               = sdk.MustNewDecFromStr("3863350")
	normInitialTotal   = util.ConvertToMicroNolusDec(calcIntegral(types.NormOffset))
	nanoSecondsInMonth = sdk.NewDecFromInt(sdk.NewInt(30).Mul(sdk.NewInt(24)).Mul(sdk.NewInt(60)).Mul(sdk.NewInt(60))).Mul(sdk.NewDec(10).Power(9))
)

func calcFunctionIncrement(nanoSecondsPassed int64) sdk.Dec {
	timePassed := sdk.NewDec(nanoSecondsPassed).Quo(nanoSecondsInMonth)
	return types.NormMonthsRange.Mul(timePassed)
}

func calcFixedIncrement(nanoSecondsPassed int64) sdk.Dec {
	return sdk.NewDec(nanoSecondsPassed).Quo(nanoSecondsInMonth)
}

func calcTimeDifference(blockTime int64, prevBlockTime int64, maxMintableSeconds int64) int64 {
	// cast to int to check for overflow
	nsecBetweenBlocks := sdk.NewInt(blockTime).Sub(sdk.NewInt(prevBlockTime)).Int64()
	if nsecBetweenBlocks > maxMintableSeconds {
		nsecBetweenBlocks = maxMintableSeconds
	}
	if nsecBetweenBlocks < 0 {
		// sanity check, this should never happen
		nsecBetweenBlocks = 0
	}
	return nsecBetweenBlocks
}

// Integral:  -1.08319 x^4 + 314.871 x^3 - 44283.6 x^2 + 3.86335×10^6 x
func calcIntegral(x sdk.Dec) sdk.Dec {
	xToPower4 := x.Power(4)
	xToPower3 := x.Power(3)
	xToPower2 := x.Power(2)
	return (quadCoef.Mul(xToPower4)).Add(cubeCoef.Mul(xToPower3)).Add(squareCoef.Mul(xToPower2)).Add(coef.Mul(x))
}

func calcTokens(blockTime int64, minter *types.Minter, maxMintableSeconds int64) sdk.Int {
	if minter.TotalMinted.GTE(types.MintingCap) {
		return sdk.ZeroInt()
	}

	if minter.PrevBlockTimestamp == 0 {
		// we do not know how much time has passed since the previous block, thus nothing will be mined
		minter.PrevBlockTimestamp = blockTime
		return sdk.ZeroInt()
	}

	nsecPassed := calcTimeDifference(blockTime, minter.PrevBlockTimestamp, maxMintableSeconds)
	if minter.NormTimePassed.LT(types.MonthsInFormula) {
		// First 96 months follow the minting formula
		// As the integral starts from NormOffset (ie > 0), previous total needs to be incremented by predetermined amount
		previousTotal := minter.TotalMinted.Add(normInitialTotal)
		newNormTime := minter.NormTimePassed.Add(calcFunctionIncrement(nsecPassed))
		nextIntegral := calcIntegral(newNormTime)

		delta := util.ConvertToMicroNolusDec(nextIntegral).Sub(previousTotal)

		return updateMinter(minter, blockTime, newNormTime, delta)
	} else {
		// After reaching 96 normalized time, mint fixed amount of tokens per month until we reach the minting cap
		normIncrement := calcFixedIncrement(nsecPassed)
		delta := (normIncrement.MulInt(types.FixedMintedAmount)).TruncateInt()

		if minter.TotalMinted.Add(delta).GT(types.MintingCap) {
			// Trim off excess tokens if the cap is reached
			delta = types.MintingCap.Sub(minter.TotalMinted)
		}

		return updateMinter(minter, blockTime, minter.NormTimePassed.Add(normIncrement), delta)
	}
}

func updateMinter(minter *types.Minter, blockTime int64, newNormTime sdk.Dec, deltaInt sdk.Int) sdk.Int {
	if deltaInt.LT(sdk.ZeroInt()) {
		// Sanity check, should not happen. However, if this were to happen,
		// do not update the minter state (primary the previous block timestamp)
		// and wait for a new block which should increase the minted amount
		return sdk.ZeroInt()
	}
	minter.NormTimePassed = newNormTime
	minter.PrevBlockTimestamp = blockTime
	minter.TotalMinted = minter.TotalMinted.Add(deltaInt)
	return deltaInt
}

// BeginBlocker mints new tokens for the previous block.
func BeginBlocker(ctx sdk.Context, k keeper.Keeper) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)

	// fetch stored minter & params
	minter := k.GetMinter(ctx)
	params := k.GetParams(ctx)

	coinAmount := calcTokens(ctx.BlockTime().UnixNano(), &minter, params.MaxMintableNanoseconds)
	ctx.Logger().Debug(fmt.Sprintf("miner: %v total, %v norm time, %v minted", minter.TotalMinted.String(), minter.NormTimePassed.String(), coinAmount.String()))

	k.SetMinter(ctx, minter)
	if coinAmount.IsPositive() {
		// mint coins, update supply
		mintedCoins := sdk.NewCoins(sdk.NewCoin(params.MintDenom, coinAmount))

		err := k.MintCoins(ctx, mintedCoins)
		if err != nil {
			panic(err)
		}

		// send the minted coins to the fee collector account
		err = k.AddCollectedFees(ctx, mintedCoins)
		if err != nil {
			panic(err)
		}
		if coinAmount.IsInt64() {
			defer telemetry.ModuleSetGauge(types.ModuleName, float32(coinAmount.Int64()), "minted_tokens")
		}
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeMint,
			sdk.NewAttribute(types.AttributeKeyDenom, params.MintDenom),
			sdk.NewAttribute(sdk.AttributeKeyAmount, coinAmount.String()),
		),
	)
}
