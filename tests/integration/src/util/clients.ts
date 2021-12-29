import {SigningCosmWasmClient, SigningCosmWasmClientOptions} from "@cosmjs/cosmwasm-stargate";
import {DirectSecp256k1Wallet, Registry} from "@cosmjs/proto-signing";
import {fromHex} from "@cosmjs/encoding";
import {
    MsgClearAdmin,
    MsgExecuteContract,
    MsgInstantiateContract,
    MsgMigrateContract,
    MsgStoreCode,
    MsgUpdateAdmin
} from "cosmjs-types/cosmwasm/wasm/v1/tx";
import {defaultRegistryTypes} from "@cosmjs/stargate";
import {MsgCreateVestingAccount, protobufPackage as vestingPackage} from "./codec/cosmos/vesting/v1beta1/tx";

let validatorPrivKey = fromHex(process.env.VALIDATOR_PRIV_KEY as string);
let periodicPrivKey = fromHex(process.env.PERIODIC_PRIV_KEY as string);
let user1PrivKey = fromHex(process.env.USR_1_PRIV_KEY as string);
let user2PrivKey = fromHex(process.env.USR_2_PRIV_KEY as string);
let delayedVestingPrivKey = fromHex(process.env.DELAYED_VESTING_PRIV_KEY as string);


export const DEFAULT_FEE = {
    amount: [{denom: "unolus", amount: "12"}],
    gas: "100000"
};

export async function getWallet(privateKey: Uint8Array): Promise<DirectSecp256k1Wallet> {
    return await DirectSecp256k1Wallet.fromKey(privateKey, "nolus");
}

export async function getClient(privateKey: Uint8Array): Promise<SigningCosmWasmClient> {
    return await SigningCosmWasmClient.connectWithSigner(process.env.NODE_URL as string, await getWallet(privateKey), getSignerOptions());
}

export async function getValidatorWallet(): Promise<DirectSecp256k1Wallet> {
    return await getWallet(validatorPrivKey);
}

export async function getValidatorClient(): Promise<SigningCosmWasmClient> {
    return await getClient(validatorPrivKey);
}

export async function getPeriodicWallet(): Promise<DirectSecp256k1Wallet> {
    return await getWallet(periodicPrivKey);
}

export async function getPeriodicClient(): Promise<SigningCosmWasmClient> {
    return await getClient(periodicPrivKey);
}

export async function getUser2Wallet(): Promise<DirectSecp256k1Wallet> {
    return await getWallet(user2PrivKey);
}

export async function getUser2Client(): Promise<SigningCosmWasmClient> {
    return await getClient(user2PrivKey);
}

export async function getUser1Wallet(): Promise<DirectSecp256k1Wallet> {
    return await getWallet(user1PrivKey);
}

export async function getUser1Client(): Promise<SigningCosmWasmClient> {
    return await getClient(user1PrivKey);
}


export async function getDelayedVestingWallet(): Promise<DirectSecp256k1Wallet> {
    return await getWallet(delayedVestingPrivKey);
}

export async function getDelayedVestingClient(): Promise<SigningCosmWasmClient> {
    return await getClient(delayedVestingPrivKey);
}

function getSignerOptions(): SigningCosmWasmClientOptions {
    // @ts-ignore
    const customRegistry = new Registry([
        ...defaultRegistryTypes,
        ["/cosmwasm.wasm.v1.MsgClearAdmin", MsgClearAdmin],
        ["/cosmwasm.wasm.v1.MsgExecuteContract", MsgExecuteContract],
        ["/cosmwasm.wasm.v1.MsgMigrateContract", MsgMigrateContract],
        ["/cosmwasm.wasm.v1.MsgStoreCode", MsgStoreCode],
        ["/cosmwasm.wasm.v1.MsgInstantiateContract", MsgInstantiateContract],
        ["/cosmwasm.wasm.v1.MsgUpdateAdmin", MsgUpdateAdmin],
        [`/${vestingPackage}.MsgCreateVestingAccount`, MsgCreateVestingAccount],
    ]);
    return {registry: customRegistry}
}

