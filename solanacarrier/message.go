package solanacarrier

import (
	"errors"
	"fmt"
)

const (
	pubKeyLen     = 32
	blockhashLen  = 32
	maxCompactU16 = 0xffff

	// minInstructionLen is the smallest an encoded instruction can be: a
	// program-id-index byte plus a zero-valued account count and data length,
	// each a single-byte compact-u16.
	minInstructionLen = 3
)

var (
	errEmptyMessage       = errors.New("empty message")
	errVersionedMessage   = errors.New("versioned message marker is set; only legacy messages are accepted")
	errNonCanonicalLength = errors.New("non-canonical compact-u16 length")
	errLengthOverflow     = errors.New("compact-u16 length exceeds 16 bits")
	errBufferOverrun      = errors.New("declared length exceeds the remaining buffer")
	errTrailingBytes      = errors.New("trailing bytes after the last instruction")
)

// CompiledInstruction is a Solana instruction with its program and accounts
// referenced by index into the message's account-key table.
type CompiledInstruction struct {
	ProgramIDIndex uint8
	Accounts       []uint8
	Data           []byte
}

// LegacyMessage is a decoded canonical Solana legacy transaction message.
type LegacyMessage struct {
	NumRequiredSignatures       uint8
	NumReadonlySignedAccounts   uint8
	NumReadonlyUnsignedAccounts uint8
	AccountKeys                 [][]byte
	RecentBlockhash             []byte
	Instructions                []CompiledInstruction
}

type reader struct {
	data []byte
	pos  int
}

func (r *reader) remaining() int {
	return len(r.data) - r.pos
}

func (r *reader) readByte() (byte, error) {
	if r.remaining() < 1 {
		return 0, fmt.Errorf("read byte: %w", errBufferOverrun)
	}
	b := r.data[r.pos]
	r.pos++
	return b, nil
}

func (r *reader) readBytes(n int) ([]byte, error) {
	if n > r.remaining() {
		return nil, fmt.Errorf("read %d bytes: %w", n, errBufferOverrun)
	}
	b := r.data[r.pos : r.pos+n]
	r.pos += n
	return b, nil
}

// readCompactU16 decodes a shortvec length, rejecting any non-minimal encoding
// (a trailing zero continuation such as 0x81 0x00) and any value above 0xffff.
func (r *reader) readCompactU16() (int, error) {
	var val uint32
	for i := 0; ; i++ {
		if i == 3 {
			return 0, errLengthOverflow
		}
		b, err := r.readByte()
		if err != nil {
			return 0, err
		}
		val |= uint32(b&0x7f) << (7 * i)
		if b&0x80 == 0 {
			if i > 0 && b == 0 {
				return 0, errNonCanonicalLength
			}
			if val > maxCompactU16 {
				return 0, errLengthOverflow
			}
			return int(val), nil
		}
	}
}

// ParseLegacyMessage decodes attacker-supplied bytes into a canonical Solana
// legacy message. It never panics, bounds every allocation against the
// remaining buffer before making it, and rejects versioned messages, trailing
// bytes, non-canonical lengths, and any message failing Solana's sanitize
// invariants.
func ParseLegacyMessage(data []byte) (*LegacyMessage, error) {
	if len(data) == 0 {
		return nil, errEmptyMessage
	}

	r := &reader{data: data}

	numRequiredSignatures, err := r.readByte()
	if err != nil {
		return nil, err
	}
	// The high bit of the first byte marks a versioned message; legacy messages
	// never set it.
	if numRequiredSignatures&0x80 != 0 {
		return nil, errVersionedMessage
	}
	numReadonlySigned, err := r.readByte()
	if err != nil {
		return nil, err
	}
	numReadonlyUnsigned, err := r.readByte()
	if err != nil {
		return nil, err
	}

	accountCount, err := r.readCompactU16()
	if err != nil {
		return nil, err
	}
	if accountCount > r.remaining()/pubKeyLen {
		return nil, fmt.Errorf("account keys: %w", errBufferOverrun)
	}
	accountKeys := make([][]byte, accountCount)
	for i := 0; i < accountCount; i++ {
		key, err := r.readBytes(pubKeyLen)
		if err != nil {
			return nil, err
		}
		accountKeys[i] = key
	}

	blockhash, err := r.readBytes(blockhashLen)
	if err != nil {
		return nil, err
	}

	instructionCount, err := r.readCompactU16()
	if err != nil {
		return nil, err
	}
	if instructionCount > r.remaining()/minInstructionLen {
		return nil, fmt.Errorf("instructions: %w", errBufferOverrun)
	}
	instructions := make([]CompiledInstruction, instructionCount)
	for i := 0; i < instructionCount; i++ {
		programIDIndex, err := r.readByte()
		if err != nil {
			return nil, err
		}

		accountCount, err := r.readCompactU16()
		if err != nil {
			return nil, err
		}
		accounts, err := r.readBytes(accountCount)
		if err != nil {
			return nil, err
		}

		dataLen, err := r.readCompactU16()
		if err != nil {
			return nil, err
		}
		data, err := r.readBytes(dataLen)
		if err != nil {
			return nil, err
		}

		instructions[i] = CompiledInstruction{
			ProgramIDIndex: programIDIndex,
			Accounts:       accounts,
			Data:           data,
		}
	}

	if r.remaining() != 0 {
		return nil, errTrailingBytes
	}

	message := &LegacyMessage{
		NumRequiredSignatures:       numRequiredSignatures,
		NumReadonlySignedAccounts:   numReadonlySigned,
		NumReadonlyUnsignedAccounts: numReadonlyUnsigned,
		AccountKeys:                 accountKeys,
		RecentBlockhash:             blockhash,
		Instructions:                instructions,
	}
	if err := message.sanitize(); err != nil {
		return nil, err
	}
	return message, nil
}

// sanitize mirrors Solana's Message::sanitize: the signing and read-only
// regions must fit within the account-key table, at least one writable signer
// must exist, and every instruction index must be in range with no program
// pointing at the fee payer at index 0.
func (m *LegacyMessage) sanitize() error {
	numKeys := len(m.AccountKeys)

	if int(m.NumRequiredSignatures)+int(m.NumReadonlyUnsignedAccounts) > numKeys {
		return fmt.Errorf("required and read-only unsigned accounts (%d) exceed account keys (%d)",
			int(m.NumRequiredSignatures)+int(m.NumReadonlyUnsignedAccounts), numKeys)
	}
	if int(m.NumReadonlySignedAccounts) >= int(m.NumRequiredSignatures) {
		return fmt.Errorf("read-only signed accounts (%d) leave no writable signer among %d required signatures",
			m.NumReadonlySignedAccounts, m.NumRequiredSignatures)
	}

	for _, instruction := range m.Instructions {
		if int(instruction.ProgramIDIndex) >= numKeys {
			return fmt.Errorf("program id index %d out of range for %d account keys", instruction.ProgramIDIndex, numKeys)
		}
		if instruction.ProgramIDIndex == 0 {
			return errors.New("program id index points at the fee payer")
		}
		for _, accountIndex := range instruction.Accounts {
			if int(accountIndex) >= numKeys {
				return fmt.Errorf("account index %d out of range for %d account keys", accountIndex, numKeys)
			}
		}
	}
	return nil
}
