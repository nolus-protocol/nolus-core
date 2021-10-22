import {DirectSecp256k1Wallet} from "@cosmjs/proto-signing";
import {fromHex} from "@cosmjs/encoding";
import {SigningCosmWasmClient} from "@cosmjs/cosmwasm-stargate";
import * as fs from "fs";
import {getValidatorClient, getValidatorWallet} from "./util/clients";

const customFees = {
    upload: {
        amount: [{ amount: "2000000", denom: "nomo" }],
        gas: "2000000",
    },
    init: {
        amount: [{ amount: "500000", denom: "nomo" }],
        gas: "500000",
    },
    exec: {
        amount: [{ amount: "500000", denom: "nomo" }],
        gas: "500000",
    },
    send: {
        amount: [{ amount: "80000", denom: "nomo" }],
        gas: "80000",
    },
}


test('sample wasm contract successfully get deployed and runs on the blockchain', async () => {
    const wallet = await getValidatorWallet()
    const client = await getValidatorClient();
    const [firstAccount] = await wallet.getAccounts();

    const wasm = fs.readFileSync("./wasm-contracts/cw_escrow.wasm");
    console.log('Uploading sample contract')
    const uploadReceipt = await client.upload(firstAccount.address, wasm, customFees.upload);
    console.log('uploadReceipt :', uploadReceipt);
    const codeId = uploadReceipt.codeId;

    const initMsg = { "arbiter": firstAccount.address,  "recipient": process.env.USR_1_ADDR }

    const funds = [{ amount: "5000", denom: "nomo" }]
    const contract = await client.instantiate(firstAccount.address, codeId, initMsg, "Sample escrow " + Math.ceil(Math.random()*10000), customFees.init, { funds });
    console.log('contract: ', contract);

    const contractAddress = contract.contractAddress;

    const previousContractBalance = await client.getBalance(contractAddress, 'nomo');
    expect(BigInt(previousContractBalance.amount)).toBe(BigInt(funds[0].amount))
    const previousReceiverBalance = await client.getBalance(process.env.USR_1_ADDR as string, 'nomo');

// Approve the escrow
    const handleMsg = { "approve":{"quantity":[{"amount":"5000","denom":"nomo"}]}};
    console.log('Approving escrow');
    const response = await client.execute(firstAccount.address, contractAddress, handleMsg, customFees.exec);

// Query again to confirm it worked
    console.log('Querying again contract for the updated balance');
    const actualContractBalance = await client.getBalance(contractAddress, 'nomo');
    const actualReceiverBalance = await client.getBalance(process.env.USR_1_ADDR as string, 'nomo');

    expect(BigInt(actualContractBalance.amount)).toBe(BigInt(0))
    expect(BigInt(actualReceiverBalance.amount)).toBe(BigInt(previousReceiverBalance.amount) + BigInt(5000))
}, 30000)