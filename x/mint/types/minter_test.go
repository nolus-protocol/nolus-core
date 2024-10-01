package types

import (
	"testing"

	sdkmath "cosmossdk.io/math"
)

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
		{
			title:          "norm time passed bigger then the minting schedule cap should return error",
			normTimePassed: TotalMonths.Add(sdkmath.LegacyMustNewDecFromStr("0.1")),
			totalMinted:    DefaultInitialMinter().TotalMinted,
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
