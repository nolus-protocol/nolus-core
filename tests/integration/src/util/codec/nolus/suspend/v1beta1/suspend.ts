/* eslint-disable */
import Long from "long";
import _m0 from "protobufjs/minimal";

export const protobufPackage = "nolus.suspend.v1beta1";

export interface SuspendedState {
  suspended: boolean;
  blockHeight: Long;
  adminAddress: string;
}

function createBaseSuspendedState(): SuspendedState {
  return { suspended: false, blockHeight: Long.ZERO, adminAddress: "" };
}

export const SuspendedState = {
  encode(
    message: SuspendedState,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    if (message.suspended === true) {
      writer.uint32(16).bool(message.suspended);
    }
    if (!message.blockHeight.isZero()) {
      writer.uint32(24).int64(message.blockHeight);
    }
    if (message.adminAddress !== "") {
      writer.uint32(34).string(message.adminAddress);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): SuspendedState {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseSuspendedState();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 2:
          message.suspended = reader.bool();
          break;
        case 3:
          message.blockHeight = reader.int64() as Long;
          break;
        case 4:
          message.adminAddress = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): SuspendedState {
    return {
      suspended: isSet(object.suspended) ? Boolean(object.suspended) : false,
      blockHeight: isSet(object.blockHeight)
        ? Long.fromString(object.blockHeight)
        : Long.ZERO,
      adminAddress: isSet(object.adminAddress)
        ? String(object.adminAddress)
        : "",
    };
  },

  toJSON(message: SuspendedState): unknown {
    const obj: any = {};
    message.suspended !== undefined && (obj.suspended = message.suspended);
    message.blockHeight !== undefined &&
      (obj.blockHeight = (message.blockHeight || Long.ZERO).toString());
    message.adminAddress !== undefined &&
      (obj.adminAddress = message.adminAddress);
    return obj;
  },

  fromPartial<I extends Exact<DeepPartial<SuspendedState>, I>>(
    object: I
  ): SuspendedState {
    const message = createBaseSuspendedState();
    message.suspended = object.suspended ?? false;
    message.blockHeight =
      object.blockHeight !== undefined && object.blockHeight !== null
        ? Long.fromValue(object.blockHeight)
        : Long.ZERO;
    message.adminAddress = object.adminAddress ?? "";
    return message;
  },
};

type Builtin =
  | Date
  | Function
  | Uint8Array
  | string
  | number
  | boolean
  | undefined;

export type DeepPartial<T> = T extends Builtin
  ? T
  : T extends Long
  ? string | number | Long
  : T extends Array<infer U>
  ? Array<DeepPartial<U>>
  : T extends ReadonlyArray<infer U>
  ? ReadonlyArray<DeepPartial<U>>
  : T extends {}
  ? { [K in keyof T]?: DeepPartial<T[K]> }
  : Partial<T>;

type KeysOfUnion<T> = T extends T ? keyof T : never;
export type Exact<P, I extends P> = P extends Builtin
  ? P
  : P & { [K in keyof P]: Exact<P[K], I[K]> } & Record<
        Exclude<keyof I, KeysOfUnion<P>>,
        never
      >;

if (_m0.util.Long !== Long) {
  _m0.util.Long = Long as any;
  _m0.configure();
}

function isSet(value: any): boolean {
  return value !== null && value !== undefined;
}
