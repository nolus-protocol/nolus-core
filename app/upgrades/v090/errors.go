package v090

import "errors"

var (
	ErrUnknownListFile    = errors.New("unknown list file")
	ErrMissingRecipient   = errors.New("missing recipient")
	ErrMultipleRecipients = errors.New("more than one recipient")
	ErrInvalidAddress     = errors.New("invalid address")
	ErrDuplicateAddress   = errors.New("duplicate address")
	ErrRecipientListed    = errors.New("recipient listed")
	ErrProtectedAddress   = errors.New("module account or contract listed")
)
