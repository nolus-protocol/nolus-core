package solanacarrier_test

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Nolus-Protocol/nolus-core/solanacarrier"
)

// Account keys the carrier is laid out with: the governance-pinned fee payer at
// index 0 whose signature is never produced, the user, and the invoked programs.
const (
	// EvysLTsKMWc2A9kNUBaeJF99kvZvPLjfbow2Yk4gQYXZ, the off-curve sentinel the
	// x/solanacarrier fee-payer parameter is seeded with.
	paramFeePayerKeyHex = "cefc037c11436dc85b6943b43c94ec50c0493696cb7e797e84c1763038e9ad70"

	// The fee payer of the 2026-08-25 Phantom + Ledger capture, which predates
	// the fee-payer parameter.
	captureFeePayerKeyHex = "48ab05fd4f9c5a20ee631f7f45e8ac8ecf133b2ab59b247b452b6ea5dd5bf948"

	// MemoSq4gqABAXKb96qnH8TysNcWxMyWCqXgDLGmfcHr.
	memoProgramKeyHex = "054a535a992921064d24e87160da387c7c35b5ddbc92bb81e41fa8404105448d"

	// ComputeBudget111111111111111111111111111111.
	computeBudgetProgramKeyHex = "0306466fe5211732ffecadba72c39be7bc8ce5bbc5f7126b2c439b3a40000000"

	blockhashHex = "d3266180b58ab52ad71ace42190f053f25dd7e4b74dba3a6c07aa8f7ec52201c"
)

// transferZeroLamports is a System program transfer of nothing, the instruction
// hardware wallets were previously paired with so they had something renderable
// to approve.
var transferZeroLamports = []byte{0x02, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

var (
	paramFeePayerKey        = hexKey(paramFeePayerKeyHex)
	captureFeePayerKey      = hexKey(captureFeePayerKeyHex)
	memoProgramKey          = hexKey(memoProgramKeyHex)
	computeBudgetProgramKey = hexKey(computeBudgetProgramKeyHex)
	systemProgramKey        = make([]byte, 32)
	carrierBlockhash        = hexKey(blockhashHex)
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

// testInstruction names its program by key rather than by account index so cases
// can be written without tracking the account table buildMessage lays out.
type testInstruction struct {
	programKey []byte
	accounts   []byte
	data       []byte
}

func memoInstruction(data []byte) testInstruction {
	return testInstruction{programKey: memoProgramKey, accounts: []byte{1}, data: data}
}

func computeBudgetInstruction(data ...byte) testInstruction {
	return testInstruction{programKey: computeBudgetProgramKey, data: data}
}

func systemTransferInstruction() testInstruction {
	return testInstruction{programKey: systemProgramKey, accounts: []byte{1, 1}, data: transferZeroLamports}
}

func unknownProgramInstruction(data []byte) testInstruction {
	return testInstruction{programKey: bytes.Repeat([]byte{0x11}, 32), accounts: []byte{1}, data: data}
}

// buildMessage lays the account table out as [feePayer, signer, program keys in
// first-use order] and compiles the instructions against it.
func buildMessage(feePayer, signerKey []byte, instructions ...testInstruction) solanaMessage {
	message := solanaMessage{
		numRequiredSignatures: 2,
		numReadonlySigned:     0,
		accountKeys:           [][]byte{bytes.Clone(feePayer), bytes.Clone(signerKey)},
		recentBlockhash:       bytes.Clone(carrierBlockhash),
	}
	for _, instruction := range instructions {
		message.instructions = append(message.instructions, solanaInstruction{
			programIDIndex: programIndex(&message, instruction.programKey),
			accounts:       instruction.accounts,
			data:           instruction.data,
		})
	}
	return message
}

// programIndex returns the account index of key, appending it to the read-only
// unsigned region of the table when it is not there yet.
func programIndex(message *solanaMessage, key []byte) byte {
	for i, existing := range message.accountKeys {
		if bytes.Equal(existing, key) {
			return byte(i)
		}
	}
	message.accountKeys = append(message.accountKeys, bytes.Clone(key))
	message.numReadonlyUnsigned++
	return byte(len(message.accountKeys) - 1)
}

// carrierMessage is the layout the client builds: the parameter's fee payer, the
// signer, and the single bound memo.
func carrierMessage(signerKey, memoData []byte) solanaMessage {
	return buildMessage(paramFeePayerKey, signerKey, memoInstruction(memoData))
}

// padToLength appends an allowlisted ComputeBudget instruction sized so the
// encoded message lands on target exactly, which is how the packet-cap boundary
// cases are built.
func padToLength(t *testing.T, base solanaMessage, target int) solanaMessage {
	t.Helper()
	for n := 0; n <= target; n++ {
		padded := base
		padded.accountKeys = append([][]byte{}, base.accountKeys...)
		padded.instructions = append([]solanaInstruction{}, base.instructions...)
		padded.instructions = append(padded.instructions, solanaInstruction{
			programIDIndex: programIndex(&padded, computeBudgetProgramKey),
			data:           make([]byte, n),
		})
		if len(padded.encode()) == target {
			return padded
		}
	}
	t.Fatalf("no ComputeBudget-instruction size encodes this message to exactly %d bytes", target)
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
