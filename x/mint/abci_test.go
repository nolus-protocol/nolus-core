package mint

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Nolus-Protocol/nolus-core/custom/util"
	"github.com/Nolus-Protocol/nolus-core/x/mint/types"
)

var (
	expectedCoins60Sec      = sdkmath.NewUint(110812965137065)
	expectedNormTime20Sec   = sdkmath.LegacyMustNewDecFromStr("119.999976851851083852")
	normTimeThreshold       = sdkmath.LegacyMustNewDecFromStr("0.0001")
	fiveMinutesInNano       = sdkmath.NewUint(uint64(time.Minute.Nanoseconds() * 5))
	expectedTokensInFormula = []int64{
		832934816948, 831271811242, 829695356341, 828248939385,
		826872082714, 825599316449, 824443601132, 823355595954, 822380119519, 821514525561,
		820713783758, 820046904203, 819429981640, 818930370019, 818537870306, 818196877181,
		817837243611, 817842608993, 817841973385, 817787896698, 817834100906, 818194410133,
		818486311253, 818880565551, 819357231376, 819936198264, 820587396977, 821307583408,
		822131886932, 823014927908, 824006183536, 825059565539, 826200776447, 827438131216,
		828725945831, 830119135298, 831565516158, 833096089173, 834721194003, 836405097569,
		838161880877, 839996042692, 841900758200, 843890971176, 845947504653, 848071921510,
		850277963103, 852541686330, 854890024849, 857297964698, 859773696598, 862315875568,
		864920472267, 867607206887, 870343235595, 873161991403, 876027269974, 878960647865,
		881968669856, 885027566950, 888160516719, 891350504585, 894584724825, 897887581250,
		901252264319, 904697728175, 908160810698, 911698015060, 915289477634, 918933072653,
		922661242632, 926417929254, 930215343691, 934077877155, 937995994556, 941979277093,
		945984492200, 950072914452, 954194638263, 958346483768, 962582902635, 966854894285,
		971148054036, 975513678053, 979916708309, 984371169138, 988869625055, 993411783806,
		997982311017, 1002627953125, 1007265684343, 1011977261615, 1016715058280, 1021503025328,
		1026335532432, 1031204039333, 1036112239206, 1041027695773, 1046012680041, 1051027734026,
		1056080248391, 1061146619517, 1066273204031, 1071412196425, 1076581263984, 1081789735633,
		1087038354691, 1092325785121, 1097597611815, 1102931622903, 1108289222364, 1113680416472,
		1119099458450, 1124526125020, 1129982107999, 1135473837965, 1140980117098, 1146516049672,
		1152083758164, 1157657672766,
	}
)

func TestTimeDifference(t *testing.T) {
	_, _, _, timeOffset := defaultParams()
	tb := sdkmath.NewUint(uint64(time.Second.Nanoseconds() * 60)) // 60 seconds
	td := calcTimeDifference(timeOffset.Add(tb), timeOffset, fiveMinutesInNano)

	require.Equal(t, td, tb)
}

func TestTimeDifference_MoreThenMax(t *testing.T) {
	_, _, _, timeOffset := defaultParams()
	tb := fiveMinutesInNano.Add(sdkmath.NewUint(1))
	td := calcTimeDifference(timeOffset.Add(tb), timeOffset, fiveMinutesInNano)

	require.Equal(t, td, fiveMinutesInNano)
}

func TestTimeDifference_InvalidTime(t *testing.T) {
	_, _, _, timeOffset := defaultParams()
	require.Panics(t, assert.PanicTestFunc(func() {
		calcTimeDifference(timeOffset, timeOffset.Add(sdkmath.NewUint(1)), fiveMinutesInNano)
	}))
}

