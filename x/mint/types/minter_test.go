package types

import (
	"testing"

	sdkmath "cosmossdk.io/math"
)

func Test_calcMintedTokens(t *testing.T) {
	expAcceptedDeviation := sdkmath.NewUint(500_000) // 0.5 token

	for _, tc := range []struct {
		title          string
		normTimePassed sdkmath.LegacyDec
		expTotalMinted sdkmath.Uint
	}{
		{
			title:          "starting at genesis",
			normTimePassed: DefaultInitialMinter().NormTimePassed,
			expTotalMinted: DefaultInitialMinter().TotalMinted,
		},
		{
			title:          "starting at the end of 1st month",
			normTimePassed: sdkmath.LegacyMustNewDecFromStr("1.46510417"),
			expTotalMinted: sdkmath.NewUintFromString("3_760_114_000_000"),
		},
		{
			title:          "starting at the end of 2nd month",
			normTimePassed: sdkmath.LegacyMustNewDecFromStr("2.46020833"),
			expTotalMinted: sdkmath.NewUintFromString("7_435_238_000_000"),
		},
		{
			title:          "starting at the end of 96th month",
			normTimePassed: sdkmath.LegacyMustNewDecFromStr("96.00000000"),
			expTotalMinted: sdkmath.NewUintFromString("147_535_257_000_000"),
		},
		{
			title:          "starting at the end of 97th month",
			normTimePassed: sdkmath.LegacyMustNewDecFromStr("97.00000000"),
			expTotalMinted: sdkmath.NewUintFromString("147_638_382_000_000"),
		},
	} {
		t.Run(tc.title, func(t *testing.T) {
			minter := Minter{
				NormTimePassed: tc.normTimePassed,
				TotalMinted:    tc.expTotalMinted,
			}

			totalMinted := calcMintedTokens(minter)
			actExpDiff := GetAbsDiff(totalMinted, tc.expTotalMinted)

			if actExpDiff.GT(expAcceptedDeviation) {
				t.Errorf("Minted exp: %v, act: %v, diff: %v", tc.expTotalMinted, totalMinted, actExpDiff)
			}
		})
	}
}

func Test_ValidateMinter(t *testing.T) {
	for _, tc := range []struct {
		title          string
		normTimePassed sdkmath.LegacyDec
		totalMinted    sdkmath.Uint
		expErr         bool
	}{
		{
			title:          "default minter should be valid",
			normTimePassed: DefaultInitialMinter().NormTimePassed,
			totalMinted:    DefaultInitialMinter().TotalMinted,
			expErr:         false,
		},
		{
			title:          "negative norm time passed should return error",
			normTimePassed: sdkmath.LegacyMustNewDecFromStr("-0.1"),
			totalMinted:    DefaultInitialMinter().TotalMinted,
			expErr:         true,
		},
		// {
		// 	title:          "norm time passed bigger then the minting schedule cap should return error",
		// 	normTimePassed: TotalMonths.Add(sdkmath.LegacyMustNewDecFromStr("0.1")),
		// 	totalMinted:    DefaultInitialMinter().TotalMinted,
		// 	expErr:         true,
		// },
		// {
		// 	title:          "total minted bigger then minting cap should return error",
		// 	normTimePassed: DefaultInitialMinter().NormTimePassed,
		// 	totalMinted:    MintingCap.Add(sdkmath.NewUint(1)),
		// 	expErr:         true,
		// },
		{
			title:          "total minted not fitting the minting schedule should return error",
			normTimePassed: sdkmath.LegacyMustNewDecFromStr("2.46020833"),
			totalMinted:    sdkmath.NewUintFromString("7_435_237_908_858").Add(sdkmath.NewUint(1)),
			expErr:         true,
		},
	} {
		t.Run(tc.title, func(t *testing.T) {
			minter := Minter{
				NormTimePassed: tc.normTimePassed,
				TotalMinted:    tc.totalMinted,
			}

			err := ValidateMinter(minter)
			if tc.expErr && err == nil {
				t.Errorf("Error expected but got nil")
			}

			if !tc.expErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}
