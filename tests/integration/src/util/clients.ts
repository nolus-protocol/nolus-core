import {SigningCosmWasmClient} from "@cosmjs/cosmwasm-stargate";
import {DirectSecp256k1Wallet} from "@cosmjs/proto-signing";
import {fromHex} from "@cosmjs/encoding";

let validatorWallet: DirectSecp256k1Wallet;
let validatorClient: SigningCosmWasmClient;
let user1Wallet: DirectSecp256k1Wallet;
let user1Client: SigningCosmWasmClient;
let user2Wallet: DirectSecp256k1Wallet;
let user2Client: SigningCosmWasmClient;

export async function getValidatorWallet(): Promise<DirectSecp256k1Wallet> {
    if (!validatorWallet) {
        validatorWallet = await DirectSecp256k1Wallet.fromKey(fromHex(process.env.VALIDATOR_PRIV_KEY as string), "nomo");
    }
    return validatorWallet;
}

export async function getValidatorClient(): Promise<SigningCosmWasmClient> {
    if (!validatorClient) {
        validatorClient = await SigningCosmWasmClient.connectWithSigner(process.env.NODE_URL as string, await getValidatorWallet());
    }
    return validatorClient;
}

export async function getUser1Wallet(): Promise<DirectSecp256k1Wallet> {
    if (!user1Wallet) {
        user1Wallet = await DirectSecp256k1Wallet.fromKey(fromHex(process.env.USR_1_PRIV_KEY as string), "nomo");
    }
    return user1Wallet;
}

export async function getUser1Client(): Promise<SigningCosmWasmClient> {
    if (!user1Client) {
        user1Client = await SigningCosmWasmClient.connectWithSigner(process.env.NODE_URL as string, await getUser1Wallet());
    }
    return user1Client;
}

export async function getUser2Wallet(): Promise<DirectSecp256k1Wallet> {
    if (!user2Wallet) {
        user2Wallet = await DirectSecp256k1Wallet.fromKey(fromHex(process.env.USR_2_PRIV_KEY as string), "nomo");
    }
    return user2Wallet;
}

export async function getUser2Client(): Promise<SigningCosmWasmClient> {
    if (!user2Client) {
        user2Client = await SigningCosmWasmClient.connectWithSigner(process.env.NODE_URL as string, await getUser2Wallet());
    }
    return user2Client;
}