func Test_CalcTokensDuringFormula_WhenUsingConstantIncrements_OutputsPredeterminedAmount(t *testing.T) {
	timeBetweenBlocks := sdkmath.NewUint(uint64(time.Second.Nanoseconds() * 60)) // 60 seconds per block
	minutesInMonth := uint64(time.Hour.Minutes()) * 24 * 30
	minutesInFormula := minutesInMonth * uint64(types.MonthsInFormula.TruncateInt64())
	minter, mintedCoins, mintedMonth, timeOffset := defaultParams()

	for i := uint64(0); i < minutesInFormula; i++ {
		coins := calcTokens(timeOffset.Add(sdkmath.NewUint(i).Mul(timeBetweenBlocks)), &minter, fiveMinutesInNano)

		mintedCoins = mintedCoins.Add(sdkmath.NewUint(coins.Uint64()))
		mintedMonth = mintedMonth.Add(sdkmath.NewUint(coins.Uint64()))

		if i%minutesInMonth == 0 {
			fmt.Printf("%v Month, %v Minted, %v Total Minted(in store), %v Returned Total, %v Norm Time, %v Received in this block \n",
				i/minutesInMonth, mintedMonth, minter.TotalMinted, mintedCoins, minter.NormTimePassed, coins)
			mintedMonth = sdkmath.ZeroUint()
		}
	}

	fmt.Printf("%v Returned Total, %v Total Minted(in store), %v Norm Time \n",
		mintedCoins, minter.TotalMinted, minter.NormTimePassed)

	if !expectedCoins60Sec.Equal(mintedCoins) || !expectedCoins60Sec.Equal(sdkmath.NewUint(minter.TotalMinted.Uint64())) {
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
	r := rand.New(rand.NewSource(util.GetCurrentTimeUnixNano()))
	monthThreshold := sdkmath.NewUint(187_500_000) // 187.5 tokens
	month := 0

	for timeOffset.LT(sdkmath.NewUint(uint64(nanoSecondsInPeriod))) {
		i := sdkmath.NewUint(randomTimeBetweenBlocks(5, 60, r))

		coins := calcTokens(timeOffset.Add(i), &minter, fiveMinutesInNano)
		if coins.LT(sdkmath.ZeroUint()) {
			t.Errorf("Minted negative %v coins", coins)
		}

		mintedCoins = mintedCoins.Add(sdkmath.NewUint(coins.Uint64()))
		mintedMonth = mintedMonth.Add(sdkmath.NewUint(coins.Uint64()))

		prevI := timeOffset.Sub(prevOffset)
		nanoSecondsInMonthUint := sdkmath.NewUint(uint64(nanoSecondsInMonth.TruncateInt64()))
		// divide (nanoseconds of time passed) by (nanoseconds in a month) to get how many months have passed.
		prevIMonths := prevI.Quo(nanoSecondsInMonthUint)
		// divide (nanoseconds of time passed + random time between blocks) by (nanoseconds in a month) to get how many months have passed.
		prevIPlusRandomTime := prevI.Add(i).Quo(nanoSecondsInMonthUint)

		if !prevIMonths.Equal(prevIPlusRandomTime) {
			month++

			fmt.Printf("%v Month, %v Minted, %v Total Minted(in store), %v Returned Total, %v Norm Time, %v Received in this block \n",
				month, mintedMonth, minter.TotalMinted, mintedCoins, minter.NormTimePassed, coins)

			if types.GetAbsDiff(mintedMonth, sdkmath.NewUint(uint64(expectedTokensInFormula[month-1]))).GT(monthThreshold) {
				t.Errorf("Minted unexpected amount of tokens for month %d, expected [%v +/- %v], actual %v",
					month, expectedTokensInFormula[month-1], monthThreshold, mintedMonth)
			}

			prevOffset = timeOffset
			mintedMonth = sdkmath.ZeroUint()
			r = rand.New(rand.NewSource(util.GetCurrentTimeUnixNano()))
		}

		timeOffset = timeOffset.Add(i)
	}

	mintThreshold := sdkmath.NewUint(20_000_000) // 20 tokens
	fmt.Printf("%v Returned Total, %v Total Minted(in store), %v Norm Time \n", mintedCoins, minter.TotalMinted, minter.NormTimePassed)

	if types.GetAbsDiff(expectedCoins60Sec, mintedCoins).GT(mintThreshold) || types.GetAbsDiff(expectedCoins60Sec, sdkmath.Uint(minter.TotalMinted)).GT(mintThreshold) {
		t.Errorf("Minted unexpected amount of tokens, expected [%v +/- %v] returned and in store, actual minted %v, actual in store %v",
			expectedCoins60Sec, mintThreshold, mintedCoins, minter.TotalMinted)
	}

	if expectedNormTime20Sec.Sub(minter.NormTimePassed).Abs().GT(normTimeThreshold) {
		t.Errorf("Received unexpected normalized time, expected [%v +/- %v], actual %v",
			expectedNormTime20Sec, normTimeThreshold, minter.NormTimePassed)
	}
}

func Test_CalcTokens_WhenMintingAllTokens_OutputsExactExpectedTokens(t *testing.T) {
	minter, mintedCoins, mintedMonth, timeOffset := defaultParams()
	prevOffset := timeOffset
	offsetNanoInPeriod := uintFromDec((nanoSecondsInMonth.Mul(sdkmath.LegacyNewDec(121))).Add(types.DecFromUint(timeOffset))) // Adding 1 extra to ensure cap is preserved
	month := 0
	r := rand.New(rand.NewSource(util.GetCurrentTimeUnixNano()))

	for timeOffset.LT(offsetNanoInPeriod) {
		i := sdkmath.NewUint(randomTimeBetweenBlocks(60, 120, r))

		coins := calcTokens(timeOffset.Add(i), &minter, fiveMinutesInNano)
		mintedCoins = mintedCoins.Add(sdkmath.NewUint(coins.Uint64()))
		mintedMonth = mintedMonth.Add(sdkmath.NewUint(coins.Uint64()))

		prevI := timeOffset.Sub(prevOffset)
		nanoSecondsInMonthUint := sdkmath.NewUint(uint64(nanoSecondsInMonth.TruncateInt64()))
		// divide (nanoseconds of time passed) by (nanoseconds in a month) to get how many months have passed.
		prevIMonths := prevI.Quo(nanoSecondsInMonthUint)
		// divide (nanoseconds of time passed + random time between blocks) by (nanoseconds in a month) to get how many months have passed.
		prevIPlusRandomTime := prevI.Add(i).Quo(nanoSecondsInMonthUint)

		if !prevIMonths.Equal(prevIPlusRandomTime) {
			month++

			r = rand.New(rand.NewSource(util.GetCurrentTimeUnixNano()))
			fmt.Printf("%v Month, %v Minted, %v Total Minted(in store), %v Returned Total, %v Norm Time, %v Received in this block \n",
				month, mintedMonth, minter.TotalMinted, mintedCoins, minter.NormTimePassed, coins)
			prevOffset = timeOffset
			mintedMonth = sdkmath.ZeroUint()
		}

		timeOffset = timeOffset.Add(i)
	}

	fmt.Printf("%v Returned Total, %v Total Minted(in store), %v Norm Time \n",
		mintedCoins, minter.TotalMinted, minter.NormTimePassed)

	// require.Equal(t, types.MintingCap, minter.TotalMinted)
	require.EqualValues(t, minter.TotalMinted, mintedCoins)
}

func Test_CalcTokens_WhenGivenBlockWithDiffBiggerThanMax_MaxMintedTokensAreCreated(t *testing.T) {
	timeOffset := time.Now()
	timeOffsetUint := sdkmath.NewUint(uint64(timeOffset.UnixNano()))
	nextOffset := sdkmath.NewUint(uint64(timeOffset.Add(time.Hour).UnixNano()))

	originalMinter := types.InitialMinter()
	originalMinter.PrevBlockTimestamp = sdkmath.NewUint(uint64(timeOffset.UnixNano()))

	minter := types.InitialMinter()
	minter.PrevBlockTimestamp = sdkmath.NewUint(uint64(timeOffset.UnixNano()))

	coins := calcTokens(nextOffset, &minter, fiveMinutesInNano)
	expectedCoins := calcTokens(timeOffsetUint.Add(fiveMinutesInNano), &originalMinter, fiveMinutesInNano)

	require.Equal(t, expectedCoins, coins)
}

// func Test_CalcIncrementDuringFormula_OutputsExpectedIncrementWithinEpsilon(t *testing.T) {
// 	increment5s := calcFractionOfMonth(sdkmath.NewUint(uint64(time.Second.Nanoseconds() * 5)))
// 	increment30s := calcFractionOfMonth(sdkmath.NewUint(uint64(time.Second.Nanoseconds() * 30)))
// 	increment60s := calcFractionOfMonth(sdkmath.NewUint(uint64(time.Second.Nanoseconds() * 60)))

// 	minutesInPeriod := int64(60) * 24 * 30 * types.MonthsInFormula.TruncateInt64()
// 	sumIncrements5s := sdkmath.LegacyNewDec(12 * minutesInPeriod).Mul(increment5s)
// 	sumIncrements30s := sdkmath.LegacyNewDec(2 * minutesInPeriod).Mul(increment30s)
// 	sumIncrements60s := sdkmath.LegacyNewDec(1 * minutesInPeriod).Mul(increment60s)

// 	if sumIncrements5s.Sub(types.AbsMonthsRange).Abs().GT(normTimeThreshold) {
// 		t.Errorf("Increment with 5 second step results in range %v, deviating with more than epsilon from expected %v",
// 			sumIncrements5s, types.AbsMonthsRange)
// 	}

// 	if sumIncrements30s.Sub(types.AbsMonthsRange).Abs().GT(normTimeThreshold) {
// 		t.Errorf("Increment with 30 second step results in range %v, deviating with more than epsilon from expected %v",
// 			sumIncrements30s, types.AbsMonthsRange)
// 	}

// 	if sumIncrements60s.Sub(types.AbsMonthsRange).Abs().GT(normTimeThreshold) {
// 		t.Errorf("Increment with 60 second step results in range %v, deviating with more than epsilon from expected %v",
// 			sumIncrements60s, types.AbsMonthsRange)
// 	}
// }

func Test_CalcFixedIncrement_OutputsExpectedIncrementWithinEpsilon(t *testing.T) {
	increment5s := calcFractionOfMonth(sdkmath.NewUint(uint64(time.Second.Nanoseconds() * 5)))
	increment30s := calcFractionOfMonth(sdkmath.NewUint(uint64(time.Second.Nanoseconds() * 30)))
	increment60s := calcFractionOfMonth(sdkmath.NewUint(uint64(time.Second.Nanoseconds() * 60)))

	minutesInMonth := int64(time.Hour.Minutes()) * 24 * 30
	sumIncrements5s := sdkmath.LegacyNewDec(12 * minutesInMonth).Mul(increment5s)
	sumIncrements30s := sdkmath.LegacyNewDec(2 * minutesInMonth).Mul(increment30s)
	sumIncrements60s := sdkmath.LegacyNewDec(1 * minutesInMonth).Mul(increment60s)

	if sumIncrements5s.Sub(sdkmath.LegacyOneDec()).Abs().GT(normTimeThreshold) {
		t.Errorf("Increment with 5 second step results in range %v, deviating with more than epsilon from expected %v",
			sumIncrements5s, sdkmath.LegacyOneDec())
	}

	if sumIncrements30s.Sub(sdkmath.LegacyOneDec()).Abs().GT(normTimeThreshold) {
		t.Errorf("Increment with 30 second step results in range %v, deviating with more than epsilon from expected %v",
			sumIncrements30s, sdkmath.LegacyOneDec())
	}

	if sumIncrements60s.Sub(sdkmath.LegacyOneDec()).Abs().GT(normTimeThreshold) {
		t.Errorf("Increment with 60 second step results in range %v, deviating with more than epsilon from expected %v",
			sumIncrements60s, sdkmath.LegacyOneDec())
	}
}

func Test_PredictMintedByIntegral_TwelveMonthsAhead(t *testing.T) {
	expAcceptedDeviation := sdkmath.NewUint(500_000) // 0.5 token

	for _, tc := range []struct {
		title             string
		normTimePassed    sdkmath.LegacyDec
		timeAhead         sdkmath.LegacyDec
		totalMinted       sdkmath.Uint
		expIntegralMinted sdkmath.Uint
		expError          bool
	}{
		{
			title:             "start from 17th month when formula is applied for the first time, 1 month calculated by integral",
			normTimePassed:    sdkmath.LegacyMustNewDecFromStr("17"),
			timeAhead:         sdkmath.LegacyMustNewDecFromStr("1"),
			totalMinted:       sdkmath.NewUintFromString("14_218_115_061_000"),
			expIntegralMinted: sdkmath.NewUintFromString("600_000_000_000"),
			expError:          false,
		},
		{
			title:             "start from 17th month when formula is applied for the first time, 12 months calculated by integral",
			normTimePassed:    sdkmath.LegacyMustNewDecFromStr("17"),
			timeAhead:         twelveMonths,
			totalMinted:       sdkmath.NewUintFromString("14_218_115_061_000"),
			expIntegralMinted: sdkmath.NewUintFromString("9_612_567_769_562"),
			expError:          false,
		},
		{
			title:             "in the 120 months range, 12 months calculated by integral",
			normTimePassed:    sdkmath.LegacyMustNewDecFromStr("19"),
			timeAhead:         twelveMonths,
			totalMinted:       sdkmath.NewUintFromString("16_218_115_061_000"),
			expIntegralMinted: sdkmath.NewUintFromString("9_259_614_547_562"),
			expError:          false,
		},
		{
			title:             "ends on the 120th month, 12 months calculated by integral",
			normTimePassed:    sdkmath.LegacyMustNewDecFromStr("108.05875000"),
			timeAhead:         twelveMonths,
			totalMinted:       sdkmath.NewUintFromString("100_218_115_061_000"),
			expIntegralMinted: sdkmath.NewUintFromString("10_594_876_939_000"),
			expError:          false,
		},
		{
			title:             "partially in the 120 months range, 1 month calculated by integral",
			normTimePassed:    sdkmath.LegacyMustNewDecFromStr("119.00489583"),
			timeAhead:         twelveMonths,
			totalMinted:       sdkmath.NewUintFromString("100_290_028_000_000"),
			expIntegralMinted: sdkmath.NewUintFromString("10_522_964_000_000"),
			expError:          false,
		},
		{
			title:             "after 120th month, 0 months calculated by integral",
			normTimePassed:    sdkmath.LegacyMustNewDecFromStr("125"),
			timeAhead:         twelveMonths,
			totalMinted:       sdkmath.NewUintFromString("147_741_507_000_000"),
			expIntegralMinted: sdkmath.ZeroUint(),
			expError:          false,
		},
		{
			title:             "negative time ahead should result in error",
			normTimePassed:    sdkmath.LegacyMustNewDecFromStr("98"),
			timeAhead:         sdkmath.LegacyMustNewDecFromStr("-1.0"),
			totalMinted:       sdkmath.ZeroUint(),
			expIntegralMinted: sdkmath.ZeroUint(),
			expError:          true,
		},
		{
			title:             "zero time ahead should not mint tokens",
			normTimePassed:    sdkmath.LegacyMustNewDecFromStr("85.05385417"),
			timeAhead:         sdkmath.LegacyZeroDec(),
			totalMinted:       sdkmath.NewUintFromString("73_487_499_360_806"),
			expIntegralMinted: sdkmath.ZeroUint(),
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

func randomTimeBetweenBlocks(min uint64, max uint64, r *rand.Rand) uint64 {
	return uint64(time.Second.Nanoseconds()) * (uint64(r.Int63n(sdkmath.NewIntFromUint64(max-min).Int64())) + min)
}

func defaultParams() (types.Minter, sdkmath.Uint, sdkmath.Uint, sdkmath.Uint) {
	minter := types.InitialMinter()
	mintedCoins := sdkmath.NewUint(0)
	mintedMonth := sdkmath.NewUint(0)
	timeOffset := sdkmath.NewUint(uint64(util.GetCurrentTimeUnixNano()))
	return minter, mintedCoins, mintedMonth, timeOffset
}

func uintFromDec(d sdkmath.LegacyDec) sdkmath.Uint {
	return sdkmath.NewUint(d.TruncateInt().Uint64())
}
