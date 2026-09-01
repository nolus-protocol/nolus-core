package solanacarrier_test

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Nolus-Protocol/nolus-core/solanacarrier"
)

// Account layout the client builds the carrier with: a funded sentinel fee payer
// whose signature is never produced, the user, and the two invoked programs.
const (
	sentinelFeePayerKeyHex = "48ab05fd4f9c5a20ee631f7f45e8ac8ecf133b2ab59b247b452b6ea5dd5bf948"

	// MemoSq4gqABAXKb96qnH8TysNcWxMyWCqXgDLGmfcHr.
	memoProgramKeyHex = "054a535a992921064d24e87160da387c7c35b5ddbc92bb81e41fa8404105448d"

	blockhashHex = "d3266180b58ab52ad71ace42190f053f25dd7e4b74dba3a6c07aa8f7ec52201c"
)

const (
	systemProgramIndex = 2
	memoProgramIndex   = 3
)

// transferZeroLamports is the System program instruction the client pairs the memo
// with so hardware wallets have something renderable to approve.
var transferZeroLamports = []byte{0x02, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

var (
	sentinelFeePayerKey = hexKey(sentinelFeePayerKeyHex)
	memoProgramKey      = hexKey(memoProgramKeyHex)
	carrierBlockhash    = hexKey(blockhashHex)
)

func hexKey(s string) []byte {
	bz, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return bz
}

type solanaInstruction struct {
	programIDIndex byte
	accounts       []byte
	data           []byte
}

type solanaMessage struct {
	numRequiredSignatures byte
	numReadonlySigned     byte
	numReadonlyUnsigned   byte
	accountKeys           [][]byte
	recentBlockhash       []byte
	instructions          []solanaInstruction
}

func compactU16(n int) []byte {
	var out []byte
	for {
		b := byte(n & 0x7f)
		n >>= 7
		if n == 0 {
			return append(out, b)
		}
		out = append(out, b|0x80)
	}
}

func (m solanaMessage) encode() []byte {
	out := []byte{m.numRequiredSignatures, m.numReadonlySigned, m.numReadonlyUnsigned}
	out = append(out, compactU16(len(m.accountKeys))...)
	for _, key := range m.accountKeys {
		out = append(out, key...)
	}
	out = append(out, m.recentBlockhash...)
	out = append(out, compactU16(len(m.instructions))...)
	for _, instruction := range m.instructions {
		out = append(out, instruction.programIDIndex)
		out = append(out, compactU16(len(instruction.accounts))...)
		out = append(out, instruction.accounts...)
		out = append(out, compactU16(len(instruction.data))...)
		out = append(out, instruction.data...)
	}
	return out
}

func carrierMessage(signerKey, memoData []byte) solanaMessage {
	return solanaMessage{
		numRequiredSignatures: 2,
		numReadonlySigned:     0,
		numReadonlyUnsigned:   2,
		accountKeys: [][]byte{
			bytes.Clone(sentinelFeePayerKey),
			bytes.Clone(signerKey),
			make([]byte, 32),
			bytes.Clone(memoProgramKey),
		},
		recentBlockhash: bytes.Clone(carrierBlockhash),
		instructions: []solanaInstruction{
			{programIDIndex: systemProgramIndex, accounts: []byte{1, 1}, data: transferZeroLamports},
			{programIDIndex: memoProgramIndex, accounts: []byte{1}, data: memoData},
		},
	}
}

// padToLength appends an inert instruction sized so the encoded message lands on
// target exactly, which is how the packet-cap boundary cases are built.
func padToLength(t *testing.T, base solanaMessage, target int) solanaMessage {
	t.Helper()
	for n := 0; n <= target; n++ {
		padded := base
		padded.instructions = append(
			append([]solanaInstruction{}, base.instructions...),
			solanaInstruction{programIDIndex: systemProgramIndex, data: make([]byte, n)},
		)
		if len(padded.encode()) == target {
			return padded
		}
	}
	t.Fatalf("no inert-instruction size encodes this message to exactly %d bytes", target)
	return solanaMessage{}
}

func repeatKeys(n int) [][]byte {
	keys := make([][]byte, n)
	for i := range keys {
		key := make([]byte, 32)
		key[0] = byte(i)
		key[1] = byte(i >> 8)
		keys[i] = key
	}
	return keys
}

func minimalMessage() solanaMessage {
	return solanaMessage{
		numRequiredSignatures: 1,
		numReadonlySigned:     0,
		numReadonlyUnsigned:   1,
		accountKeys:           repeatKeys(2),
		recentBlockhash:       bytes.Repeat([]byte{0x03}, 32),
		instructions: []solanaInstruction{
			{programIDIndex: 1, accounts: []byte{0}, data: []byte{0x09, 0x08}},
		},
	}
}

func TestParseLegacyMessageReadsEveryField(t *testing.T) {
	message := minimalMessage()

	got, err := solanacarrier.ParseLegacyMessage(message.encode())
	require.NoError(t, err)
	require.Equal(t, &solanacarrier.LegacyMessage{
		NumRequiredSignatures:       1,
		NumReadonlySignedAccounts:   0,
		NumReadonlyUnsignedAccounts: 1,
		AccountKeys:                 message.accountKeys,
		RecentBlockhash:             message.recentBlockhash,
		Instructions: []solanacarrier.CompiledInstruction{
			{ProgramIDIndex: 1, Accounts: []uint8{0}, Data: []byte{0x09, 0x08}},
		},
	}, got)
}

func TestParseLegacyMessageReadsCapturedCarrier(t *testing.T) {
	got, err := solanacarrier.ParseLegacyMessage(mustDecodeHex(t, phantomCarrierMessageHex))
	require.NoError(t, err)
	require.Len(t, got.Instructions, 4, "the captured carrier holds the two client instructions plus Phantom's two ComputeBudget injections")
}

func TestParseLegacyMessageAcceptsCompactU16Widths(t *testing.T) {
	tests := []struct {
		name    string
		message func() solanaMessage
		assert  func(t *testing.T, got *solanacarrier.LegacyMessage)
	}{
		{
			name: "account count at the one-byte maximum",
			message: func() solanaMessage {
				message := minimalMessage()
				message.accountKeys = repeatKeys(0x7f)
				return message
			},
			assert: func(t *testing.T, got *solanacarrier.LegacyMessage) {
				t.Helper()
				require.Len(t, got.AccountKeys, 0x7f)
			},
		},
		{
			name: "account count at the two-byte minimum",
			message: func() solanaMessage {
				message := minimalMessage()
				message.accountKeys = repeatKeys(0x80)
				return message
			},
			assert: func(t *testing.T, got *solanacarrier.LegacyMessage) {
				t.Helper()
				require.Len(t, got.AccountKeys, 0x80)
			},
		},
		{
			name: "instruction data at the three-byte maximum",
			message: func() solanaMessage {
				message := minimalMessage()
				message.instructions[0].data = bytes.Repeat([]byte{0xab}, 16383)
				return message
			},
			assert: func(t *testing.T, got *solanacarrier.LegacyMessage) {
				t.Helper()
				require.Len(t, got.Instructions[0].Data, 16383)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := solanacarrier.ParseLegacyMessage(tc.message().encode())
			require.NoError(t, err)
			tc.assert(t, got)
		})
	}
}

func TestParseLegacyMessageRejectsMalformedInput(t *testing.T) {
	tests := []struct {
		name  string
		input func(t *testing.T) []byte
	}{
		{
			name:  "empty input",
			input: func(t *testing.T) []byte { t.Helper(); return nil },
		},
		{
			name:  "header only",
			input: func(t *testing.T) []byte { t.Helper(); return []byte{0x01, 0x00, 0x01} },
		},
		{
			name: "truncated account keys",
			input: func(t *testing.T) []byte {
				t.Helper()
				return minimalMessage().encode()[:3+1+40]
			},
		},
		{
			name: "truncated blockhash",
			input: func(t *testing.T) []byte {
				t.Helper()
				return minimalMessage().encode()[:3+1+64+10]
			},
		},
		{
			name: "missing instruction count",
			input: func(t *testing.T) []byte {
				t.Helper()
				return minimalMessage().encode()[:3+1+64+32]
			},
		},
		{
			name: "truncated instruction data",
			input: func(t *testing.T) []byte {
				t.Helper()
				encoded := minimalMessage().encode()
				return encoded[:len(encoded)-1]
			},
		},
		{
			name: "trailing bytes after the last instruction",
			input: func(t *testing.T) []byte {
				t.Helper()
				return append(minimalMessage().encode(), 0x00)
			},
		},
		{
			name: "versioned message prefix",
			input: func(t *testing.T) []byte {
				t.Helper()
				return append([]byte{0x80}, minimalMessage().encode()...)
			},
		},
		{
			name: "non-canonical compact-u16 length",
			input: func(t *testing.T) []byte {
				t.Helper()
				out := []byte{0x01, 0x00, 0x00, 0x81, 0x00}
				out = append(out, bytes.Repeat([]byte{0x01}, 32)...)
				out = append(out, bytes.Repeat([]byte{0x03}, 32)...)
				return append(out, 0x00)
			},
		},
		{
			name: "program id index out of range",
			input: func(t *testing.T) []byte {
				t.Helper()
				message := minimalMessage()
				message.instructions[0].programIDIndex = byte(len(message.accountKeys))
				return message.encode()
			},
		},
		{
			name: "program id index points at the fee payer",
			input: func(t *testing.T) []byte {
				t.Helper()
				message := minimalMessage()
				message.instructions[0].programIDIndex = 0
				return message.encode()
			},
		},
		{
			name: "instruction account index out of range",
			input: func(t *testing.T) []byte {
				t.Helper()
				message := minimalMessage()
				message.instructions[0].accounts = []byte{byte(len(message.accountKeys))}
				return message.encode()
			},
		},
		{
			name: "required signatures exceed account keys",
			input: func(t *testing.T) []byte {
				t.Helper()
				message := minimalMessage()
				message.numRequiredSignatures = byte(len(message.accountKeys)) + 1
				return message.encode()
			},
		},
		{
			name: "no writable signer",
			input: func(t *testing.T) []byte {
				t.Helper()
				message := minimalMessage()
				message.numReadonlySigned = message.numRequiredSignatures
				return message.encode()
			},
		},
		{
			name: "signing and read-only areas overlap",
			input: func(t *testing.T) []byte {
				t.Helper()
				message := minimalMessage()
				message.numRequiredSignatures = 2
				message.numReadonlySigned = 1
				message.numReadonlyUnsigned = 1
				return message.encode()
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := solanacarrier.ParseLegacyMessage(tc.input(t))
			require.Error(t, err)
		})
	}
}

func FuzzParseLegacyMessage(f *testing.F) {
	f.Add([]byte(nil))
	f.Add([]byte{0x00})
	f.Add([]byte{0x01, 0x00, 0x01})
	f.Add([]byte{0x80, 0x01, 0x00, 0x01})
	f.Add(minimalMessage().encode())
	f.Add(hexKey(phantomCarrierMessageHex))

	f.Fuzz(func(t *testing.T, data []byte) {
		message, err := solanacarrier.ParseLegacyMessage(data)
		if err != nil {
			return
		}

		require.Len(t, message.RecentBlockhash, 32)
		require.LessOrEqual(t, int(message.NumRequiredSignatures), len(message.AccountKeys))
		for _, key := range message.AccountKeys {
			require.Len(t, key, 32)
		}
		for _, instruction := range message.Instructions {
			require.Less(t, int(instruction.ProgramIDIndex), len(message.AccountKeys))
			for _, index := range instruction.Accounts {
				require.Less(t, int(index), len(message.AccountKeys))
			}
		}
	})
}
