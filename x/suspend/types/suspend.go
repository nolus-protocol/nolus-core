package types

import "fmt"

func NewSuspendedState(adminAddress string, suspended bool, blockHeight int64) SuspendedState {
	return SuspendedState{
		AdminAddress: adminAddress,
		Suspended:    suspended,
		BlockHeight:  blockHeight,
	}
}

func DefaultSuspendedState() SuspendedState {
	return NewSuspendedState("", false, 0)
}

func (m SuspendedState) Validate() error {
	if m.BlockHeight < 0 {
		return fmt.Errorf("suspend block height must be positive %d", m.BlockHeight)

	}
	return nil
}
