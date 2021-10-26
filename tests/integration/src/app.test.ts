import { CosmWasmClient } from "@cosmjs/cosmwasm-stargate";

test('blockchain is running', async () => {
    const client = await CosmWasmClient.connect(process.env.NODE_URL as string)
    // Query chain ID
    const chainId = await client.getChainId()
    // Query chain height
    const height = await client.getHeight()

    expect(chainId).toBeDefined()
    expect(height).toBeGreaterThan(0)
})