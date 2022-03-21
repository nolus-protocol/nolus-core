import {Coin} from "./codec/cosmos/base/v1beta1/coin";

export const DEFAULT_FEE = {
    amount: [{denom: "unolus", amount: "8424"}],
    gas: "120000"
};

export const TEN_NOLUS: Coin[] = [{denom: "unolus", amount: "10_000_000"}]


export async function sleep(ms: number): Promise<void> {
    await new Promise(r => setTimeout(r, ms));
}
