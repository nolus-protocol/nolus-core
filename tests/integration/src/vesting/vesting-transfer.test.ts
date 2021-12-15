import {
    CosmWasmClient
} from "@cosmjs/cosmwasm-stargate";
import {
    assertIsBroadcastTxSuccess,
    isBroadcastTxFailure
} from "@cosmjs/stargate";
import {
    getPeriodicClient,
    getPeriodicWallet
} from "../util/clients";
describe('vesting transfers', () => {
    test('vesting account has positive balance', async () => {
        const client = await CosmWasmClient.connect(process.env.NODE_URL as string)
	const balance = await client.getBalance(process.env.PERIODIC_VEST_ADDR as string, "nolus");
        expect(BigInt(balance.amount) > 0).toBeTruthy()
    }) 

    test('vesting account can send tokens', async () => {
        const wallet = await getPeriodicWallet();
        const client = await getPeriodicClient();
        const [firstAccount] = await wallet.getAccounts();
        const amount = {
            denom: "nolus",
            amount: "1000",
        };
        const fee = {
            amount: [{
                denom: "nolus",
                amount: "12"
            }],
            gas: "100000"
        };
        const result = await client.sendTokens(firstAccount.address, process.env.USR_1_ADDR as string, [amount], fee, "Testing send transaction");
        expect(isBroadcastTxFailure(result)).toBeTruthy();
    })
})
