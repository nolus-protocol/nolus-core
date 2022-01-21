/* eslint-disable */
import Long from "long";
import _m0 from "protobufjs/minimal";
import { SuspendedState } from "../../../nolus/suspend/v1beta1/suspend";

export const protobufPackage = "nolus.suspend.v1beta1";

/** QueryParamsRequest is the request type for the Query/Params RPC method. */
export interface QuerySuspendRequest {}

/** QueryParamsResponse is the response type for the Query/Params RPC method. */
export interface QuerySuspendResponse {
  state?: SuspendedState;
}

function createBaseQuerySuspendRequest(): QuerySuspendRequest {
  return {};
}

export const QuerySuspendRequest = {
  encode(
    _: QuerySuspendRequest,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): QuerySuspendRequest {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseQuerySuspendRequest();
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

  fromJSON(_: any): QuerySuspendRequest {
    return {};
  },

  toJSON(_: QuerySuspendRequest): unknown {
    const obj: any = {};
    return obj;
  },

  fromPartial<I extends Exact<DeepPartial<QuerySuspendRequest>, I>>(
    _: I
  ): QuerySuspendRequest {
    const message = createBaseQuerySuspendRequest();
    return message;
  },
};

function createBaseQuerySuspendResponse(): QuerySuspendResponse {
  return { state: undefined };
}

export const QuerySuspendResponse = {
  encode(
    message: QuerySuspendResponse,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    if (message.state !== undefined) {
      SuspendedState.encode(message.state, writer.uint32(10).fork()).ldelim();
    }
    return writer;
  },

  decode(
    input: _m0.Reader | Uint8Array,
    length?: number
  ): QuerySuspendResponse {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseQuerySuspendResponse();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.state = SuspendedState.decode(reader, reader.uint32());
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  fromJSON(object: any): QuerySuspendResponse {
    return {
      state: isSet(object.state)
        ? SuspendedState.fromJSON(object.state)
        : undefined,
    };
  },

  toJSON(message: QuerySuspendResponse): unknown {
    const obj: any = {};
    message.state !== undefined &&
      (obj.state = message.state
        ? SuspendedState.toJSON(message.state)
        : undefined);
    return obj;
  },

  fromPartial<I extends Exact<DeepPartial<QuerySuspendResponse>, I>>(
    object: I
  ): QuerySuspendResponse {
    const message = createBaseQuerySuspendResponse();
    message.state =
      object.state !== undefined && object.state !== null
        ? SuspendedState.fromPartial(object.state)
        : undefined;
    return message;
  },
};

/** Query defines the gRPC querier service. */
export interface Query {
  /** this line is used by starport scaffolding # 2 */
  SuspendedState(request: QuerySuspendRequest): Promise<QuerySuspendResponse>;
}

export class QueryClientImpl implements Query {
  private readonly rpc: Rpc;
  constructor(rpc: Rpc) {
    this.rpc = rpc;
    this.SuspendedState = this.SuspendedState.bind(this);
  }
  SuspendedState(request: QuerySuspendRequest): Promise<QuerySuspendResponse> {
    const data = QuerySuspendRequest.encode(request).finish();
    const promise = this.rpc.request(
      "nolus.suspend.v1beta1.Query",
      "SuspendedState",
      data
    );
    return promise.then((data) =>
      QuerySuspendResponse.decode(new _m0.Reader(data))
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
