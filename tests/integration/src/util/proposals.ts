import { Any } from "cosmjs-types/google/protobuf/any";
import { Plan } from "cosmjs-types/cosmos/upgrade/v1beta1/upgrade";
import { Reader, Writer } from "protobufjs/minimal";

declare type Builtin = Date | Function | Uint8Array | string | number | boolean | undefined | Long;
export declare type DeepPartial<T> = T extends Builtin ? T : T extends Array<infer U> ? Array<DeepPartial<U>> : T extends ReadonlyArray<infer U> ? ReadonlyArray<DeepPartial<U>> : T extends {} ? {
    [K in keyof T]?: DeepPartial<T[K]>;
} : Partial<T>;

// IBC breaking upgrade proposal
export interface UpgradeProposal {
    title: string;
    description: string;
    plan?: Plan;
    upgradedClientState?: Any;
}

export class UpgradeProposal {
    public static encode(message: UpgradeProposal, writer?: Writer): Writer {
        if (writer === undefined) {
            writer = Writer.create();
        }
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
    }

    public static decode(input: Reader | Uint8Array, length?: number | undefined): UpgradeProposal {
        const reader = input instanceof Reader ? input : new Reader(input);
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = new UpgradeProposal();
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
                    message.plan = exports.Plan.decode(reader, reader.uint32());
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
    }
    public static fromJSON(object: any): UpgradeProposal {
        const message = new UpgradeProposal;
        if (object.title !== undefined && object.title !== null) {
            message.title = String(object.title);
        }
        else {
            message.title = "";
        }
        if (object.description !== undefined && object.description !== null) {
            message.description = String(object.description);
        }
        else {
            message.description = "";
        }
        if (object.plan !== undefined && object.plan !== null) {
            message.plan = exports.Plan.fromJSON(object.plan);
        }
        else {
            message.plan = undefined;
        }
        if (object.upgradedClientState !== undefined && object.upgradedClientState !== null) {
            message.upgradedClientState = Any.fromPartial(object.upgradedClientState);
        }
        else {
            message.upgradedClientState = undefined;
        }
        return message;
    }
    public static toJSON(message: UpgradeProposal): unknown {
        const obj: any = {};
        message.title !== undefined && (obj.title = message.title);
        message.description !== undefined && (obj.description = message.description);
        message.plan !== undefined && (obj.plan = message.plan ? exports.Plan.toJSON(message.plan) : undefined);
        message.upgradedClientState !== undefined &&
            (obj.upgradedClientState = message.upgradedClientState
                ? Any.toJSON(message.upgradedClientState)
                : undefined);
        return obj;
    }
    public static fromPartial(object: DeepPartial<UpgradeProposal>) {
        const message = new UpgradeProposal;
        if (object.title !== undefined && object.title !== null) {
            message.title = object.title;
        }
        else {
            message.title = "";
        }
        if (object.description !== undefined && object.description !== null) {
            message.description = object.description;
        }
        else {
            message.description = "";
        }
        if (object.plan !== undefined && object.plan !== null) {
            message.plan = exports.Plan.fromPartial(object.plan);
        }
        else {
            message.plan = undefined;
        }
        if (object.upgradedClientState !== undefined && object.upgradedClientState !== null) {
            message.upgradedClientState = Any.fromPartial(object.upgradedClientState);
        }
        else {
            message.upgradedClientState = undefined;
        }
        return message;
    }
};

// IBC client upgrade proposal
export interface ClientUpdateProposal {
    title: string;
    description: string;
    subjectClientId: string;
    substituteClientId: string;
}

export class ClientUpdateProposal {
    public static encode(message: ClientUpdateProposal, writer?: Writer) {
        if (writer === undefined) {
            writer = Writer.create()
        }
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
    }

    public static decode(input: Reader | Uint8Array, length?: number | undefined) {
        const reader = input instanceof Reader ? input : new Reader(input);
        let end = length === undefined ? reader.len : reader.pos + length;
        const message = new ClientUpdateProposal();
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
    }

    public static fromJSON(object: any): ClientUpdateProposal {
        const message = new ClientUpdateProposal();
        if (object.title !== undefined && object.title !== null) {
            message.title = String(object.title);
        }
        else {
            message.title = "";
        }
        if (object.description !== undefined && object.description !== null) {
            message.description = String(object.description);
        }
        else {
            message.description = "";
        }
        if (object.subjectClientId !== undefined && object.subjectClientId !== null) {
            message.subjectClientId = String(object.subjectClientId);
        }
        else {
            message.subjectClientId = "";
        }
        if (object.substituteClientId !== undefined && object.substituteClientId !== null) {
            message.substituteClientId = String(object.substituteClientId);
        }
        else {
            message.substituteClientId = "";
        }
        return message;
    }

    public static toJSON(message: ClientUpdateProposal): unknown {
        const obj: any = {};
        message.title !== undefined && (obj.title = message.title);
        message.description !== undefined && (obj.description = message.description);
        message.subjectClientId !== undefined && (obj.subjectClientId = message.subjectClientId);
        message.substituteClientId !== undefined && (obj.substituteClientId = message.substituteClientId);
        return obj;
    }
    
    public static fromPartial(object: DeepPartial<ClientUpdateProposal>) {
        const message = new ClientUpdateProposal();
        if (object.title !== undefined && object.title !== null) {
            message.title = object.title;
        }
        else {
            message.title = "";
        }
        if (object.description !== undefined && object.description !== null) {
            message.description = object.description;
        }
        else {
            message.description = "";
        }
        if (object.subjectClientId !== undefined && object.subjectClientId !== null) {
            message.subjectClientId = object.subjectClientId;
        }
        else {
            message.subjectClientId = "";
        }
        if (object.substituteClientId !== undefined && object.substituteClientId !== null) {
            message.substituteClientId = object.substituteClientId;
        }
        else {
            message.substituteClientId = "";
        }
        return message;
    }
};