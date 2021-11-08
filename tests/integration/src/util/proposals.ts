import Long from "long";
import _m0 from "protobufjs/minimal";
import { Any } from "cosmjs-types/google/protobuf/any";
import { Plan } from "cosmjs-types/cosmos/upgrade/v1beta1/upgrade";

// The following interfaces and encoding/decoding functions were generated using a modified version of:
// https://github.com/confio/cosmjs-types/blob/main/scripts/codegen.sh

/**
 * ClientUpdateProposal is a governance proposal. If it passes, the substitute
 * client's latest consensus state is copied over to the subject client. The proposal
 * handler may fail if the subject and the substitute do not match in client and
 * chain parameters (with exception to latest height, frozen height, and chain-id).
 */
export interface ClientUpdateProposal {
    /** the title of the update proposal */
    title: string;
    /** the description of the proposal */
    description: string;
    /** the client identifier for the client to be updated if the proposal passes */
    subjectClientId: string;
    /**
     * the substitute client identifier for the client standing in for the subject
     * client
     */
    substituteClientId: string;
}
  
/**
 * UpgradeProposal is a gov Content type for initiating an IBC breaking
 * upgrade.
 */
export interface UpgradeProposal {
    title: string;
    description: string;
    plan?: Plan;
    /**
     * An UpgradedClientState must be provided to perform an IBC breaking upgrade.
     * This will make the chain commit to the correct upgraded (self) client state
     * before the upgrade occurs, so that connecting chains can verify that the
     * new upgraded client is valid by verifying a proof on the previous version
     * of the chain. This will allow IBC connections to persist smoothly across
     * planned chain upgrades
     */
    upgradedClientState?: Any;
}

const baseClientUpdateProposal: object = {
    title: "",
    description: "",
    subjectClientId: "",
    substituteClientId: "",
};
  
export const ClientUpdateProposal = {
    encode(message: ClientUpdateProposal, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
        if (message.title !== "") {
            writer.uint32(10).string(message.title);
        }
        if (message.description !== "") {
            writer.uint32(18).string(message.description);
        }
        if (message.subjectClientId !== "") {
            writer.uint32(26).string(message.subjectClientId);
        }
        if (message.substituteClientId !== "") {
            writer.uint32(34).string(message.substituteClientId);
        }
        return writer;
    },
  
    decode(input: _m0.Reader | Uint8Array, length?: number): ClientUpdateProposal {
        const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = { ...baseClientUpdateProposal } as ClientUpdateProposal;
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.title = reader.string();
                    break;
                case 2:
                    message.description = reader.string();
                    break;
                case 3:
                    message.subjectClientId = reader.string();
                    break;
                case 4:
                    message.substituteClientId = reader.string();
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
            }
        }
        return message;
    },
  
    fromJSON(object: any): ClientUpdateProposal {
        const message = { ...baseClientUpdateProposal } as ClientUpdateProposal;
        if (object.title !== undefined && object.title !== null) {
            message.title = String(object.title);
        } else {
            message.title = "";
        }
        if (object.description !== undefined && object.description !== null) {
            message.description = String(object.description);
        } else {
            message.description = "";
        }
        if (object.subjectClientId !== undefined && object.subjectClientId !== null) {
            message.subjectClientId = String(object.subjectClientId);
        } else {
            message.subjectClientId = "";
        }
        if (object.substituteClientId !== undefined && object.substituteClientId !== null) {
            message.substituteClientId = String(object.substituteClientId);
        } else {
            message.substituteClientId = "";
        }
        return message;
    },
  
    toJSON(message: ClientUpdateProposal): unknown {
        const obj: any = {};
        message.title !== undefined && (obj.title = message.title);
        message.description !== undefined && (obj.description = message.description);
        message.subjectClientId !== undefined && (obj.subjectClientId = message.subjectClientId);
        message.substituteClientId !== undefined && (obj.substituteClientId = message.substituteClientId);
        return obj;
    },
  
    fromPartial(object: DeepPartial<ClientUpdateProposal>): ClientUpdateProposal {
        const message = { ...baseClientUpdateProposal } as ClientUpdateProposal;
        message.title = object.title ?? "";
        message.description = object.description ?? "";
        message.subjectClientId = object.subjectClientId ?? "";
        message.substituteClientId = object.substituteClientId ?? "";
        return message;
    },
  };

