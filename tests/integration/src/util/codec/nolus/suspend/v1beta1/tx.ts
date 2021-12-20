/* eslint-disable */
import Long from "long";
import _m0 from "protobufjs/minimal";

export const protobufPackage = "nolus.suspend.v1beta1";

export interface MsgChangeSuspended {
  fromAddress: string;
  suspended: boolean;
  blockHeight: Long;
}

export interface MsgChangeSuspendedResponse {}

const baseMsgChangeSuspended: object = {
  fromAddress: "",
  suspended: false,
  blockHeight: Long.ZERO,
};

export const MsgChangeSuspended = {
  encode(
    message: MsgChangeSuspended,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    if (message.fromAddress !== "") {
      writer.uint32(10).string(message.fromAddress);
    }
    if (message.suspended === true) {
      writer.uint32(16).bool(message.suspended);
    }
    if (!message.blockHeight.isZero()) {
      writer.uint32(24).int64(message.blockHeight);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): MsgChangeSuspended {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = { ...baseMsgChangeSuspended } as MsgChangeSuspended;
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.fromAddress = reader.string();
          break;
        case 2:
          message.suspended = reader.bool();
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

  fromJSON(object: any): MsgChangeSuspended {
    const message = { ...baseMsgChangeSuspended } as MsgChangeSuspended;
    message.fromAddress =
      object.fromAddress !== undefined && object.fromAddress !== null
        ? String(object.fromAddress)
        : "";
    message.suspended =
      object.suspended !== undefined && object.suspended !== null
        ? Boolean(object.suspended)
        : false;
    message.blockHeight =
      object.blockHeight !== undefined && object.blockHeight !== null
        ? Long.fromString(object.blockHeight)
        : Long.ZERO;
    return message;
  },

  toJSON(message: MsgChangeSuspended): unknown {
    const obj: any = {};
    message.fromAddress !== undefined &&
      (obj.fromAddress = message.fromAddress);
    message.suspended !== undefined && (obj.suspended = message.suspended);
    message.blockHeight !== undefined &&
      (obj.blockHeight = (message.blockHeight || Long.ZERO).toString());
    return obj;
  },

  fromPartial<I extends Exact<DeepPartial<MsgChangeSuspended>, I>>(
    object: I
  ): MsgChangeSuspended {
    const message = { ...baseMsgChangeSuspended } as MsgChangeSuspended;
    message.fromAddress = object.fromAddress ?? "";
    message.suspended = object.suspended ?? false;
    message.blockHeight =
      object.blockHeight !== undefined && object.blockHeight !== null
        ? Long.fromValue(object.blockHeight)
        : Long.ZERO;
    return message;
  },
};

const baseMsgChangeSuspendedResponse: object = {};

export const MsgChangeSuspendedResponse = {
  encode(
    _: MsgChangeSuspendedResponse,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    return writer;
  },

  decode(
    input: _m0.Reader | Uint8Array,
    length?: number
  ): MsgChangeSuspendedResponse {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = {
      ...baseMsgChangeSuspendedResponse,
    } as MsgChangeSuspendedResponse;
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

  fromJSON(_: any): MsgChangeSuspendedResponse {
    const message = {
      ...baseMsgChangeSuspendedResponse,
    } as MsgChangeSuspendedResponse;
    return message;
  },

  toJSON(_: MsgChangeSuspendedResponse): unknown {
    const obj: any = {};
    return obj;
  },

  fromPartial<I extends Exact<DeepPartial<MsgChangeSuspendedResponse>, I>>(
    _: I
  ): MsgChangeSuspendedResponse {
    const message = {
      ...baseMsgChangeSuspendedResponse,
    } as MsgChangeSuspendedResponse;
    return message;
  },
};

/** Msg defines the Msg service. */
export interface Msg {
  /** this line is used by starport scaffolding # proto/tx/rpc */
  ChangeSuspend(request: MsgChangeSuspended): Promise<MsgChangeSuspendedResponse>;
}

export class MsgClientImpl implements Msg {
  private readonly rpc: Rpc;
  constructor(rpc: Rpc) {
    this.rpc = rpc;
    this.ChangeSuspend = this.ChangeSuspend.bind(this);
  }
  ChangeSuspend(request: MsgChangeSuspended): Promise<MsgChangeSuspendedResponse> {
    const data = MsgChangeSuspended.encode(request).finish();
    const promise = this.rpc.request(
      "nolus.suspend.v1beta1.Msg",
      "ChangeSuspend",
      data
    );
    return promise.then((data) =>
      MsgChangeSuspendedResponse.decode(new _m0.Reader(data))
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
