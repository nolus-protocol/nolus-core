package v090

import (
	"bytes"
	"maps"
	"slices"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func testAddr(b byte) sdk.AccAddress {
	return bytes.Repeat([]byte{b}, 20)
}

func bech32(t *testing.T, prefix string, addr []byte) string {
	t.Helper()
	s, err := sdk.Bech32ifyAddressBytes(prefix, addr)
	require.NoError(t, err, "test address encodes to bech32")
	return s
}

func file(content string) *fstest.MapFile {
	return &fstest.MapFile{Data: []byte(content)}
}

func mustConfig(t *testing.T, recipient sdk.AccAddress, compromised []sdk.AccAddress) ChainConfig {
	t.Helper()
	cfg, err := NewChainConfig(recipient, compromised)
	require.NoError(t, err, "test config is valid")
	return cfg
}

func TestParseListsAcceptsValidLists(t *testing.T) {
	recipient, a, b := testAddr(1), testAddr(2), testAddr(3)
	fsys := fstest.MapFS{
		"chain-a/recipient.txt":   file("# treasury\n\n  " + bech32(t, "nolus", recipient) + "  \n"),
		"chain-a/compromised.txt": file(bech32(t, "nolus", a) + "\n# second\n" + bech32(t, "nolus", b) + "\n"),
		"chain-b/recipient.txt":   file(bech32(t, "nolus", recipient)),
	}

	got, err := parseLists(fsys)
	require.NoError(t, err, "valid lists must parse")
	require.Equal(t, map[string]ChainConfig{
		"chain-a": mustConfig(t, recipient, []sdk.AccAddress{a, b}),
		"chain-b": mustConfig(t, recipient, nil),
	}, got, "every chain directory parses to its config")
}

func TestParseListsRejectsInvalidLists(t *testing.T) {
	recipient, a := testAddr(1), testAddr(2)
	valid := bech32(t, "nolus", recipient)
	other := bech32(t, "nolus", a)

	tests := []struct {
		name    string
		fsys    fstest.MapFS
		wantErr error
	}{
		{
			name:    "no recipient file",
			fsys:    fstest.MapFS{"c/compromised.txt": file(other)},
			wantErr: ErrMissingRecipient,
		},
		{
			name:    "recipient file with only comments",
			fsys:    fstest.MapFS{"c/recipient.txt": file("# none\n\n")},
			wantErr: ErrMissingRecipient,
		},
		{
			name:    "invalid bech32 recipient",
			fsys:    fstest.MapFS{"c/recipient.txt": file("nolus1notanaddress")},
			wantErr: ErrInvalidAddress,
		},
		{
			name:    "two recipients",
			fsys:    fstest.MapFS{"c/recipient.txt": file(valid + "\n" + other)},
			wantErr: ErrMultipleRecipients,
		},
		{
			name:    "invalid bech32",
			fsys:    fstest.MapFS{"c/recipient.txt": file(valid), "c/compromised.txt": file("nolus1notanaddress")},
			wantErr: ErrInvalidAddress,
		},
		{
			name:    "wrong prefix",
			fsys:    fstest.MapFS{"c/recipient.txt": file(valid), "c/compromised.txt": file(bech32(t, "cosmos", a))},
			wantErr: ErrInvalidAddress,
		},
		{
			name:    "32-byte address",
			fsys:    fstest.MapFS{"c/recipient.txt": file(valid), "c/compromised.txt": file(bech32(t, "nolus", bytes.Repeat([]byte{4}, 32)))},
			wantErr: ErrInvalidAddress,
		},
		{
			name:    "32-byte recipient",
			fsys:    fstest.MapFS{"c/recipient.txt": file(bech32(t, "nolus", bytes.Repeat([]byte{4}, 32)))},
			wantErr: ErrInvalidAddress,
		},
		{
			name:    "duplicate compromised address",
			fsys:    fstest.MapFS{"c/recipient.txt": file(valid), "c/compromised.txt": file(other + "\n" + other)},
			wantErr: ErrDuplicateAddress,
		},
		{
			name:    "recipient among compromised",
			fsys:    fstest.MapFS{"c/recipient.txt": file(valid), "c/compromised.txt": file(valid)},
			wantErr: ErrRecipientListed,
		},
		{
			name:    "unknown file in a chain directory",
			fsys:    fstest.MapFS{"c/recipient.txt": file(valid), "c/notes.txt": file("")},
			wantErr: ErrUnknownListFile,
		},
		{
			name:    "directory in a chain directory",
			fsys:    fstest.MapFS{"c/recipient.txt": file(valid), "c/extra/recipient.txt": file(valid)},
			wantErr: ErrUnknownListFile,
		},
		{
			name:    "list file name used for a directory",
			fsys:    fstest.MapFS{"c/recipient.txt/x": file(valid)},
			wantErr: ErrUnknownListFile,
		},
		{
			name:    "file outside a chain directory",
			fsys:    fstest.MapFS{"c/recipient.txt": file(valid), "README": file("")},
			wantErr: ErrUnknownListFile,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := parseLists(tc.fsys)
			require.ErrorIs(t, err, tc.wantErr, "invalid lists must be refused with the matching error")
		})
	}
}

func TestCommittedListsParse(t *testing.T) {
	recipient, err := sdk.GetFromBech32("nolus1luvaejqplajhq33xpn42ffsy6558fzg4dvkqnz", "nolus")
	require.NoError(t, err, "the recovery recipient is a nolus address")

	first, err := sdk.GetFromBech32("nolus1p52cmyqtzggu6hxaajh3vrntcugtkrk28fvxzm", "nolus")
	require.NoError(t, err, "the first pirin-1 victim is a nolus address")
	last, err := sdk.GetFromBech32("nolus1qpxdzl5hcu8qjk6g5yfyr2mr7079zlsxyusf7j", "nolus")
	require.NoError(t, err, "the last pirin-1 attacker is a nolus address")

	got := mustParseEmbedded()
	require.Equal(t, []string{"pirin-1", "rila-3"}, slices.Sorted(maps.Keys(got)), "only pirin-1 and rila-3 carry lists")
	require.Equal(t, mustConfig(t, recipient, nil), got["rila-3"], "rila-3 carries only the recovery recipient")

	pirin := got["pirin-1"]
	// the 97 victims and 2 attackers are data; the list is checked by size, order at both ends and recipient
	require.Equal(t, sdk.AccAddress(recipient), pirin.recipient, "pirin-1 sweeps to the recovery recipient")
	require.Len(t, pirin.compromised, 99, "pirin-1 lists 97 victims and 2 attackers")
	require.Equal(t, []sdk.AccAddress{first, last}, []sdk.AccAddress{pirin.compromised[0], pirin.compromised[98]},
		"pirin-1 keeps the file order")
}
