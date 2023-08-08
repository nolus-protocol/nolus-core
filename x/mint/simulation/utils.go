package simulation

// DONTCOVER

import (
	"math/rand"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
)

// refactor: decide if we want to use this in simulations
// RandomMaxMintableNanoSeconds generates a random maximum mintable nano seconds in the range of [lowerRange, upperRange]
func RandomMaxMintableNanoSeconds(r *rand.Rand, lowerRange, upperRange int) sdktypes.Uint {
	randomMaxMintableNanoSeconds := r.Intn(upperRange) + lowerRange
	return sdktypes.NewUint(uint64(randomMaxMintableNanoSeconds))
}
