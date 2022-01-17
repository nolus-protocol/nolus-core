/* eslint-disable */
import Long from "long";
import _m0 from "protobufjs/minimal";

export const protobufPackage = "nolus.suspend.v1beta1";

export interface MsgSuspend {
  fromAddress: string;
  blockHeight: Long;
}

export interface MsgUnsuspend {
  fromAddress: string;
}

export interface MsgSuspendResponse {}

export interface MsgUnsuspendResponse {}

function createBaseMsgSuspend(): MsgSuspend {
  return { fromAddress: "", blockHeight: Long.ZERO };
}

export const MsgSuspend = {
  encode(
    message: MsgSuspend,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    if (message.fromAddress !== "") {
      writer.uint32(10).string(message.fromAddress);
    }
    if (!message.blockHeight.isZero()) {
      writer.uint32(24).int64(message.blockHeight);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): MsgSuspend {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseMsgSuspend();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.fromAddress = reader.string();
          break;
        case 3:
          message.blockHeight = reader.int64() as Long;
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): MsgSuspend {
    return {
      fromAddress: isSet(object.fromAddress) ? String(object.fromAddress) : "",
      blockHeight: isSet(object.blockHeight)
        ? Long.fromString(object.blockHeight)
        : Long.ZERO,
    };
  },

  toJSON(message: MsgSuspend): unknown {
    const obj: any = {};
    message.fromAddress !== undefined &&
      (obj.fromAddress = message.fromAddress);
    message.blockHeight !== undefined &&
      (obj.blockHeight = (message.blockHeight || Long.ZERO).toString());
    return obj;
  },

  fromPartial<I extends Exact<DeepPartial<MsgSuspend>, I>>(
    object: I
  ): MsgSuspend {
    const message = createBaseMsgSuspend();
    message.fromAddress = object.fromAddress ?? "";
    message.blockHeight =
      object.blockHeight !== undefined && object.blockHeight !== null
        ? Long.fromValue(object.blockHeight)
        : Long.ZERO;
    return message;
  },
};

function createBaseMsgUnsuspend(): MsgUnsuspend {
  return { fromAddress: "" };
}

export const MsgUnsuspend = {
  encode(
    message: MsgUnsuspend,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    if (message.fromAddress !== "") {
      writer.uint32(10).string(message.fromAddress);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): MsgUnsuspend {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseMsgUnsuspend();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.fromAddress = reader.string();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): MsgUnsuspend {
    return {
      fromAddress: isSet(object.fromAddress) ? String(object.fromAddress) : "",
    };
  },

  toJSON(message: MsgUnsuspend): unknown {
    const obj: any = {};
    message.fromAddress !== undefined &&
      (obj.fromAddress = message.fromAddress);
    return obj;
  },

  fromPartial<I extends Exact<DeepPartial<MsgUnsuspend>, I>>(
    object: I
  ): MsgUnsuspend {
    const message = createBaseMsgUnsuspend();
    message.fromAddress = object.fromAddress ?? "";
    return message;
  },
};

function createBaseMsgSuspendResponse(): MsgSuspendResponse {
  return {};
}

export const MsgSuspendResponse = {
  encode(
    _: MsgSuspendResponse,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): MsgSuspendResponse {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseMsgSuspendResponse();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(_: any): MsgSuspendResponse {
    return {};
  },

  toJSON(_: MsgSuspendResponse): unknown {
    const obj: any = {};
    return obj;
  },

  fromPartial<I extends Exact<DeepPartial<MsgSuspendResponse>, I>>(
    _: I
  ): MsgSuspendResponse {
    const message = createBaseMsgSuspendResponse();
    return message;
  },
};

function createBaseMsgUnsuspendResponse(): MsgUnsuspendResponse {
  return {};
}

export const MsgUnsuspendResponse = {
  encode(
    _: MsgUnsuspendResponse,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    return writer;
  },

  decode(
    input: _m0.Reader | Uint8Array,
    length?: number
  ): MsgUnsuspendResponse {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseMsgUnsuspendResponse();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(_: any): MsgUnsuspendResponse {
    return {};
  },

  toJSON(_: MsgUnsuspendResponse): unknown {
    const obj: any = {};
    return obj;
  },

  fromPartial<I extends Exact<DeepPartial<MsgUnsuspendResponse>, I>>(
    _: I
  ): MsgUnsuspendResponse {
    const message = createBaseMsgUnsuspendResponse();
    return message;
  },
};

/** Msg defines the Msg service. */
export interface Msg {
  Suspend(request: MsgSuspend): Promise<MsgSuspendResponse>;
  /** this line is used by starport scaffolding # proto/tx/rpc */
  Unsuspend(request: MsgUnsuspend): Promise<MsgUnsuspendResponse>;
}

export class MsgClientImpl implements Msg {
  private readonly rpc: Rpc;
  constructor(rpc: Rpc) {
    this.rpc = rpc;
    this.Suspend = this.Suspend.bind(this);
    this.Unsuspend = this.Unsuspend.bind(this);
  }
  Suspend(request: MsgSuspend): Promise<MsgSuspendResponse> {
    const data = MsgSuspend.encode(request).finish();
    const promise = this.rpc.request(
      "nolus.suspend.v1beta1.Msg",
      "Suspend",
      data
    );
    return promise.then((data) =>
      MsgSuspendResponse.decode(new _m0.Reader(data))
    );
  }

  Unsuspend(request: MsgUnsuspend): Promise<MsgUnsuspendResponse> {
    const data = MsgUnsuspend.encode(request).finish();
    const promise = this.rpc.request(
      "nolus.suspend.v1beta1.Msg",
      "Unsuspend",
      data
    );
    return promise.then((data) =>
      MsgUnsuspendResponse.decode(new _m0.Reader(data))
    );
  }
}

interface Rpc {
  request(
    service: string,
    method: string,
    data: Uint8Array
  ): Promise<Uint8Array>;
}

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
