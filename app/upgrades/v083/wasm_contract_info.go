package v083

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math"

	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// wasmd v0.61.1 numbered ContractInfo.ibc2_port_id 7 and pushed extension to 8.
// v0.61.5 reverted that, restoring extension to 7 and giving ibc2_port_id 8.
// Every ContractInfo written while the chain ran v0.61.1 therefore stores the port
// id under the field number v0.61.8 decodes as extension, which fails to unmarshal.
const (
	legacyIBC2PortIDField = 7
	ibc2PortIDField       = 8

	protoWireVarint = 0
	protoWire64Bit  = 1
	protoWireBytes  = 2
	protoWire32Bit  = 5
)

// MigrateContractInfoIBC2PortID rewrites every x/wasm ContractInfo record from the
// field numbering wasmd v0.61.1 used to the numbering restored in v0.61.8, and returns
// the number of records rewritten. Records last written before the chain ran v0.61.1
// carry no field 7 at all and are left untouched.
//
// It must run before the module migrations, because anything that reads a contract
// panics on the unmigrated encoding.
func MigrateContractInfoIBC2PortID(ctx sdk.Context, wasmStoreKey storetypes.StoreKey, cdc codec.BinaryCodec) (int, error) {
	if wasmStoreKey == nil {
		return 0, errors.New("nil x/wasm store key")
	}
	contractStore := prefix.NewStore(ctx.KVStore(wasmStoreKey), wasmtypes.ContractKeyPrefix)

	rewritten, err := collectRewrites(contractStore, cdc)
	if err != nil {
		return 0, err
	}
	for _, record := range rewritten {
		contractStore.Set(record.key, record.value)
	}
	return len(rewritten), nil
}

type contractRecord struct {
	key   []byte
	value []byte
}

// collectRewrites buffers the rewritten records instead of writing them back during
// iteration, which is undefined behaviour on a cached KVStore. The buffer is a slice
// in iterator order rather than a map, so the writes stay deterministic across nodes.
func collectRewrites(contractStore prefix.Store, cdc codec.BinaryCodec) (rewritten []contractRecord, err error) {
	iter := contractStore.Iterator(nil, nil)
	defer func() {
		if closeErr := iter.Close(); closeErr != nil && err == nil {
			rewritten, err = nil, closeErr
		}
	}()

	for ; iter.Valid(); iter.Next() {
		migrated, changed, err := migrateContractInfoBz(iter.Value())
		if err != nil {
			return nil, fmt.Errorf("contract %X: %w", iter.Key(), err)
		}
		if !changed {
			continue
		}

		var contractInfo wasmtypes.ContractInfo
		if err := cdc.Unmarshal(migrated, &contractInfo); err != nil {
			return nil, fmt.Errorf("contract %X: migrated record does not decode: %w", iter.Key(), err)
		}
		if contractInfo.IBC2PortID == "" {
			return nil, fmt.Errorf("contract %X: migrated record lost its IBC v2 port id", iter.Key())
		}

		rewritten = append(rewritten, contractRecord{key: bytes.Clone(iter.Key()), value: migrated})
	}
	return rewritten, nil
}

// migrateContractInfoBz moves an ibc2_port_id stored under field 7 to field 8. Both
// tags are a single byte of the same wire type, so the payload and the field ordering
// are untouched and the result is byte-identical to what v0.61.8 encodes natively.
func migrateContractInfoBz(bz []byte) ([]byte, bool, error) {
	var out []byte
	sawField8 := false

	for i := 0; i < len(bz); {
		tagStart := i
		tag, tagLen := binary.Uvarint(bz[i:])
		if tagLen <= 0 {
			return nil, false, fmt.Errorf("malformed field tag at offset %d", i)
		}
		i += tagLen

		fieldNum := tag >> 3
		wireType := tag & 0x7

		size, err := payloadSize(bz[i:], wireType)
		if err != nil {
			return nil, false, fmt.Errorf("field %d: %w", fieldNum, err)
		}
		if i+size > len(bz) {
			return nil, false, fmt.Errorf("field %d: payload runs past end of record", fieldNum)
		}

		if fieldNum == ibc2PortIDField {
			sawField8 = true
		}

		if fieldNum == legacyIBC2PortIDField {
			if wireType != protoWireBytes {
				return nil, false, fmt.Errorf("field %d has wire type %d, want %d", fieldNum, wireType, protoWireBytes)
			}
			if tagLen != 1 {
				return nil, false, fmt.Errorf("field %d has a %d-byte tag, want 1", fieldNum, tagLen)
			}

			valueLen, lenLen := binary.Uvarint(bz[i:])
			value := bz[i+lenLen : i+lenLen+int(valueLen)]
			if !bytes.HasPrefix(value, []byte(wasmkeeper.PortIDPrefixV2)) {
				return nil, false, fmt.Errorf("field %d holds %q, which is not an IBC v2 port id", fieldNum, value)
			}

			if out == nil {
				out = bytes.Clone(bz)
			}
			out[tagStart] = byte(ibc2PortIDField<<3 | protoWireBytes)
		}

		i += size
	}

	if out != nil && sawField8 {
		return nil, false, fmt.Errorf(
			"record carries both field %d and field %d; rewriting would collide two values onto field %d",
			legacyIBC2PortIDField, ibc2PortIDField, ibc2PortIDField)
	}

	return out, out != nil, nil
}

func payloadSize(bz []byte, wireType uint64) (int, error) {
	switch wireType {
	case protoWireVarint:
		_, n := binary.Uvarint(bz)
		if n <= 0 {
			return 0, errors.New("malformed varint")
		}
		return n, nil
	case protoWire64Bit:
		return 8, nil
	case protoWireBytes:
		length, n := binary.Uvarint(bz)
		if n <= 0 {
			return 0, errors.New("malformed length prefix")
		}
		if length > math.MaxInt32 {
			return 0, fmt.Errorf("length prefix %d out of range", length)
		}
		return n + int(length), nil
	case protoWire32Bit:
		return 4, nil
	default:
		return 0, fmt.Errorf("unsupported wire type %d", wireType)
	}
}
