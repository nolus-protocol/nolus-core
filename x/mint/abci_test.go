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
	expectedCoins60Sec      = sdkmath.NewUint(110671242932283)
	expectedNormTime20Sec   = sdkmath.LegacyMustNewDecFromStr("119.999976884641925729")
	normTimeThreshold       = sdkmath.LegacyMustNewDecFromStr("0.0001")
	fiveMinutesInNano       = sdkmath.NewUint(uint64(time.Minute.Nanoseconds() * 5))
	expectedTokensInFormula = []int64{
		831471167387, 829820863761, 828281868588, 826839730229,
		825478208594, 824245045239, 823079423313, 822040420023, 821068654579, 820213875087,
		819440568386, 818769585963, 818188244433, 817710710496, 817307924485, 816997527407,
		816801218448, 816665071819, 816653649327, 816687516174, 816829020601, 817076870515,
		817381227849, 817797546633, 818285112383, 818851440472, 819521147414, 820244244256,
		821076677649, 821989017422, 822969470941, 824035436872, 825179899902, 826414307226,
		827719521455, 829106029879, 830572546616, 832114890622, 833716613001, 835411110262,
		837177165020, 839021213069, 840930930851, 842916649359, 844972836509, 847097499967,
		849298799875, 851576789297, 853903163473, 856331938931, 858791497471, 861331985362,
		863962242560, 866625630338, 869363060252, 872175658740, 875054103746, 877980110097,
		880975946858, 884048218881, 887171089932, 890339954723, 893573962494, 896878411923,
		900234010041, 903659884903, 907126102331, 910673269050, 914266110944, 917897409726,
		921597576949, 925364386158, 929161815914, 933015513595, 936936824067, 940877689459,
		944912156067, 948980488929, 953081003798, 957231607980, 961453590558, 965711597981,
		970004010105, 974358571556, 978754847098, 983204268459, 987669657610, 992198558014,
		996767161256, 1001386360978, 1006046700901, 1010749072213, 1015471626361, 1020258693631,
		1025054584775, 1029909842031, 1034789697625, 1039716743761, 1044692845653, 1049702564826,
		1054707264879, 1059775597373, 1064882345537, 1070005834258, 1075175024704, 1080380322095,
		1085587142549, 1090852470274, 1096136767172, 1101460461185, 1106804686306, 1112155724140,
		1117554744480, 1122993046929, 1128426460549, 1133897580718, 1139405112592, 1144914091087,
		1150478593260, 1156032420164,
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

	mintThreshold := sdkmath.NewUint(20_000_000) // 10 tokens
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

func Test_CalcIncrementDuringFormula_OutputsExpectedIncrementWithinEpsilon(t *testing.T) {
	increment5s := calcFunctionIncrement(sdkmath.NewUint(uint64(time.Second.Nanoseconds() * 5)))
	increment30s := calcFunctionIncrement(sdkmath.NewUint(uint64(time.Second.Nanoseconds() * 30)))
	increment60s := calcFunctionIncrement(sdkmath.NewUint(uint64(time.Second.Nanoseconds() * 60)))

	minutesInPeriod := int64(60) * 24 * 30 * types.MonthsInFormula.TruncateInt64()
	sumIncrements5s := sdkmath.LegacyNewDec(12 * minutesInPeriod).Mul(increment5s)
	sumIncrements30s := sdkmath.LegacyNewDec(2 * minutesInPeriod).Mul(increment30s)
	sumIncrements60s := sdkmath.LegacyNewDec(1 * minutesInPeriod).Mul(increment60s)

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
	increment5s := calcFixedIncrement(sdkmath.NewUint(uint64(time.Second.Nanoseconds() * 5)))
	increment30s := calcFixedIncrement(sdkmath.NewUint(uint64(time.Second.Nanoseconds() * 30)))
	increment60s := calcFixedIncrement(sdkmath.NewUint(uint64(time.Second.Nanoseconds() * 60)))

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
			title:             "start from genesis, 1 month calculated by integral",
			normTimePassed:    sdkmath.LegacyMustNewDecFromStr("0.17"),
			timeAhead:         sdkmath.LegacyMustNewDecFromStr("1"),
			totalMinted:       sdkmath.ZeroUint(),
			expIntegralMinted: sdkmath.NewUintFromString("831_474_371_998"),
			expError:          false,
		},
		{
			title:             "start from genesis, 12 months calculated by integral",
			normTimePassed:    sdkmath.LegacyMustNewDecFromStr("0.17"),
			timeAhead:         twelveMonths,
			totalMinted:       sdkmath.ZeroUint(),
			expIntegralMinted: sdkmath.NewUintFromString("9_890_844_457_754"),
			expError:          false,
		},
		{
			title:             "in the 120 months range, 12 months calculated by integral",
			normTimePassed:    sdkmath.LegacyMustNewDecFromStr("2.04"),
			timeAhead:         twelveMonths,
			totalMinted:       sdkmath.NewUintFromString("141_722_242_835"),
			expIntegralMinted: sdkmath.NewUintFromString("11_280_925_194_119"),
			expError:          false,
		},
		{
			title:             "ends on the 120th month, 12 months calculated by integral",
			normTimePassed:    sdkmath.LegacyMustNewDecFromStr("84.05875000"),
			timeAhead:         twelveMonths,
			totalMinted:       sdkmath.NewUintFromString("1_977_230_000_000"),
			expIntegralMinted: sdkmath.NewUintFromString("82_441_220_298_725"),
			expError:          false,
		},
		{
			title:             "partially in the 120 months range, 1 month calculated by integral",
			normTimePassed:    sdkmath.LegacyMustNewDecFromStr("119.00489583"),
			timeAhead:         twelveMonths,
			totalMinted:       sdkmath.NewUintFromString("100_290_028_000_000"),
			expIntegralMinted: sdkmath.NewUintFromString("10_381_241_757_165"),
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
			totalMinted:       sdkmath.NewUintFromString("73_345_777_103_641"),
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
