package solanacarrier_test

import (
	"encoding/json"
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	types "github.com/Nolus-Protocol/nolus-core/solanacarrier"
)

var update = flag.Bool("update", false, "rewrite the golden files under testdata")

const proposalExamplePath = "testdata/update-params-proposal.json"

type proposalFile struct {
	Messages  []json.RawMessage `json:"messages"`
	Metadata  string            `json:"metadata"`
	Deposit   string            `json:"deposit"`
	Title     string            `json:"title"`
	Summary   string            `json:"summary"`
	Expedited bool              `json:"expedited"`
}

func proposalCodec() *codec.ProtoCodec {
	registry := codectypes.NewInterfaceRegistry()
	types.RegisterInterfaces(registry)
	return codec.NewProtoCodec(registry)
}

func govAuthority(t *testing.T) string {
	t.Helper()
	authority, err := sdk.Bech32ifyAddressBytes("nolus", authtypes.NewModuleAddress(govtypes.ModuleName))
	require.NoError(t, err)
	return authority
}

func TestUpdateParamsProposalExampleMatchesTheDefaults(t *testing.T) {
	cdc := proposalCodec()

	msgJSON, err := cdc.MarshalInterfaceJSON(&types.MsgUpdateParams{
		Authority: govAuthority(t),
		Params:    types.DefaultParams(),
	})
	require.NoError(t, err)

	generated, err := json.MarshalIndent(proposalFile{
		Messages: []json.RawMessage{msgJSON},
		Deposit:  "30000000000unls",
		Title:    "Update the Solana carrier program allowlist",
		Summary: "Sets the programs a SIGN_MODE_SOLANA_TX_CARRIER carrier may invoke besides its bound memo. " +
			"The list replaces the stored one whole.",
	}, "", "  ")
	require.NoError(t, err)
	generated = append(generated, '\n')

	if *update {
		require.NoError(t, os.WriteFile(proposalExamplePath, generated, 0o600))
	}

	committed, err := os.ReadFile(proposalExamplePath)
	require.NoError(t, err)
	require.Equal(t, string(generated), string(committed), "run go test ./x/solanacarrier/ -run ProposalExample -update")
}

func TestUpdateParamsProposalExampleDecodes(t *testing.T) {
	cdc := proposalCodec()

	committed, err := os.ReadFile(proposalExamplePath)
	require.NoError(t, err)
	require.Contains(t, string(committed),
		`"program_id": "BN+teWL/sd2SXQqftebQDOYZW6i7OpH9B++YYMXpe7g="`) // Lighthouse

	var file proposalFile
	require.NoError(t, json.Unmarshal(committed, &file))
	require.Len(t, file.Messages, 1)

	var msg sdk.Msg
	require.NoError(t, cdc.UnmarshalInterfaceJSON(file.Messages[0], &msg))
	require.Equal(t, &types.MsgUpdateParams{
		Authority: govAuthority(t),
		Params:    types.DefaultParams(),
	}, msg)
}
