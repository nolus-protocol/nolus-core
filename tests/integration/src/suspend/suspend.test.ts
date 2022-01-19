import {MsgSuspend, MsgUnsuspend, protobufPackage} from "../util/codec/nolus/suspend/v1beta1/tx";
import {SigningCosmWasmClient} from "@cosmjs/cosmwasm-stargate";
import {
    createWallet,
    getClient,
    getSuspendAdminClient,
    getSuspendAdminWallet,
    getUser1Wallet,
    getValidatorClient,
    getValidatorWallet
} from "../util/clients";
import Long from "long";
import {EncodeObject} from "@cosmjs/proto-signing/build/registry";
import {getSuspendQueryClient} from "./suspend-client";
import {Query} from "../util/codec/nolus/suspend/v1beta1/query";
import {AccountData} from "@cosmjs/proto-signing";
import {isBroadcastTxFailure} from "@cosmjs/stargate";
import {DEFAULT_FEE, TEN_NOLUS} from "../util/utils";


describe("suspend module", () => {
    const DUMMY_TRANSFER_MSG = [{denom: "unolus", "amount": "1"}];
    let suspendAdminClient: SigningCosmWasmClient;
    let suspendAdminAccount: AccountData
    let genUserClient: SigningCosmWasmClient;
    let genUserAccount: AccountData
    let suspendQueryClient: Query

    beforeAll(async () => {
        suspendAdminClient = await getSuspendAdminClient();
        [suspendAdminAccount] = await (await getSuspendAdminWallet()).getAccounts();
        let genWallet = await createWallet();
        [genUserAccount] = await genWallet.getAccounts()
        genUserClient = await getClient(genWallet);
        suspendQueryClient = await getSuspendQueryClient(process.env.NODE_URL as string);

        // create & fund account
        const validatorClient = await getValidatorClient()
        const [validatorAccount] = await (await getValidatorWallet()).getAccounts()
        await validatorClient.sendTokens(validatorAccount.address, genUserAccount.address, TEN_NOLUS, DEFAULT_FEE);
        await validatorClient.sendTokens(validatorAccount.address, suspendAdminAccount.address, TEN_NOLUS, DEFAULT_FEE);
    })

    afterEach(async () => {
        const suspendedMsg = asMsgUnsuspend(suspendAdminAccount);
        await suspendAdminClient.signAndBroadcast(suspendAdminAccount.address, [suspendedMsg], DEFAULT_FEE);
    })

    describe("given admin account signer", () => {
        test("suspended state can be enabled and then disabled", async () => {


            const suspendedMsg = asMsgSuspend(suspendAdminAccount, 1);
            await suspendAdminClient.signAndBroadcast(suspendAdminAccount.address, [suspendedMsg], DEFAULT_FEE);

            let suspendResponse = await suspendQueryClient.SuspendedState({});
            expect(suspendResponse.state).toBeDefined();
            expect(suspendResponse.state?.suspended).toBeTruthy();
            expect(suspendResponse.state?.blockHeight).toEqual(Long.fromNumber(1));

            const unsuspendedMsg = asMsgUnsuspend(suspendAdminAccount);
            await suspendAdminClient.signAndBroadcast(suspendAdminAccount.address, [unsuspendedMsg], DEFAULT_FEE);

            suspendResponse = await suspendQueryClient.SuspendedState({});
            expect(suspendResponse.state).toBeDefined();
            expect(suspendResponse.state?.suspended).toBeFalsy();
        })

        test("when suspended and height reached then other transactions are unauthorized", async () => {
            let currentHeight = await suspendAdminClient.getHeight();
            const suspendedMsg = asMsgSuspend(suspendAdminAccount, currentHeight + 1);
            await suspendAdminClient.signAndBroadcast(suspendAdminAccount.address, [suspendedMsg], DEFAULT_FEE);

            const broadcast = () => suspendAdminClient.sendTokens(suspendAdminAccount.address, genUserAccount.address, DUMMY_TRANSFER_MSG, DEFAULT_FEE)
            await expect(broadcast).rejects.toThrow(/^.*unauthorized: node is suspended/)
        })

        test("when suspended but height is not reached then other transactions are authorized", async () => {
            let currentHeight = await suspendAdminClient.getHeight();
            const suspendedMsg = asMsgSuspend(suspendAdminAccount, currentHeight + 1000);
            await suspendAdminClient.signAndBroadcast(suspendAdminAccount.address, [suspendedMsg], DEFAULT_FEE);

            const result = await suspendAdminClient.sendTokens(suspendAdminAccount.address, genUserAccount.address, DUMMY_TRANSFER_MSG, DEFAULT_FEE)
            expect(isBroadcastTxFailure(result)).toBeFalsy()
        })

        test("when suspended can send multiple messages when one of them is MsgUnsuspend", async () => {
            let currentHeight = await suspendAdminClient.getHeight();
            const suspendedMsg = asMsgSuspend(suspendAdminAccount, currentHeight);
            await suspendAdminClient.signAndBroadcast(suspendAdminAccount.address, [suspendedMsg], DEFAULT_FEE);

            const unsuspendMsg = asMsgUnsuspend(suspendAdminAccount);
            const [user1Account] = await (await getUser1Wallet()).getAccounts()

            const fullTransferMsg1 = dummyMsgTransfer(suspendAdminAccount, genUserAccount);
            const fullTransferMsg2 = dummyMsgTransfer(suspendAdminAccount, user1Account);

            const result = await suspendAdminClient.signAndBroadcast(suspendAdminAccount.address, [fullTransferMsg1, unsuspendMsg, fullTransferMsg2], DEFAULT_FEE);
            expect(isBroadcastTxFailure(result)).toBeFalsy();

            const suspendResponse = await suspendQueryClient.SuspendedState({});
            expect(suspendResponse.state?.suspended).toBeFalsy();
        })
    });

    describe("given non admin account signer", () => {
        test("state suspended cannot be changed", async () => {
            // ensure state cannot be modified while being in state suspended
            const suspendedMsg = asMsgSuspend(suspendAdminAccount, 0);
            await suspendAdminClient.signAndBroadcast(suspendAdminAccount.address, [suspendedMsg], DEFAULT_FEE);

            let invalidMsg = asMsgUnsuspend(genUserAccount);
            let result = await genUserClient.signAndBroadcast(genUserAccount.address, [invalidMsg], DEFAULT_FEE)
            expect(isBroadcastTxFailure(result)).toBeTruthy();

            // ensure state cannot be modified while also being in state unsuspended
            const unsuspendedMsg = asMsgUnsuspend(suspendAdminAccount);
            await suspendAdminClient.signAndBroadcast(suspendAdminAccount.address, [unsuspendedMsg], DEFAULT_FEE);


            invalidMsg = asMsgUnsuspend(genUserAccount);
            result = await genUserClient.signAndBroadcast(genUserAccount.address, [invalidMsg], DEFAULT_FEE)
            expect(isBroadcastTxFailure(result)).toBeTruthy();
        })

        test("state height cannot be changed", async () => {
            let invalidMsg = asMsgSuspend(genUserAccount, 100);
            let result = await genUserClient.signAndBroadcast(genUserAccount.address, [invalidMsg], DEFAULT_FEE)
            expect(isBroadcastTxFailure(result)).toBeTruthy();
        })

        test("message cannot be send with forged from address", async () => {
            let invalidMsg = asMsgSuspend(suspendAdminAccount, 100);
            let broadcast = () => genUserClient.signAndBroadcast(genUserAccount.address, [invalidMsg], DEFAULT_FEE);
            await expect(broadcast).rejects.toThrow(/^Broadcasting transaction failed with code 8*/)
        })

        test("when suspended and height reached then other transactions are unauthorized", async () => {
            const suspendedMsg = asMsgSuspend(suspendAdminAccount, 0);
            await suspendAdminClient.signAndBroadcast(suspendAdminAccount.address, [suspendedMsg], DEFAULT_FEE);

            const broadcast = () => genUserClient.sendTokens(genUserAccount.address, suspendAdminAccount.address, DUMMY_TRANSFER_MSG, DEFAULT_FEE)
            await expect(broadcast).rejects.toThrow(/^.*unauthorized: node is suspended$/)
        })

        test("when suspended cannot bypass it by sending multiple messages including MsgSuspend and MsgUnsuspend", async () => {
            const suspendedMsg = asMsgSuspend(suspendAdminAccount, 0);
            await suspendAdminClient.signAndBroadcast(suspendAdminAccount.address, [suspendedMsg], DEFAULT_FEE);

            const unsuspendMsg = asMsgUnsuspend(genUserAccount);
            const [user1Account] = await (await getUser1Wallet()).getAccounts()

            const fullTransferMsg1 = dummyMsgTransfer(genUserAccount, suspendAdminAccount)
            const fullTransferMsg2 = dummyMsgTransfer(genUserAccount, user1Account)

            const result = await genUserClient.signAndBroadcast(genUserAccount.address, [fullTransferMsg1, unsuspendMsg, fullTransferMsg2], DEFAULT_FEE);
            console.log(result)
            expect(isBroadcastTxFailure(result)).toBeTruthy();
        })
    })

    function asMsgSuspend(fromAddress: AccountData, blockHeight: number): EncodeObject {
        const suspendMsg: MsgSuspend = {
            fromAddress: fromAddress.address,
            blockHeight: Long.fromNumber(blockHeight)
        }
        return {
            typeUrl: `/${protobufPackage}.MsgSuspend`,
            value: suspendMsg,
        };
    }

    function asMsgUnsuspend(fromAddress: AccountData): EncodeObject {
        const suspendMsg: MsgUnsuspend = {
            fromAddress: fromAddress.address
        }
        return {
            typeUrl: `/${protobufPackage}.MsgUnsuspend`,
            value: suspendMsg,
        };
    }

    function dummyMsgTransfer(fromAddress: AccountData, toAddress: AccountData): EncodeObject {
        return {
            typeUrl: "/cosmos.bank.v1beta1.MsgSend",
            value: {
                fromAddress: fromAddress.address,
                toAddress: toAddress.address,
                amount: DUMMY_TRANSFER_MSG
            }
        };
    }
})
