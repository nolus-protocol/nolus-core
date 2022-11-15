package mint

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"gitlab-nomo.credissimo.net/nomo/nolus-core/x/mint/types"
)

var (
	expectedCoins60Sec      = sdk.MustNewDecFromStr("147535251163101").TruncateInt()
	expectedNormTime20Sec   = sdk.MustNewDecFromStr("95.999976965179227961")
	normTimeThreshold       = sdk.MustNewDecFromStr("0.0001")
	fiveMinutesInNano       = time.Minute.Nanoseconds() * 5
	expectedTokensInFormula = []int64{3759989678764, 3675042190671, 3591959455921, 3510492761731,
		3430894735556, 3352957640645, 3276743829430, 3202299947048, 3129456689610, 3058269447752,
		2988635387331, 2920578426197, 2854081215478, 2789206352970, 2725694534821, 2663751456612,
		2603392455686, 2544229951801, 2486586312604, 2430266441855, 2375370186150, 2321790458408,
		2269471441070, 2218404771718, 2168676312642, 2120106446370, 2072730033868, 2026569326236,
		1981502209633, 1937662505273, 1894776317892, 1853047004297, 1812400331621, 1772672353478,
		1734048861880, 1696309269859, 1659581717535, 1623727160904, 1588788875111, 1554677158679,
		1521456570881, 1489014024628, 1457399071357, 1426569077784, 1396420182708, 1367035157098,
		1338295086558, 1310242203296, 1282828095325, 1256078098440, 1229806138168, 1204170896528,
		1179101215073, 1154454133557, 1130346086275, 1106674161508, 1083471952854, 1060674232689,
		1038228517507, 1016164727286, 994432638055, 972990871407, 951827390485, 930951168170,
		910275692536, 889867731221, 869554578094, 849428221672, 829451772393, 809554555703,
		789741963552, 769980269406, 750255593001, 730519445685, 710773940732, 690976498669,
		671085262062, 651131377798, 631008346988, 610747683954, 590316688226, 569699591966,
		548832098993, 527723637864, 506308309490, 484613046274, 462585855489, 440201215001,
		417441533390, 394262128601, 370667272533, 346593248698, 322056063563, 297009874733,
		271424507593, 245276309734,
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
	require.Panics(t, assert.PanicTestFunc(func() {
		calcTimeDifference(timeOffset-1, timeOffset, fiveMinutesInNano)
	}))
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
	monthThreshold := sdk.NewInt(187_500_000) // 187.5 tokens
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
	mintThreshold := sdk.NewInt(2_437_500) // 2.4375 tokens is the max deviation
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
