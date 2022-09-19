package mint

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"gitlab-nomo.credissimo.net/nomo/cosmzone/x/mint/types"
)

var (
	expectedCoins60Sec      = sdk.MustNewDecFromStr("78685482213530").TruncateInt()
	expectedNormTime20Sec   = sdk.MustNewDecFromStr("95.999976965179227961")
	normTimeThreshold       = sdk.MustNewDecFromStr("0.0001")
	fiveMinutesInNano       = time.Minute.Nanoseconds() * 5
	expectedTokensInFormula = []int64{2005393779413, 1960065211135, 1915701583592, 1872292817451, 1829821427258,
		1788274927426, 1747641763937, 1707904450838, 1669053479559, 1631072341866, 1593949550419, 1557669553370,
		1522218863850, 1487586973442, 1453755394621, 1420715576293, 1388449010421, 1356946272040, 1326192726904, 1296171971295,
		1266873433466, 1238282626336, 1210387040993, 1183170189298, 1156621501589, 1130725550707, 1105469766714, 1080839681981,
		1056822747547, 1033404455241, 1010572355682, 988309900621, 966607542783, 945449851062, 924822259131, 904712277524,
		885106416684, 865990111420, 847350928368, 829175302677, 811447763141, 794156745187, 777288815041, 760828407932,
		744763034158, 729079184970, 713763370269, 698803044753, 684180718108, 669887881926, 655907045544, 642227664463,
		628832247789, 615711287449, 602849274537, 590231718334, 577847074639, 565680852982, 553718544290, 541947658048, 530354632472,
		518925029729, 507645305912, 496503917208, 485484425469, 474574269364, 463760957802, 453029982379, 442366834512,
		431760023172, 421194004353, 410656287539, 400132381360, 389609724801, 379073879265, 368512283906, 357910464737, 347253860867,
		336532033940, 325728422712, 314830535989, 303824883001, 292696902129, 281434155665, 270023081451, 258449189544,
		246699988128, 234759933673, 222617535475, 210259284621, 197668672536, 184835208599, 171745348920, 158381602124,
		144736459803, 130799835426,
	}
)

func TestTimeDifference(t *testing.T) {
	_, _, _, timeOffset := defaultParams()
	tb := time.Second.Nanoseconds() * 60 // 60 seconds
	td := calcTimeDifference(timeOffset+tb, timeOffset, fiveMinutesInNano)

	require.Equal(t, td, tb)
}

func TestTimeDifference_MoreThenMax(t *testing.T) {
	_, _, _, timeOffset := defaultParams()
	tb := fiveMinutesInNano + 1
	td := calcTimeDifference(timeOffset+tb, timeOffset, fiveMinutesInNano)

	require.Equal(t, td, fiveMinutesInNano)
}

func TestTimeDifference_InvalidTime(t *testing.T) {
	_, _, _, timeOffset := defaultParams()
	td := calcTimeDifference(timeOffset-1, timeOffset, fiveMinutesInNano)

	require.Zero(t, td)
}

func Test_CalcTokensDuringFormula_WhenUsingConstantIncrements_OutputsPredeterminedAmount(t *testing.T) {
	timeBetweenBlocks := time.Second.Nanoseconds() * 60 // 60 seconds per block
	minutesInMonth := int64(60) * 24 * 30
	minutesInFormula := minutesInMonth * types.MonthsInFormula.TruncateInt64()
	minter, mintedCoins, mintedMonth, timeOffset := defaultParams()
	for i := int64(0); i < minutesInFormula; i++ {
		coins := calcTokens(timeOffset+i*timeBetweenBlocks, &minter, fiveMinutesInNano)
		mintedCoins = mintedCoins.Add(coins)
		mintedMonth = mintedMonth.Add(coins)
		if i%minutesInMonth == 0 {
			fmt.Printf("%v Month, %v Minted, %v Total Minted(in store), %v Returned Total, %v Norm Time, %v Recieved in this block \n",
				i/minutesInMonth, mintedMonth, minter.TotalMinted, mintedCoins, minter.NormTimePassed, coins)
			mintedMonth = sdk.ZeroInt()
		}
	}
	fmt.Printf("%v Returned Total, %v Total Minted(in store), %v Norm Time \n",
		mintedCoins, minter.TotalMinted, minter.NormTimePassed)
	if !expectedCoins60Sec.Equal(mintedCoins) || !expectedCoins60Sec.Equal(minter.TotalMinted) {
		t.Errorf("Minted unexpected amount of tokens, expected %v returned and in store, actual minted %v, actual in store %v",
			expectedCoins60Sec, mintedCoins, minter.TotalMinted)
	}
	if !expectedNormTime20Sec.Equal(minter.NormTimePassed) {
		t.Errorf("Received unexpected normalized time, expected %v, actual %v", expectedNormTime20Sec, minter.NormTimePassed)
	}
}

