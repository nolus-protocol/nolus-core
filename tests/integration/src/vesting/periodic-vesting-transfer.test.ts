import {SigningCosmWasmClient} from "@cosmjs/cosmwasm-stargate";
import {assertIsDeliverTxSuccess, Coin, isDeliverTxFailure} from "@cosmjs/stargate";
import {getPeriodicClient, getPeriodicWallet, getValidatorWallet} from "../util/clients";
import {AccountData} from "@cosmjs/proto-signing";
import {DEFAULT_FEE} from "../util/utils";

describe('periodic vesting transfers', () => {
    const VESTED_AMOUNT: Coin = {denom: "unolus", amount: "118000"}; // + 63 remainder, if needed for taxes
    let periodicAccount: AccountData;
    let validatorAccount: AccountData;
    let periodicClient: SigningCosmWasmClient;

    beforeEach(async () => {
        [periodicAccount] = await (await getPeriodicWallet()).getAccounts();
        [validatorAccount] = await (await getValidatorWallet()).getAccounts();
        periodicClient = await getPeriodicClient();
    })

    test('periodic vesting account has positive balance', async () => {
        const balance = await periodicClient.getBalance(periodicAccount.address, "unolus");
        expect(BigInt(balance.amount) > 0).toBeTruthy()
    })

    test('periodic account\'s vested amount can be send', async () => {
        let result = await periodicClient.sendTokens(periodicAccount.address, validatorAccount.address, [VESTED_AMOUNT], DEFAULT_FEE)
        assertIsDeliverTxSuccess(result)
    })

    test('periodic account\s vesting amount cannot be send', async () => {
        let result = await periodicClient.sendTokens(periodicAccount.address, validatorAccount.address, [VESTED_AMOUNT], DEFAULT_FEE)
        expect(isDeliverTxFailure(result)).toBeTruthy();
    })
})
