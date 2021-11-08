import {SigningCosmWasmClient} from "@cosmjs/cosmwasm-stargate";
import {DirectSecp256k1Wallet} from "@cosmjs/proto-signing";
import {fromHex} from "@cosmjs/encoding";

let validatorWallet: DirectSecp256k1Wallet;
let validatorClient: SigningCosmWasmClient;

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