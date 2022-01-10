export const DEFAULT_FEE = {
    amount: [{denom: "unolus", amount: "12"}],
    gas: "100000"
};

export async function sleep(ms: number): Promise<void> {
    await new Promise(r => setTimeout(r, ms));
}