package types

// DefaultIndex is the default capability global index
const DefaultIndex uint64 = 1

func NewGenesis(suspend bool, blockHeight int64) *GenesisState {
	return &GenesisState{
		Suspend: suspend,
		BlockHeight: blockHeight,
	}
}

// DefaultGenesis returns the default Capability genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		// this line is used by starport scaffolding # genesis/types/default
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	return nil
}
