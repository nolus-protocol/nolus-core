import { CosmWasmClient } from "@cosmjs/cosmwasm-stargate";
import { assertIsBroadcastTxSuccess } from "@cosmjs/stargate";
import { getValidatorClient, getValidatorWallet } from "../util/clients";

describe('native transfers', () => {

    test('validator has positive balance', async () => {
        const client = await CosmWasmClient.connect(process.env.NODE_URL as string);
        const balance = await client.getBalance(process.env.VALIDATOR_ADDR as string, "nomo");

        console.log(`Validator balance=(denom=${balance.denom}, amount=${balance.amount})`);

        expect(BigInt(balance.amount) > 0).toBeTruthy();
    });

    test('validator can send tokens', async () => {
        const wallet = await getValidatorWallet();
        const client = await getValidatorClient();
        const [firstAccount] = await wallet.getAccounts();
        const amount = {
            denom: "nomo",
            amount: "1234",
        };
        const feeAmount = "12";
        const fee = {
            amount: [{denom: "nomo", amount: feeAmount}],
            gas: "100000"
        };
        const previousValidatorBalance = await client.getBalance(firstAccount.address as string, "nomo");
        const previousUsrBalance = await client.getBalance(process.env.USR_1_ADDR as string, "nomo");
        const result = await client.sendTokens(firstAccount.address, process.env.USR_1_ADDR as string, [amount], fee, "Testing send transaction");
        const nextValidatorBalance = await client.getBalance(firstAccount.address as string, "nomo");
        const nextUsrBalance = await client.getBalance(process.env.USR_1_ADDR as string, "nomo");

        console.log(`Validator balance before=(${previousValidatorBalance.denom}, ${previousValidatorBalance.amount})`);
        console.log(`Validator balance after=(${nextValidatorBalance.denom}, ${nextValidatorBalance.amount})`);
        console.log(`User balance before=(${previousUsrBalance.denom}, ${previousUsrBalance.amount})`);
        console.log(`User balance after=(${nextUsrBalance.denom}, ${nextUsrBalance.amount})`);

        assertIsBroadcastTxSuccess(result);
        expect(BigInt(nextUsrBalance.amount)).toBe(BigInt(previousUsrBalance.amount) + BigInt(amount.amount));
        expect(BigInt(nextValidatorBalance.amount)).toBe(BigInt(previousValidatorBalance.amount) - BigInt(amount.amount) - BigInt(feeAmount));
    });
});
