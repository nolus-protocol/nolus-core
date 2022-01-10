import {
    getDelayedVestingClient,
    getDelayedVestingWallet,
    getValidatorClient,
    getValidatorWallet
} from "../util/clients";
import {AccountData, EncodeObject} from "@cosmjs/proto-signing";
import {SigningCosmWasmClient} from "@cosmjs/cosmwasm-stargate";
import {MsgCreateVestingAccount, protobufPackage as vestingPackage} from "../util/codec/cosmos/vesting/v1beta1/tx";

import Long from "long";
import {assertIsBroadcastTxSuccess} from "@cosmjs/stargate";
import {DEFAULT_FEE, sleep} from "../util/utils";

describe("delayed vesting", () => {
    let validatorClient: SigningCosmWasmClient;
    let validatorAccount: AccountData;
    let delayedClient: SigningCosmWasmClient;
    let delayedAccount: AccountData;

    beforeEach(async () => {
        validatorClient = await getValidatorClient();
        [validatorAccount] = await (await getValidatorWallet()).getAccounts();
        delayedClient = await getDelayedVestingClient();
        [delayedAccount] = await (await getDelayedVestingWallet()).getAccounts();
    })

    test("created delayed vesting account works as expected", async () => {
        const createVestingAccountMsg: MsgCreateVestingAccount = {
            fromAddress: validatorAccount.address,
            toAddress: delayedAccount.address,
            amount: [{denom: "unolus", amount: "1000"}],
            endTime: Long.fromNumber((new Date().getTime() / 1000) + 7),
            delayed: true,
        }
        const encodedMsg: EncodeObject = {
            typeUrl: `/${vestingPackage}.MsgCreateVestingAccount`,
            value: createVestingAccountMsg,
        }
        let result = await validatorClient.signAndBroadcast(validatorAccount.address, [encodedMsg], DEFAULT_FEE)
        assertIsBroadcastTxSuccess(result)

        let broadcast = () => delayedClient.sendTokens(delayedAccount.address, validatorAccount.address, DEFAULT_FEE.amount, DEFAULT_FEE)
        await expect(broadcast).rejects.toThrow(/^.*insufficient funds: insufficient funds.*/)
        await sleep(7000) // sleep for 7 seconds
        assertIsBroadcastTxSuccess(await delayedClient.sendTokens(delayedAccount.address, validatorAccount.address, DEFAULT_FEE.amount, DEFAULT_FEE))
    })

})