func randomTimeBetweenBlocks(min int64, max int64) int64 {
	return time.Second.Nanoseconds() * (rand.Int63n(max-min) + min)
}

func Test_CalcTokensDuringFormula_WhenUsingVaryingIncrements_OutputExpectedTokensWithinEpsilon(t *testing.T) {
	minter, mintedCoins, mintedMonth, timeOffset := defaultParams()
	prevOffset := timeOffset
	nanoSecondsInPeriod := (nanoSecondsInMonth.Mul(types.MonthsInFormula)).Add(sdk.NewDec(timeOffset)).TruncateInt64()
	rand.Seed(time.Now().UnixNano())
	monthThreshold := sdk.NewInt(100_000_000) // 100 tokens
	month := 0
	for i := int64(0); timeOffset < nanoSecondsInPeriod; {
		i = randomTimeBetweenBlocks(5, 60)
		coins := calcTokens(timeOffset+i, &minter, fiveMinutesInNano)
		if coins.LT(sdk.ZeroInt()) {
			t.Errorf("Minted negative %v coins", coins)
		}
		mintedCoins = mintedCoins.Add(coins)
		mintedMonth = mintedMonth.Add(coins)
		if (timeOffset-prevOffset)/nanoSecondsInMonth.TruncateInt64() != (timeOffset+i-prevOffset)/nanoSecondsInMonth.TruncateInt64() {
			month += 1
			fmt.Printf("%v Month, %v Minted, %v Total Minted(in store), %v Returned Total, %v Norm Time, %v Received in this block \n",
				month, mintedMonth, minter.TotalMinted, mintedCoins, minter.NormTimePassed, coins)
			if mintedMonth.Sub(sdk.NewInt(expectedTokensInFormula[month-1])).Abs().GT(monthThreshold) {
				t.Errorf("Minted unexpected amount of tokens for month %d, expected [%v +/- 100*10^6], actual %v",
					month, expectedTokensInFormula[month-1], mintedMonth)
			}
			prevOffset = timeOffset
			mintedMonth = sdk.ZeroInt()
			rand.Seed(time.Now().UnixNano())
		}
		timeOffset += i
	}
	mintThreshold := sdk.NewInt(10_000_000) // 10 tokens
	fmt.Printf("%v Returned Total, %v Total Minted(in store), %v Norm Time \n", mintedCoins, minter.TotalMinted, minter.NormTimePassed)
	if expectedCoins60Sec.Sub(mintedCoins).Abs().GT(mintThreshold) || expectedCoins60Sec.Sub(minter.TotalMinted).Abs().GT(mintThreshold) {
		t.Errorf("Minted unexpected amount of tokens, expected [%v +/- 10*10^6] returned and in store, actual minted %v, actual in store %v",
			expectedCoins60Sec, mintedCoins, minter.TotalMinted)
	}
	if expectedNormTime20Sec.Sub(minter.NormTimePassed).Abs().GT(normTimeThreshold) {
		t.Errorf("Received unexpected normalized time, expected [%v +/- 0.000001], actual %v",
			expectedNormTime20Sec, minter.NormTimePassed)
	}
}

func Test_CalcTokensFixed_WhenNotHittingMintCapInAMonth_OutputsExpectedTokensWithinEpsilon(t *testing.T) {
	timeOffset := time.Now().UnixNano()
	offsetNanoInMonth := nanoSecondsInMonth.Add(sdk.NewDec(timeOffset)).TruncateInt64()
	minter := types.NewMinter(types.MonthsInFormula, sdk.ZeroInt(), timeOffset)
	mintedCoins := sdk.NewInt(0)
	rand.Seed(time.Now().UnixNano())
	for i := int64(0); timeOffset < offsetNanoInMonth; {
		i = randomTimeBetweenBlocks(5, 60)
		coins := calcTokens(timeOffset+i, &minter, fiveMinutesInNano)
		if coins.LT(sdk.ZeroInt()) {
			t.Errorf("Minted negative %v coins", coins)
		}
		mintedCoins = mintedCoins.Add(coins)
		timeOffset += i
	}
	fmt.Printf("%v Returned Total, %v Total Minted(in store), %v Norm Time \n",
		mintedCoins, minter.TotalMinted, minter.NormTimePassed)
	mintThreshold := sdk.NewInt(1_300_000) // 1.3 tokens is the max deviation
	if types.FixedMintedAmount.Sub(mintedCoins).Abs().GT(mintThreshold) || types.FixedMintedAmount.Sub(minter.TotalMinted).Abs().GT(mintThreshold) {
		t.Errorf("Minted unexpected amount of tokens, expected [%v +/- 10^6] returned and in store, actual minted %v, actual in store %v",
			types.FixedMintedAmount, mintedCoins, minter.TotalMinted)
	}
	if (types.MonthsInFormula.Add(sdk.OneDec())).Sub(minter.NormTimePassed).Abs().GT(normTimeThreshold) {
		t.Errorf("Received unexpected normalized time, expected [%v +/- 0.000001], actual %v", expectedNormTime20Sec, minter.NormTimePassed)
	}
}

