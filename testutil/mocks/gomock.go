package mocks

//go:generate mockgen -source=./../../x/tax/types/expected_keepers.go -destination ./tax/types/expected_keepers.go
//go:generate mockgen -source=./../../x/contractmanager/types/expected_keepers.go -destination ./contractmanager/types/expected_keepers.go
//go:generate mockgen -source=./../../x/feerefunder/types/expected_keepers.go -destination ./feerefunder/types/keepers.go
//go:generate mockgen -source=./../../x/interchaintxs/types/expected_keepers.go -destination ./interchaintxs/types/expected_keepers.go
//go:generate mockgen -source=./../../x/transfer/types/expected_keepers.go -destination ./transfer/types/expected_keepers.go
