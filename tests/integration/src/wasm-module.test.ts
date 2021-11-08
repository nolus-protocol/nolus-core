import * as fs from "fs";
import { getValidatorClient, getValidatorWallet } from "./util/clients";

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
};

test('sample wasm contract successfully get deployed and runs on the blockchain', async () => {
    const wallet = await getValidatorWallet();
    const client = await getValidatorClient();
    const [firstAccount] = await wallet.getAccounts();

    const wasm = fs.readFileSync("./wasm-contracts/cw_escrow.wasm");
    const previousValidatorBalance = await client.getBalance(firstAccount.address as string, "nomo");

    console.log('Uploading sample contract');
    const uploadReceipt = await client.upload(firstAccount.address, wasm, customFees.upload);
    console.log('uploadReceipt:', uploadReceipt);
    const codeId = uploadReceipt.codeId;

    const initMsg = {
        "arbiter": firstAccount.address,
        "recipient": process.env.USR_1_ADDR
    };
    const funds = [{ amount: "5000", denom: "nomo" }];
    const contract = await client.instantiate(firstAccount.address, codeId, initMsg, "Sample escrow " + Math.ceil(Math.random() * 10000), customFees.init, { funds });
    console.log('contract: ', contract);

    const contractAddress = contract.contractAddress;

    const previousContractBalance = await client.getBalance(contractAddress, 'nomo');
    expect(BigInt(previousContractBalance.amount)).toBe(BigInt(funds[0].amount));
    const previousReceiverBalance = await client.getBalance(process.env.USR_1_ADDR as string, 'nomo');

    // Approve the escrow
    console.log('Approving escrow');
    const handleMsg = { "approve": { "quantity": [{ "amount": "5000", "denom": "nomo" }] } };
    const response = await client.execute(firstAccount.address, contractAddress, handleMsg, customFees.exec);

    // Query again to confirm it worked
    console.log('Querying again contract for the updated balance');
    const actualValidatorBalance = await client.getBalance(firstAccount.address as string, "nomo");
    const actualContractBalance = await client.getBalance(contractAddress, 'nomo');
    const actualReceiverBalance = await client.getBalance(process.env.USR_1_ADDR as string, 'nomo');

    console.log(`Validator balance before=(${previousValidatorBalance.denom}, ${previousValidatorBalance.amount})`);
    console.log(`Validator balance after=(${actualValidatorBalance.denom}, ${actualValidatorBalance.amount})`);

    expect(BigInt(actualValidatorBalance.amount)).toBe(BigInt(previousValidatorBalance.amount) - BigInt(customFees.upload.amount[0].amount) - BigInt(customFees.init.amount[0].amount) - BigInt(funds[0].amount) - BigInt(customFees.exec.amount[0].amount));
    expect(BigInt(actualContractBalance.amount)).toBe(BigInt(0));
    expect(BigInt(actualReceiverBalance.amount)).toBe(BigInt(previousReceiverBalance.amount) + BigInt(5000));
}, 30000);