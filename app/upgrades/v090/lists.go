// Package v090 is the v0.9.0 upgrade: module migrations, then a recovery sweep configured per chain in
// lists/<chain-id>/, holding recipient.txt (required) and compromised.txt (optional), one bech32 address
// per line with blank and '#' lines skipped. Any other file or an invalid entry panics at start.
package v090

import (
	"embed"
	"fmt"
	"io/fs"
	"path"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec/address"
	sdk "github.com/cosmos/cosmos-sdk/types"

	appparams "github.com/Nolus-Protocol/nolus-core/app/params"
)

const (
	recipientFile   = "recipient.txt"
	compromisedFile = "compromised.txt"
)

//go:embed lists
var embedded embed.FS

// ChainConfig is one chain's validated recovery sweep configuration.
type ChainConfig struct {
	recipient   sdk.AccAddress
	compromised []sdk.AccAddress
}

// NewChainConfig returns a config whose addresses are 20 bytes and whose compromised addresses are unique
// and not the recipient.
func NewChainConfig(recipient sdk.AccAddress, compromised []sdk.AccAddress) (ChainConfig, error) {
	if err := checkAddressLength(recipient); err != nil {
		return ChainConfig{}, fmt.Errorf("recipient: %w", err)
	}
	if err := checkAddressList(recipient, compromised); err != nil {
		return ChainConfig{}, fmt.Errorf("compromised: %w", err)
	}
	return ChainConfig{recipient: recipient, compromised: compromised}, nil
}

func checkAddressLength(addr sdk.AccAddress) error {
	if len(addr) != 20 {
		return fmt.Errorf("%w %X: %d bytes, want 20", ErrInvalidAddress, addr.Bytes(), len(addr))
	}
	return nil
}

func checkAddressList(recipient sdk.AccAddress, addrs []sdk.AccAddress) error {
	seen := make(map[string]struct{}, len(addrs))
	for _, addr := range addrs {
		if err := checkAddressLength(addr); err != nil {
			return err
		}
		if addr.Equals(recipient) {
			return fmt.Errorf("%w: %s", ErrRecipientListed, addr)
		}
		if _, dup := seen[string(addr)]; dup {
			return fmt.Errorf("%w: %s", ErrDuplicateAddress, addr)
		}
		seen[string(addr)] = struct{}{}
	}
	return nil
}

func parseLists(fsys fs.FS) (map[string]ChainConfig, error) {
	chains, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return nil, fmt.Errorf("read chain directories: %w", err)
	}
	configs := make(map[string]ChainConfig, len(chains))
	for _, chain := range chains {
		if !chain.IsDir() {
			return nil, fmt.Errorf("%w: %s", ErrUnknownListFile, chain.Name())
		}
		cfg, err := parseChain(fsys, chain.Name())
		if err != nil {
			return nil, fmt.Errorf("chain %s: %w", chain.Name(), err)
		}
		configs[chain.Name()] = cfg
	}
	return configs, nil
}

func parseChain(fsys fs.FS, dir string) (ChainConfig, error) {
	files, err := fs.ReadDir(fsys, dir)
	if err != nil {
		return ChainConfig{}, fmt.Errorf("read directory: %w", err)
	}
	lists := make(map[string][]string, len(files))
	for _, file := range files {
		switch file.Name() {
		case recipientFile, compromisedFile:
			if file.IsDir() {
				return ChainConfig{}, fmt.Errorf("%w: %s", ErrUnknownListFile, file.Name())
			}
		default:
			return ChainConfig{}, fmt.Errorf("%w: %s", ErrUnknownListFile, file.Name())
		}
		entries, err := readEntries(fsys, path.Join(dir, file.Name()))
		if err != nil {
			return ChainConfig{}, err
		}
		lists[file.Name()] = entries
	}

	switch n := len(lists[recipientFile]); {
	case n == 0:
		return ChainConfig{}, ErrMissingRecipient
	case n > 1:
		return ChainConfig{}, fmt.Errorf("%w: %d entries", ErrMultipleRecipients, n)
	}
	recipient, err := parseAddresses(lists[recipientFile])
	if err != nil {
		return ChainConfig{}, fmt.Errorf("%s: %w", recipientFile, err)
	}
	compromised, err := parseAddresses(lists[compromisedFile])
	if err != nil {
		return ChainConfig{}, fmt.Errorf("%s: %w", compromisedFile, err)
	}
	return NewChainConfig(recipient[0], compromised)
}

func readEntries(fsys fs.FS, name string) ([]string, error) {
	content, err := fs.ReadFile(fsys, name)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", name, err)
	}
	var entries []string
	for _, line := range strings.Split(string(content), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		entries = append(entries, line)
	}
	return entries, nil
}

func parseAddresses(entries []string) ([]sdk.AccAddress, error) {
	codec := address.NewBech32Codec(appparams.Bech32PrefixAccAddr)
	var addrs []sdk.AccAddress
	for _, entry := range entries {
		bz, err := codec.StringToBytes(entry)
		if err != nil {
			return nil, fmt.Errorf("%w %q: %w", ErrInvalidAddress, entry, err)
		}
		addrs = append(addrs, bz)
	}
	return addrs, nil
}

func mustParseEmbedded() map[string]ChainConfig {
	fsys, err := fs.Sub(embedded, "lists")
	if err != nil {
		panic(fmt.Errorf("%s: open embedded lists: %w", UpgradeName, err))
	}
	configs, err := parseLists(fsys)
	if err != nil {
		panic(fmt.Errorf("%s: parse embedded lists: %w", UpgradeName, err))
	}
	return configs
}
