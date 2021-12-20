package types

func NewGenesisState(suspendState SuspendedState) *GenesisState {
	return &GenesisState{
		State: suspendState,
	}
}

// DefaultGenesisState returns the default Capability genesis state
func DefaultGenesisState() *GenesisState {
	return NewGenesisState(DefaultSuspendedState())
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (m GenesisState) Validate() error {
	return m.State.Validate()
}
