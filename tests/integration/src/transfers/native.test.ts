import { CosmWasmClient } from "@cosmjs/cosmwasm-stargate";
import { SigningCosmWasmClient } from "@cosmjs/cosmwasm-stargate";
import { assertIsBroadcastTxSuccess, BroadcastTxResponse, Coin } from "@cosmjs/stargate";
import { getValidatorClient, getValidatorWallet, getUser1Wallet, getUser2Wallet, getUser1Client } from "../util/clients";

describe("Native transfers", () => {
    test("Validator has positive balance", async () => {
        const client: CosmWasmClient = await CosmWasmClient.connect(process.env.NODE_URL as string);
        const balance: Coin = await client.getBalance(process.env.VALIDATOR_ADDR as string, "nomo");

        console.log(`Validator balance=(${balance.denom}, ${balance.amount})`);

        expect(BigInt(balance.amount) > 0).toBeTruthy();
    });

    test("Users can transfer tokens", async () => {
        const validatorClient: SigningCosmWasmClient = await getValidatorClient();
        const user1Client: SigningCosmWasmClient = await getUser1Client();
        const [validatorAccount] = await (await getValidatorWallet()).getAccounts();
        const [user1Account] = await (await getUser1Wallet()).getAccounts();
        const [user2Account] = await (await getUser2Wallet()).getAccounts();
        const transfer1 = {
            denom: "nomo",
            amount: "1234",
        };
        const transfer2 = {
            denom: "nomo",
            amount: "1000",
        };
        const fee = {
            amount: [{denom: "nomo", amount: "12"}],
            gas: "100000"
        };

        // validator -> user1
        let previousValidatorBalance: Coin = await validatorClient.getBalance(validatorAccount.address, "nomo");
        let previousUser1Balance: Coin = await validatorClient.getBalance(user1Account.address, "nomo");
        let broadcastTxResponse1: BroadcastTxResponse = await validatorClient.sendTokens(validatorAccount.address, user1Account.address, [transfer1], fee, "Testing send transaction");
        let nextValidatorBalance: Coin = await validatorClient.getBalance(validatorAccount.address, "nomo");
        let nextUser1Balance: Coin = await validatorClient.getBalance(user1Account.address, "nomo");

        console.log(`Validator balance before=(${previousValidatorBalance.denom}, ${previousValidatorBalance.amount}) after=(${nextValidatorBalance.denom}, ${nextValidatorBalance.amount})`);
        console.log(`User1 balance before=(${previousUser1Balance.denom}, ${previousUser1Balance.amount}) after=(${nextUser1Balance.denom}, ${nextUser1Balance.amount})`);

        assertIsBroadcastTxSuccess(broadcastTxResponse1);
        expect(BigInt(nextValidatorBalance.amount)).toBe(BigInt(previousValidatorBalance.amount) - BigInt(transfer1.amount) - BigInt(fee.amount[0].amount));
        expect(BigInt(nextUser1Balance.amount)).toBe(BigInt(previousUser1Balance.amount) + BigInt(transfer1.amount));

        // user1 -> user2
        previousUser1Balance = await validatorClient.getBalance(user1Account.address, "nomo");
        let previousUser2Balance: Coin = await validatorClient.getBalance(user2Account.address, "nomo");
        let broadcastTxResponse2: BroadcastTxResponse = await user1Client.sendTokens(user1Account.address, user2Account.address, [transfer2], fee, "Testing send transaction");
        nextUser1Balance = await validatorClient.getBalance(user1Account.address, "nomo");
        let nextUser2Balance: Coin = await validatorClient.getBalance(user2Account.address, "nomo");

        console.log(`User1 balance before=(${previousUser1Balance.denom}, ${previousUser1Balance.amount}) after=(${nextUser1Balance.denom}, ${nextUser1Balance.amount})`);
        console.log(`User2 balance before=(${previousUser2Balance.denom}, ${previousUser2Balance.amount}) after=(${nextUser2Balance.denom}, ${nextUser2Balance.amount})`);

        assertIsBroadcastTxSuccess(broadcastTxResponse2);
        expect(BigInt(nextUser1Balance.amount)).toBe(BigInt(previousUser1Balance.amount) - BigInt(transfer2.amount) - BigInt(fee.amount[0].amount));
        expect(BigInt(nextUser2Balance.amount)).toBe(BigInt(previousUser2Balance.amount) + BigInt(transfer2.amount));
    });
});