const baseUpgradeProposal: object = { title: "", description: "" };

export const UpgradeProposal = {
    encode(message: UpgradeProposal, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
        if (message.title !== "") {
            writer.uint32(10).string(message.title);
        }
        if (message.description !== "") {
            writer.uint32(18).string(message.description);
        }
        if (message.plan !== undefined) {
            Plan.encode(message.plan, writer.uint32(26).fork()).ldelim();
        }
        if (message.upgradedClientState !== undefined) {
            Any.encode(message.upgradedClientState, writer.uint32(34).fork()).ldelim();
        }
        return writer;
    },

    decode(input: _m0.Reader | Uint8Array, length?: number): UpgradeProposal {
        const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = { ...baseUpgradeProposal } as UpgradeProposal;
        while (reader.pos < end) {
            const tag = reader.uint32();
            switch (tag >>> 3) {
                case 1:
                    message.title = reader.string();
                    break;
                case 2:
                    message.description = reader.string();
                    break;
                case 3:
                    message.plan = Plan.decode(reader, reader.uint32());
                    break;
                case 4:
                    message.upgradedClientState = Any.decode(reader, reader.uint32());
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
            }
        }
        return message;
    },

    fromJSON(object: any): UpgradeProposal {
        const message = { ...baseUpgradeProposal } as UpgradeProposal;
        if (object.title !== undefined && object.title !== null) {
            message.title = String(object.title);
        } else {
            message.title = "";
        }
        if (object.description !== undefined && object.description !== null) {
            message.description = String(object.description);
        } else {
            message.description = "";
        }
        if (object.plan !== undefined && object.plan !== null) {
            message.plan = Plan.fromJSON(object.plan);
        } else {
            message.plan = undefined;
        }
        if (object.upgradedClientState !== undefined && object.upgradedClientState !== null) {
            message.upgradedClientState = Any.fromJSON(object.upgradedClientState);
        } else {
            message.upgradedClientState = undefined;
        }
        return message;
    },

    toJSON(message: UpgradeProposal): unknown {
        const obj: any = {};
        message.title !== undefined && (obj.title = message.title);
        message.description !== undefined && (obj.description = message.description);
        message.plan !== undefined && (obj.plan = message.plan ? Plan.toJSON(message.plan) : undefined);
        message.upgradedClientState !== undefined &&
            (obj.upgradedClientState = message.upgradedClientState
                ? Any.toJSON(message.upgradedClientState)
                : undefined);
        return obj;
  },

    fromPartial(object: DeepPartial<UpgradeProposal>): UpgradeProposal {
        const message = { ...baseUpgradeProposal } as UpgradeProposal;
        message.title = object.title ?? "";
        message.description = object.description ?? "";
        if (object.plan !== undefined && object.plan !== null) {
            message.plan = Plan.fromPartial(object.plan);
        } else {
            message.plan = undefined;
        }
        if (object.upgradedClientState !== undefined && object.upgradedClientState !== null) {
            message.upgradedClientState = Any.fromPartial(object.upgradedClientState);
        } else {
            message.upgradedClientState = undefined;
        }
        return message;
    },
};

type Builtin = Date | Function | Uint8Array | string | number | boolean | undefined | Long;
export type DeepPartial<T> = T extends Builtin
    ? T
    : T extends Array<infer U>
    ? Array<DeepPartial<U>>
    : T extends ReadonlyArray<infer U>
    ? ReadonlyArray<DeepPartial<U>>
    : T extends {}
    ? { [K in keyof T]?: DeepPartial<T[K]> }
    : Partial<T>;

if (_m0.util.Long !== Long) {
    _m0.util.Long = Long as any;
    _m0.configure();
}