func Test_CalcTokensFixed_WhenHittingMintCapInAMonth_DoesNotExceedMaxMintingCap(t *testing.T) {
	timeOffset := time.Now().UnixNano()
	offsetNanoInMonth := nanoSecondsInMonth.Add(sdk.NewDec(timeOffset)).TruncateInt64()
	halfFixedAmount := types.FixedMintedAmount.QuoRaw(2)
	totalMinted := types.MintingCap.Sub(halfFixedAmount)
	minter := types.NewMinter(types.MonthsInFormula, totalMinted, timeOffset)
	mintedCoins := sdk.NewInt(0)
	rand.Seed(time.Now().UnixNano())
	for i := int64(0); timeOffset < offsetNanoInMonth; {
		i = randomTimeBetweenBlocks(5, 60)
		coins := calcTokens(timeOffset+i, &minter, fiveMinutesInNano)
		mintedCoins = mintedCoins.Add(coins)
		timeOffset += i
	}
	fmt.Printf("%v Returned Total, %v Total Minted(in store), %v Norm Time \n",
		mintedCoins, minter.TotalMinted, minter.NormTimePassed)
	mintThreshold := sdk.NewInt(1_000_000) // 1 token
	if types.MintingCap.Sub(minter.TotalMinted).Abs().GT(sdk.ZeroInt()) {
		t.Errorf("Minting Cap exeeded, minted total %v, with minting cap %v",
			minter.TotalMinted, types.MintingCap)
	}
	if halfFixedAmount.Sub(mintedCoins).Abs().GT(mintThreshold) {
		t.Errorf("Minted unexpected amount of tokens, expected [%v +/- 10^6] returned and in store, actual minted %v",
			halfFixedAmount, mintedCoins)
	}
	if (types.MonthsInFormula.Add(sdk.MustNewDecFromStr("0.5"))).Sub(minter.NormTimePassed).Abs().GT(normTimeThreshold) {
		t.Errorf("Received unexpected normalized time, expected [%v +/- 0.000001], actual %v",
			types.MonthsInFormula.Add(sdk.MustNewDecFromStr("0.5")), minter.NormTimePassed)
	}
}

func Test_CalcTokens_WhenMintingAllTokens_OutputsExactExpectedTokens(t *testing.T) {
	minter, mintedCoins, mintedMonth, timeOffset := defaultParams()
	prevOffset := timeOffset
	offsetNanoInPeriod := (nanoSecondsInMonth.Mul(sdk.NewDec(121))).Add(sdk.NewDec(timeOffset)).TruncateInt64() // Adding 1 extra to ensure cap is preserved
	month := 0
	rand.Seed(time.Now().UnixNano())
	for i := int64(0); timeOffset < offsetNanoInPeriod; {
		i = randomTimeBetweenBlocks(60, 120)
		coins := calcTokens(timeOffset+i, &minter, fiveMinutesInNano)
		mintedCoins = mintedCoins.Add(coins)
		mintedMonth = mintedMonth.Add(coins)
		if (timeOffset-prevOffset)/nanoSecondsInMonth.TruncateInt64() != (timeOffset+i-prevOffset)/nanoSecondsInMonth.TruncateInt64() {
			month += 1
			rand.Seed(time.Now().UnixNano())
			fmt.Printf("%v Month, %v Minted, %v Total Minted(in store), %v Returned Total, %v Norm Time, %v Received in this block \n",
				month, mintedMonth, minter.TotalMinted, mintedCoins, minter.NormTimePassed, coins)
			prevOffset = timeOffset
			mintedMonth = sdk.ZeroInt()
		}
		timeOffset += i
	}
	fmt.Printf("%v Returned Total, %v Total Minted(in store), %v Norm Time \n",
		mintedCoins, minter.TotalMinted, minter.NormTimePassed)
	require.Equal(t, types.MintingCap, minter.TotalMinted)
	require.Equal(t, minter.TotalMinted, mintedCoins)
}

