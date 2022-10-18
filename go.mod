module gitlab-nomo.credissimo.net/nomo/cosmzone

go 1.16

require (
	github.com/CosmWasm/wasmd v0.27.0
	github.com/cosmos/cosmos-sdk v0.45.9
	github.com/cosmos/ibc-go/v3 v3.0.0
	github.com/gogo/protobuf v1.3.3
	github.com/golang/protobuf v1.5.2
	github.com/gorilla/mux v1.8.0
	github.com/grpc-ecosystem/grpc-gateway v1.16.0
	github.com/neutron-org/neutron v0.0.0-20220909075558-caf377b38db1
	github.com/spf13/cast v1.5.0
	github.com/spf13/cobra v1.5.0
	github.com/stretchr/testify v1.8.0
	github.com/tendermint/spm v0.1.9
	github.com/tendermint/tendermint v0.34.21
	github.com/tendermint/tm-db v0.6.7
	google.golang.org/genproto v0.0.0-20220822174746-9e6da59bd2fc
	google.golang.org/grpc v1.48.0
	gopkg.in/yaml.v2 v2.4.0
)

replace (

	github.com/99designs/keyring => github.com/cosmos/keyring v1.1.7-0.20210622111912-ef00f8ac3d76
	// For more info https://github.com/cosmos/cosmos-sdk/releases/tag/v0.45.9
	github.com/confio/ics23/go => github.com/cosmos/cosmos-sdk/ics23/go v0.8.0
	github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1
	google.golang.org/grpc => google.golang.org/grpc v1.33.2
)
