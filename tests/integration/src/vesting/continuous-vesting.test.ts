import {createWallet, DEFAULT_FEE, getClient, getValidatorClient, getValidatorWallet} from "../util/clients";
import {AccountData, EncodeObject} from "@cosmjs/proto-signing";
import {SigningCosmWasmClient} from "@cosmjs/cosmwasm-stargate";
import {MsgCreateVestingAccount, protobufPackage as vestingPackage} from "../util/codec/cosmos/vesting/v1beta1/tx";

import Long from "long";
import {assertIsBroadcastTxSuccess, isBroadcastTxFailure} from "@cosmjs/stargate";
import {sleep} from "../util/utils";
import {Coin} from "../util/codec/cosmos/base/v1beta1/coin";

describe("continuous vesting", () => {
    let validatorClient: SigningCosmWasmClient;
    let validatorAccount: AccountData;
    let continuousClient: SigningCosmWasmClient;
    let continuousAccount: AccountData;

    beforeAll(async () => {
        validatorClient = await getValidatorClient();
        [validatorAccount] = await (await getValidatorWallet()).getAccounts();
        const contWallet = await createWallet();
        continuousClient = await getClient(contWallet);
        [continuousAccount] = await contWallet.getAccounts();
    })

    test("created continuous vesting account works as expected", async () => {
        const FULL_AMOUNT: Coin = {denom: "unolus", amount: "10000"};
        const HALF_AMOUNT: Coin = {denom: "unolus", amount: "5000"};

        const createVestingAccountMsg: MsgCreateVestingAccount = {
            fromAddress: validatorAccount.address,
            toAddress: continuousAccount.address,
            amount: [FULL_AMOUNT],
            endTime: Long.fromNumber((new Date().getTime() / 1000) + 7), // 7 seconds
            delayed: false,
        }
        const encodedMsg: EncodeObject = {
            typeUrl: `/${vestingPackage}.MsgCreateVestingAccount`,
            value: createVestingAccountMsg,
        }
        let result = await validatorClient.signAndBroadcast(validatorAccount.address, [encodedMsg], DEFAULT_FEE)
        assertIsBroadcastTxSuccess(result)

        let sendFailTx = await continuousClient.sendTokens(continuousAccount.address, validatorAccount.address, [HALF_AMOUNT], DEFAULT_FEE);
        expect(isBroadcastTxFailure(sendFailTx)).toBeTruthy()
        await expect(sendFailTx.rawLog).toMatch(/^.*smaller than 5000unolus: insufficient funds.*/)
        await sleep(4000) // sleep for 4000 seconds
        assertIsBroadcastTxSuccess(await continuousClient.sendTokens(continuousAccount.address, validatorAccount.address, [HALF_AMOUNT], DEFAULT_FEE))
    })

})