func Test_CalcTokens_WhenGivenBlockWithDiffBiggerThanMax_MaxMintedTokensAreCreated(t *testing.T) {
	timeOffset := time.Now()
	nextOffset := timeOffset.Add(time.Hour)
	originalMinter := types.InitialMinter()
	originalMinter.PrevBlockTimestamp = timeOffset.UnixNano()
	minter := types.InitialMinter()
	minter.PrevBlockTimestamp = timeOffset.UnixNano()

	coins := calcTokens(nextOffset.UnixNano(), &minter, fiveMinutesInNano)
	expectedCoins := calcTokens(timeOffset.UnixNano()+fiveMinutesInNano, &originalMinter, fiveMinutesInNano)
	require.Equal(t, expectedCoins, coins)
}

func Test_CalcIncrementDuringFormula_OutputsExpectedIncrementWithinEpsilon(t *testing.T) {
	increment5s := calcFunctionIncrement(time.Second.Nanoseconds() * 5)
	increment30s := calcFunctionIncrement(time.Second.Nanoseconds() * 30)
	increment60s := calcFunctionIncrement(time.Second.Nanoseconds() * 60)
	minutesInPeriod := int64(60) * 24 * 30 * types.MonthsInFormula.TruncateInt64()
	sumIncrements5s := sdk.NewDec(12 * minutesInPeriod).Mul(increment5s)
	sumIncrements30s := sdk.NewDec(2 * minutesInPeriod).Mul(increment30s)
	sumIncrements60s := sdk.NewDec(1 * minutesInPeriod).Mul(increment60s)
	if sumIncrements5s.Sub(types.AbsMonthsRange).Abs().GT(normTimeThreshold) {
		t.Errorf("Increment with 5 second step results in range %v, deviating with more than epsilon from expected %v",
			sumIncrements5s, types.AbsMonthsRange)
	}
	if sumIncrements30s.Sub(types.AbsMonthsRange).Abs().GT(normTimeThreshold) {
		t.Errorf("Increment with 30 second step results in range %v, deviating with more than epsilon from expected %v",
			sumIncrements30s, types.AbsMonthsRange)
	}
	if sumIncrements60s.Sub(types.AbsMonthsRange).Abs().GT(normTimeThreshold) {
		t.Errorf("Increment with 60 second step results in range %v, deviating with more than epsilon from expected %v",
			sumIncrements60s, types.AbsMonthsRange)
	}
}

func Test_CalcFixedIncrement_OutputsExpectedIncrementWithinEpsilon(t *testing.T) {
	increment5s := calcFixedIncrement(time.Second.Nanoseconds() * 5)
	increment30s := calcFixedIncrement(time.Second.Nanoseconds() * 30)
	increment60s := calcFixedIncrement(time.Second.Nanoseconds() * 60)
	minutesInMonth := int64(60) * 24 * 30
	sumIncrements5s := sdk.NewDec(12 * minutesInMonth).Mul(increment5s)
	sumIncrements30s := sdk.NewDec(2 * minutesInMonth).Mul(increment30s)
	sumIncrements60s := sdk.NewDec(1 * minutesInMonth).Mul(increment60s)
	if sumIncrements5s.Sub(sdk.OneDec()).Abs().GT(normTimeThreshold) {
		t.Errorf("Increment with 5 second step results in range %v, deviating with more than epsilon from expected %v",
			sumIncrements5s, sdk.OneDec())
	}
	if sumIncrements30s.Sub(sdk.OneDec()).Abs().GT(normTimeThreshold) {
		t.Errorf("Increment with 30 second step results in range %v, deviating with more than epsilon from expected %v",
			sumIncrements30s, sdk.OneDec())
	}
	if sumIncrements60s.Sub(sdk.OneDec()).Abs().GT(normTimeThreshold) {
		t.Errorf("Increment with 60 second step results in range %v, deviating with more than epsilon from expected %v",
			sumIncrements60s, sdk.OneDec())
	}
}

func defaultParams() (types.Minter, sdk.Int, sdk.Int, int64) {
	minter := types.InitialMinter()
	mintedCoins := sdk.NewInt(0)
	mintedMonth := sdk.NewInt(0)
	return minter, mintedCoins, mintedMonth, time.Now().UnixNano()
}
