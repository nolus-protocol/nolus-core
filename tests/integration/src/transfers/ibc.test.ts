import { SigningCosmWasmClient } from "@cosmjs/cosmwasm-stargate";
import { assertIsDeliverTxSuccess, DeliverTxResponse } from "@cosmjs/stargate";
import { getValidatorClient, getValidatorWallet, getUser1Wallet, getUser1Client, getUser2Wallet, getUser2Client } from "../util/clients";
import { AccountData } from "@cosmjs/amino";

describe("IBC transfers", () => {
    test("test", async () => {
        const validatorClient: SigningCosmWasmClient = await getValidatorClient();
        const [validatorAccount] = await (await getValidatorWallet()).getAccounts();
        const user1Client: SigningCosmWasmClient = await getUser1Client();
        const [user1Account] = await (await getUser1Wallet()).getAccounts();

        let ibcToken = process.env.IBC_TOKEN as string;
        let transferAmount = "1000";

        let initialValidatorBalance = await validatorClient.getBalance(validatorAccount.address, ibcToken);
        let initialUser1Balance = await user1Client.getBalance(user1Account.address, ibcToken);

        console.log("Validator before balance:", initialValidatorBalance);
        console.log("User before balance:", initialUser1Balance);

        expect(ibcToken).toBeDefined();
        expect(ibcToken.length > 0).toBeTruthy();
        expect(BigInt(initialValidatorBalance.amount) > 0).toBeTruthy();

        const transfer = {
            denom: ibcToken,
            amount: transferAmount,
        };
        const fee = {
            amount: [{denom: "unolus", amount: "12"}],
            gas: "100000"
        };
        let sendTokensResponse: DeliverTxResponse = await validatorClient.sendTokens(validatorAccount.address, user1Account.address, [transfer], fee, "Testing send transaction");
        assertIsDeliverTxSuccess(sendTokensResponse);

        let nextValidatorBalance = await validatorClient.getBalance(validatorAccount.address, ibcToken);
        let nextUser1Balance = await user1Client.getBalance(user1Account.address, ibcToken);

        console.log("Validator after balance:", nextValidatorBalance);
        console.log("User after balance:", nextUser1Balance);

        expect(BigInt(nextValidatorBalance.amount)).toBe(BigInt(initialValidatorBalance.amount) - BigInt(transferAmount));
        expect(BigInt(nextUser1Balance.amount)).toBe(BigInt(initialUser1Balance.amount) + BigInt(transferAmount));
    });
});
