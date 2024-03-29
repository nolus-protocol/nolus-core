package types

var (
	// MinterKey is the key to use for the keeper store.
	MinterKey = []byte{0x00}
	ParamsKey = []byte{0x01}
)

const (
	// module name.
	ModuleName = "mint"

	// StoreKey is the default store key for mint.
	StoreKey = ModuleName

	// QuerierRoute is the querier route for the minting store.
	QuerierRoute = StoreKey

	// Query endpoints supported by the minting querier.
	QueryParameters = "parameters"
	QueryMintState  = "mintState"
	QueryInflation  = "inflation"
)
