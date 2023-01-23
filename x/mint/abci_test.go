package mint

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/Nolus-Protocol/nolus-core/custom/util"

	"github.com/stretchr/testify/assert"

	"github.com/Nolus-Protocol/nolus-core/x/mint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

var (
	expectedCoins60Sec      = sdk.NewUint(147535251163101)
	expectedNormTime20Sec   = sdk.MustNewDecFromStr("95.999976965179227961")
	normTimeThreshold       = sdk.MustNewDecFromStr("0.0001")
	fiveMinutesInNano       = sdk.NewUint(uint64(time.Minute.Nanoseconds() * 5))
	expectedTokensInFormula = []int64{
		3759989678764, 3675042190671, 3591959455921, 3510492761731,
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
	tb := sdk.NewUint(uint64(time.Second.Nanoseconds() * 60)) // 60 seconds
	td := calcTimeDifference(timeOffset.Add(tb), timeOffset, fiveMinutesInNano)

	require.Equal(t, td, tb)
}

func TestTimeDifference_MoreThenMax(t *testing.T) {
	_, _, _, timeOffset := defaultParams()
	tb := fiveMinutesInNano.Add(sdk.NewUint(1))
	td := calcTimeDifference(timeOffset.Add(tb), timeOffset, fiveMinutesInNano)

	require.Equal(t, td, fiveMinutesInNano)
}

func TestTimeDifference_InvalidTime(t *testing.T) {
	_, _, _, timeOffset := defaultParams()
	require.Panics(t, assert.PanicTestFunc(func() {
		calcTimeDifference(timeOffset, timeOffset.Add(sdk.NewUint(1)), fiveMinutesInNano)
	}))
}

func Test_CalcTokensDuringFormula_WhenUsingConstantIncrements_OutputsPredeterminedAmount(t *testing.T) {
	timeBetweenBlocks := sdk.NewUint(uint64(time.Second.Nanoseconds() * 60)) // 60 seconds per block
	minutesInMonth := uint64(time.Hour.Minutes()) * 24 * 30
	minutesInFormula := minutesInMonth * uint64(types.MonthsInFormula.TruncateInt64())
	minter, mintedCoins, mintedMonth, timeOffset := defaultParams()

	for i := uint64(0); i < minutesInFormula; i++ {
		coins := calcTokens(timeOffset.Add(sdk.NewUint(i).Mul(timeBetweenBlocks)), &minter, fiveMinutesInNano)

		mintedCoins = mintedCoins.Add(sdk.NewUint(coins.Uint64()))
		mintedMonth = mintedMonth.Add(sdk.NewUint(coins.Uint64()))

		if i%minutesInMonth == 0 {
			fmt.Printf("%v Month, %v Minted, %v Total Minted(in store), %v Returned Total, %v Norm Time, %v Received in this block \n",
				i/minutesInMonth, mintedMonth, minter.TotalMinted, mintedCoins, minter.NormTimePassed, coins)
			mintedMonth = sdk.ZeroUint()
		}
	}

	fmt.Printf("%v Returned Total, %v Total Minted(in store), %v Norm Time \n",
		mintedCoins, minter.TotalMinted, minter.NormTimePassed)

	if !expectedCoins60Sec.Equal(mintedCoins) || !expectedCoins60Sec.Equal(sdk.NewUint(minter.TotalMinted.Uint64())) {
		t.Errorf("Minted unexpected amount of tokens, expected %v returned and in store, actual minted %v, actual in store %v",
			expectedCoins60Sec, mintedCoins, minter.TotalMinted)
	}
	if !expectedNormTime20Sec.Equal(minter.NormTimePassed) {
		t.Errorf("Received unexpected normalized time, expected %v, actual %v", expectedNormTime20Sec, minter.NormTimePassed)
	}
}

func Test_CalcTokensDuringFormula_WhenUsingVaryingIncrements_OutputExpectedTokensWithinEpsilon(t *testing.T) {
	minter, mintedCoins, mintedMonth, timeOffset := defaultParams()
	prevOffset := timeOffset
	nanoSecondsInPeriod := nanoSecondsInMonth.Mul(types.MonthsInFormula).Add(types.DecFromUint(timeOffset)).TruncateInt64()
	rand.Seed(util.GetCurrentTimeUnixNano())
	monthThreshold := sdk.NewUint(187_500_000) // 187.5 tokens
	month := 0

	for timeOffset.LT(sdk.NewUint(uint64(nanoSecondsInPeriod))) {
		i := sdk.NewUint(randomTimeBetweenBlocks(5, 60))

		coins := calcTokens(timeOffset.Add(i), &minter, fiveMinutesInNano)
		if coins.LT(sdk.ZeroUint()) {
			t.Errorf("Minted negative %v coins", coins)
		}

		mintedCoins = mintedCoins.Add(sdk.NewUint(coins.Uint64()))
		mintedMonth = mintedMonth.Add(sdk.NewUint(coins.Uint64()))

		prevI := timeOffset.Sub(prevOffset)
		nanoSecondsInMonthUint := sdk.NewUint(uint64(nanoSecondsInMonth.TruncateInt64()))
		// TODO: understand what is a and b and name accordingly
		a := prevI.Quo(nanoSecondsInMonthUint)
		b := prevI.Add(i).Quo(nanoSecondsInMonthUint)

		if !a.Equal(b) {
			month++

			fmt.Printf("%v Month, %v Minted, %v Total Minted(in store), %v Returned Total, %v Norm Time, %v Received in this block \n",
				month, mintedMonth, minter.TotalMinted, mintedCoins, minter.NormTimePassed, coins)

			if types.GetAbsDiff(mintedMonth, sdk.NewUint(uint64(expectedTokensInFormula[month-1]))).GT(monthThreshold) {
				t.Errorf("Minted unexpected amount of tokens for month %d, expected [%v +/- %v], actual %v",
					month, expectedTokensInFormula[month-1], monthThreshold, mintedMonth)
			}

			prevOffset = timeOffset
			mintedMonth = sdk.ZeroUint()
			rand.Seed(util.GetCurrentTimeUnixNano())
		}

		timeOffset = timeOffset.Add(i)
	}

	mintThreshold := sdk.NewUint(10_000_000) // 10 tokens
	fmt.Printf("%v Returned Total, %v Total Minted(in store), %v Norm Time \n", mintedCoins, minter.TotalMinted, minter.NormTimePassed)

	if types.GetAbsDiff(expectedCoins60Sec, mintedCoins).GT(mintThreshold) || types.GetAbsDiff(expectedCoins60Sec, sdk.Uint(minter.TotalMinted)).GT(mintThreshold) {
		t.Errorf("Minted unexpected amount of tokens, expected [%v +/- %v] returned and in store, actual minted %v, actual in store %v",
			expectedCoins60Sec, mintThreshold, mintedCoins, minter.TotalMinted)
	}

	if expectedNormTime20Sec.Sub(minter.NormTimePassed).Abs().GT(normTimeThreshold) {
		t.Errorf("Received unexpected normalized time, expected [%v +/- %v], actual %v",
			expectedNormTime20Sec, normTimeThreshold, minter.NormTimePassed)
	}
}

func Test_CalcTokensFixed_WhenNotHittingMintCapInAMonth_OutputsExpectedTokensWithinEpsilon(t *testing.T) {
	_, _, _, timeOffset := defaultParams()

	offsetNanoInMonth := timeOffset.Add(uintFromDec(nanoSecondsInMonth))
	minter := types.NewMinter(types.MonthsInFormula, sdk.ZeroUint(), timeOffset, sdk.ZeroUint())
	mintedCoins := sdk.ZeroUint()
	rand.Seed(util.GetCurrentTimeUnixNano())

	for timeOffset.LT(offsetNanoInMonth) {
		i := sdk.NewUint(randomTimeBetweenBlocks(5, 60))
		coins := calcTokens(timeOffset.Add(i), &minter, fiveMinutesInNano)

		if coins.LT(sdk.ZeroUint()) {
			t.Errorf("Minted negative %v coins", coins)
		}

		mintedCoins = mintedCoins.Add(coins)
		timeOffset = timeOffset.Add(i)
	}

	fmt.Printf("%v Returned Total, %v Total Minted(in store), %v Norm Time \n",
		mintedCoins, minter.TotalMinted, minter.NormTimePassed)
	mintThreshold := sdk.NewUint(2_437_500) // 2.4375 tokens is the max deviation

	if types.GetAbsDiff(types.FixedMintedAmount, mintedCoins).GT(mintThreshold) || types.GetAbsDiff(types.FixedMintedAmount, minter.TotalMinted).GT(mintThreshold) {
		t.Errorf("Minted unexpected amount of tokens, expected [%v +/- %v] returned and in store, actual minted %v, actual in store %v",
			types.FixedMintedAmount, mintThreshold, mintedCoins, minter.TotalMinted)
	}

	if (types.MonthsInFormula.Add(sdk.OneDec())).Sub(minter.NormTimePassed).Abs().GT(normTimeThreshold) {
		t.Errorf("Received unexpected normalized time, expected [%v +/- %v], actual %v", expectedNormTime20Sec, normTimeThreshold, minter.NormTimePassed)
	}
}

func Test_CalcTokensFixed_WhenHittingMintCapInAMonth_DoesNotExceedMaxMintingCap(t *testing.T) {
	_, _, _, timeOffset := defaultParams()

	offsetNanoInMonth := timeOffset.Add(uintFromDec(nanoSecondsInMonth))

	halfFixedAmount := types.FixedMintedAmount.Quo(sdk.NewUint(2))
	totalMinted := types.MintingCap.Sub(halfFixedAmount)
	minter := types.NewMinter(types.MonthsInFormula, totalMinted, timeOffset, sdk.ZeroUint())
	mintedCoins := sdk.NewUint(0)
	rand.Seed(util.GetCurrentTimeUnixNano())

	for timeOffset.LT(offsetNanoInMonth) {
		i := sdk.NewUint(randomTimeBetweenBlocks(5, 60))

		coins := calcTokens(timeOffset.Add(i), &minter, fiveMinutesInNano)
		mintedCoins = mintedCoins.Add(coins)
		timeOffset = timeOffset.Add(i)
	}

	fmt.Printf("%v Returned Total, %v Total Minted(in store), %v Norm Time \n",
		mintedCoins, minter.TotalMinted, minter.NormTimePassed)
	mintThreshold := sdk.NewUint(1_000_000) // 1 token
	if types.MintingCap.Sub(minter.TotalMinted).GT(sdk.ZeroUint()) {
		t.Errorf("Minting Cap exeeded, minted total %v, with minting cap %v",
			minter.TotalMinted, types.MintingCap)
	}
	if types.GetAbsDiff(halfFixedAmount, mintedCoins).GT(mintThreshold) {
		t.Errorf("Minted unexpected amount of tokens, expected [%v +/- %v] returned and in store, actual minted %v",
			halfFixedAmount, mintThreshold, mintedCoins)
	}
	if (types.MonthsInFormula.Add(sdk.MustNewDecFromStr("0.5"))).Sub(minter.NormTimePassed).Abs().GT(normTimeThreshold) {
		t.Errorf("Received unexpected normalized time, expected [%v +/- %v], actual %v",
			types.MonthsInFormula.Add(sdk.MustNewDecFromStr("0.5")), normTimeThreshold, minter.NormTimePassed)
	}
}

func Test_CalcTokens_WhenMintingAllTokens_OutputsExactExpectedTokens(t *testing.T) {
	minter, mintedCoins, mintedMonth, timeOffset := defaultParams()
	prevOffset := timeOffset
	offsetNanoInPeriod := uintFromDec((nanoSecondsInMonth.Mul(sdk.NewDec(121))).Add(types.DecFromUint(timeOffset))) // Adding 1 extra to ensure cap is preserved
	month := 0
	rand.Seed(util.GetCurrentTimeUnixNano())

	for timeOffset.LT(offsetNanoInPeriod) {
		i := sdk.NewUint(randomTimeBetweenBlocks(60, 120))

		coins := calcTokens(timeOffset.Add(i), &minter, fiveMinutesInNano)
		mintedCoins = mintedCoins.Add(sdk.NewUint(coins.Uint64()))
		mintedMonth = mintedMonth.Add(sdk.NewUint(coins.Uint64()))

		prevI := timeOffset.Sub(prevOffset)
		nanoSecondsInMonthUint := sdk.NewUint(uint64(nanoSecondsInMonth.TruncateInt64()))
		// TODO: understand what is a and b and name accordingly
		a := prevI.Quo(nanoSecondsInMonthUint)
		b := prevI.Add(i).Quo(nanoSecondsInMonthUint)

		if !a.Equal(b) {
			month++

			rand.Seed(util.GetCurrentTimeUnixNano())
			fmt.Printf("%v Month, %v Minted, %v Total Minted(in store), %v Returned Total, %v Norm Time, %v Received in this block \n",
				month, mintedMonth, minter.TotalMinted, mintedCoins, minter.NormTimePassed, coins)
			prevOffset = timeOffset
			mintedMonth = sdk.ZeroUint()
		}

		timeOffset = timeOffset.Add(i)
	}

	fmt.Printf("%v Returned Total, %v Total Minted(in store), %v Norm Time \n",
		mintedCoins, minter.TotalMinted, minter.NormTimePassed)

	require.Equal(t, types.MintingCap, minter.TotalMinted)
	require.EqualValues(t, minter.TotalMinted, mintedCoins)
}

func Test_CalcTokens_WhenGivenBlockWithDiffBiggerThanMax_MaxMintedTokensAreCreated(t *testing.T) {
	timeOffset := time.Now()
	timeOffsetUint := sdk.NewUint(uint64(timeOffset.UnixNano()))
	nextOffset := sdk.NewUint(uint64(timeOffset.Add(time.Hour).UnixNano()))

	originalMinter := types.InitialMinter()
	originalMinter.PrevBlockTimestamp = sdk.NewUint(uint64(timeOffset.UnixNano()))

	minter := types.InitialMinter()
	minter.PrevBlockTimestamp = sdk.NewUint(uint64(timeOffset.UnixNano()))

	coins := calcTokens(nextOffset, &minter, fiveMinutesInNano)
	expectedCoins := calcTokens(timeOffsetUint.Add(fiveMinutesInNano), &originalMinter, fiveMinutesInNano)

	require.Equal(t, expectedCoins, coins)
}

func Test_CalcIncrementDuringFormula_OutputsExpectedIncrementWithinEpsilon(t *testing.T) {
	increment5s := calcFunctionIncrement(sdk.NewUint(uint64(time.Second.Nanoseconds() * 5)))
	increment30s := calcFunctionIncrement(sdk.NewUint(uint64(time.Second.Nanoseconds() * 30)))
	increment60s := calcFunctionIncrement(sdk.NewUint(uint64(time.Second.Nanoseconds() * 60)))

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
	increment5s := calcFixedIncrement(sdk.NewUint(uint64(time.Second.Nanoseconds() * 5)))
	increment30s := calcFixedIncrement(sdk.NewUint(uint64(time.Second.Nanoseconds() * 30)))
	increment60s := calcFixedIncrement(sdk.NewUint(uint64(time.Second.Nanoseconds() * 60)))

	minutesInMonth := int64(time.Hour.Minutes()) * 24 * 30
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

func Test_PredictMintedByIntegral_TwelveMonthsAhead(t *testing.T) {
	expAcceptedDeviation := sdk.NewUint(500_000) // 0.5 token

	for _, tc := range []struct {
		title             string
		normTimePassed    sdk.Dec
		timeAhead         sdk.Dec
		totalMinted       sdk.Uint
		expIntegralMinted sdk.Uint
		expError          bool
	}{
		{
			title:             "start from genesis, 1 month calculated by integral",
			normTimePassed:    sdk.MustNewDecFromStr("0.47"),
			timeAhead:         sdk.MustNewDecFromStr("1"),
			totalMinted:       sdk.ZeroUint(),
			expIntegralMinted: sdk.NewUintFromString("3_760_114_000_000"),
			expError:          false,
		},
		{
			title:             "start from genesis, 12 months calculated by integral",
			normTimePassed:    sdk.MustNewDecFromStr("0.47"),
			timeAhead:         twelveMonths,
			totalMinted:       sdk.ZeroUint(),
			expIntegralMinted: sdk.NewUintFromString("39_897_845_000_000"),
			expError:          false,
		},
		{
			title:             "in the 96 months range, 12 months calculated by integral",
			normTimePassed:    sdk.MustNewDecFromStr("5.44552083"),
			timeAhead:         twelveMonths,
			totalMinted:       sdk.NewUintFromString("14_537_732_000_000"),
			expIntegralMinted: sdk.NewUintFromString("38_996_481_000_000"),
			expError:          false,
		},
		{
			title:             "ends on the 96th month, 12 months calculated by integral",
			normTimePassed:    sdk.MustNewDecFromStr("84.05875000"),
			timeAhead:         twelveMonths,
			totalMinted:       sdk.NewUintFromString("142_977_230_000_000"),
			expIntegralMinted: sdk.NewUintFromString("4_558_027_000_000"),
			expError:          false,
		},
		{
			title:             "partially in the 96 months range, 1 month calculated by integral",
			normTimePassed:    sdk.MustNewDecFromStr("95.00489583"),
			timeAhead:         twelveMonths,
			totalMinted:       sdk.NewUintFromString("147_290_028_000_000"),
			expIntegralMinted: sdk.NewUintFromString("245_229_000_000"),
			expError:          false,
		},
		{
			title:             "after 96th months, 0 months calculated by integral",
			normTimePassed:    sdk.MustNewDecFromStr("98"),
			timeAhead:         twelveMonths,
			totalMinted:       sdk.NewUintFromString("147_741_507_000_000"),
			expIntegralMinted: sdk.ZeroUint(),
			expError:          false,
		},
		{
			title:             "negative time ahead should result in error",
			normTimePassed:    sdk.MustNewDecFromStr("98"),
			timeAhead:         sdk.MustNewDecFromStr("-1.0"),
			totalMinted:       sdk.ZeroUint(),
			expIntegralMinted: sdk.ZeroUint(),
			expError:          true,
		},
		{
			title:             "zero time ahead should not mint tokens",
			normTimePassed:    sdk.MustNewDecFromStr("85.05385417"),
			timeAhead:         sdk.ZeroDec(),
			totalMinted:       sdk.NewUintFromString("143_483_520_000_000"),
			expIntegralMinted: sdk.ZeroUint(),
			expError:          false,
		},
	} {
		t.Run(tc.title, func(t *testing.T) {
			minter := &types.Minter{
				NormTimePassed: tc.normTimePassed,
				TotalMinted:    tc.totalMinted,
			}

			newlyMinted, err := predictMintedByIntegral(minter.TotalMinted, minter.NormTimePassed, tc.timeAhead)
			if tc.expError && err == nil {
				t.Error("Error is expected")
			}

			actExpDiff := types.GetAbsDiff(newlyMinted, tc.expIntegralMinted)

			if actExpDiff.GT(expAcceptedDeviation) {
				t.Errorf("Minted exp: %v, act: %v, diff: %v", tc.expIntegralMinted, newlyMinted, actExpDiff)
			}
		})
	}
}

func Test_PredictMintedByFixedAmount_TwelveMonthsAhead(t *testing.T) {
	expAcceptedDeviation := sdk.NewUint(500) // 0.0005 token

	for _, tc := range []struct {
		title          string
		normTimePassed sdk.Dec
		timeAhead      sdk.Dec
		totalMinted    sdk.Uint
		expFixedMinted sdk.Uint
		expError       bool
	}{
		{
			title:          "in the 96 months range, 0 months calculated by fixed amount",
			normTimePassed: sdk.MustNewDecFromStr("0.47"),
			timeAhead:      twelveMonths,
			totalMinted:    sdk.ZeroUint(),
			expFixedMinted: sdk.ZeroUint(),
			expError:       false,
		},
		{
			title:          "starts on the 96th month, 1 month calculated by fixed amount",
			normTimePassed: sdk.MustNewDecFromStr("96"),
			timeAhead:      sdk.MustNewDecFromStr("1"),
			totalMinted:    sdk.NewUintFromString("147_535_257_000_000"),
			expFixedMinted: sdk.NewUintFromString("103_125_000_000"),
			expError:       false,
		},
		{
			title:          "partially in the 96 months range, 1 month calculated by fixed amount",
			normTimePassed: sdk.MustNewDecFromStr("85.05385417"),
			timeAhead:      twelveMonths,
			totalMinted:    sdk.NewUintFromString("143_483_520_000_000"),
			expFixedMinted: sdk.NewUintFromString("103_125_000_000"),
			expError:       false,
		},
		{
			title:          "starts on the 96th month, all months calculated by fixed amount",
			normTimePassed: sdk.MustNewDecFromStr("96"),
			timeAhead:      twelveMonths,
			totalMinted:    sdk.NewUintFromString("147_535_257_000_000"),
			expFixedMinted: sdk.NewUintFromString("103_125_000_000").MulUint64(12),
			expError:       false,
		},
		{
			title:          "partially in the 96-120 month range, few days calculated by fixed amount",
			normTimePassed: sdk.MustNewDecFromStr("119.0"),
			timeAhead:      twelveMonths,
			totalMinted:    sdk.NewUintFromString("149_900_000_000_000"),
			expFixedMinted: sdk.NewUintFromString("100_000_000_000"),
			expError:       false,
		},
		{
			title:          "after minting cap reached, 0 months calculated by fixed amount",
			normTimePassed: sdk.MustNewDecFromStr("119.9"),
			timeAhead:      twelveMonths,
			totalMinted:    sdk.NewUintFromString("150_000_000_000_000"),
			expFixedMinted: sdk.ZeroUint(),
			expError:       false,
		},
		{
			title:          "negative time ahead should result in error",
			normTimePassed: sdk.MustNewDecFromStr("98"),
			timeAhead:      sdk.MustNewDecFromStr("-1.0"),
			totalMinted:    sdk.ZeroUint(),
			expFixedMinted: sdk.ZeroUint(),
			expError:       true,
		},
		{
			title:          "zero time ahead should not mint tokens",
			normTimePassed: sdk.MustNewDecFromStr("85.05385417"),
			timeAhead:      sdk.ZeroDec(),
			totalMinted:    sdk.NewUintFromString("143_483_520_000_000"),
			expFixedMinted: sdk.ZeroUint(),
			expError:       false,
		},
	} {
		t.Run(tc.title, func(t *testing.T) {
			minter := &types.Minter{
				NormTimePassed: tc.normTimePassed,
				TotalMinted:    tc.totalMinted,
			}

			newlyMinted, err := predictMintedByFixedAmount(minter.TotalMinted, minter.NormTimePassed, tc.timeAhead)
			if tc.expError && err == nil {
				t.Error("Error is expected")
			}

			actExpDiff := types.GetAbsDiff(newlyMinted, tc.expFixedMinted)
			if actExpDiff.GT(expAcceptedDeviation) {
				t.Errorf("Minted exp: %v, act: %v, diff: %v", tc.expFixedMinted, newlyMinted, actExpDiff)
			}
		})
	}
}

func randomTimeBetweenBlocks(min uint64, max uint64) uint64 {
	return uint64(time.Second.Nanoseconds()) * (uint64(rand.Int63n(sdk.NewIntFromUint64(max-min).Int64())) + min)
}

func defaultParams() (types.Minter, sdk.Uint, sdk.Uint, sdk.Uint) {
	minter := types.InitialMinter()
	mintedCoins := sdk.NewUint(0)
	mintedMonth := sdk.NewUint(0)
	timeOffset := sdk.NewUint(uint64(util.GetCurrentTimeUnixNano()))
	return minter, mintedCoins, mintedMonth, timeOffset
}

func uintFromDec(d sdk.Dec) sdk.Uint {
	return sdk.NewUint(d.TruncateInt().Uint64())
}
