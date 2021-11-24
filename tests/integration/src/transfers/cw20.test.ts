import * as fs from "fs";
import { SigningCosmWasmClient, InstantiateResult } from "@cosmjs/cosmwasm-stargate";
import { DirectSecp256k1Wallet } from "@cosmjs/proto-signing";
import { getValidatorClient, getValidatorWallet, getUser1Wallet, getUser1Client } from "../util/clients";
import { AccountData } from "@cosmjs/amino";

const customFees = {
    upload: {
        amount: [{ amount: "2000000", denom: "nolus" }],
        gas: "2000000",
    },
    init: {
        amount: [{ amount: "500000", denom: "nolus" }],
        gas: "500000",
    },
    exec: {
        amount: [{ amount: "500000", denom: "nolus" }],
        gas: "500000",
    }
};

describe("CW20 transfers", () => {
    let validatorClient: SigningCosmWasmClient;
    let validatorAccount: AccountData;
    let contractAddress: string;
    let tokenName = "Test";
    let tokenSymbol = "TST";
    let tokenDecimals = 18;
    let totalSupply = "1000000000000000000";

    beforeEach(async() =>{
        validatorClient = await getValidatorClient();
        [validatorAccount] = await (await getValidatorWallet()).getAccounts();

        // get wasm binary file
        const wasmBinary: Buffer = fs.readFileSync("./wasm-contracts/cw20_base.wasm");

        // upload wasm binary
        const uploadReceipt = await validatorClient.upload(validatorAccount.address, wasmBinary, customFees.upload);
        const codeId = uploadReceipt.codeId;
        console.log("uploadReceipt:", uploadReceipt);

        // instantiate the contract
        const instatiateMsg = {
            "name": tokenName,
            "symbol": tokenSymbol,
            "decimals": tokenDecimals,
            "initial_balances": [
                {
                    "address": validatorAccount.address,
                    "amount": totalSupply
                }
            ]
        };
        const contract: InstantiateResult = await validatorClient.instantiate(validatorAccount.address, codeId, instatiateMsg, "Sample CW20", customFees.init);
        contractAddress = contract.contractAddress;
        console.log("contract address:", contractAddress);

        // get token info
        const tokenInfoMsg = {
            "token_info": {}
        };
        const tokenInfoResponse = await validatorClient.queryContractSmart(contractAddress, tokenInfoMsg);
        console.log("token_info: ", tokenInfoResponse);

        expect(tokenInfoResponse.name).toBe(tokenName);
        expect(tokenInfoResponse.symbol).toBe(tokenSymbol);
        expect(tokenInfoResponse.decimals).toBe(tokenDecimals);
        expect(tokenInfoResponse["total_supply"]).toBe(totalSupply);

        // get validator balance
        const balanceMsg = {
            "balance": {
                "address": validatorAccount.address
            }
        };
        const validatorBalanceMsgResponse = await validatorClient.queryContractSmart(contractAddress, balanceMsg);
        console.log("Validator balance:", validatorBalanceMsgResponse);

        expect(validatorBalanceMsgResponse.balance).toBe(totalSupply);
    });

    test("Users can transfer tokens", async () => {
        const user1Client = await getUser1Client();
        const [user1Account] = await (await getUser1Wallet()).getAccounts();
        let user1BalanceBefore;
        let user1BalanceAfter;
        let amountToTransfer = "1000";

        const balanceMsg = {
            "balance": {
                "address": user1Account.address
            }
        };
        user1BalanceBefore = (await user1Client.queryContractSmart(contractAddress, balanceMsg)).balance;
        console.log("User1 before balance:", user1BalanceBefore);

        const transferMsg = {
            "transfer": {
                "recipient": user1Account.address,
                "amount": amountToTransfer,
            }
        };
        await validatorClient.execute(validatorAccount.address, contractAddress, transferMsg, customFees.exec);

        user1BalanceAfter = (await user1Client.queryContractSmart(contractAddress, balanceMsg)).balance;
        console.log("User1 after balance:", user1BalanceAfter);

        expect(BigInt(user1BalanceAfter)).toBe(BigInt(user1BalanceBefore) + BigInt(amountToTransfer));
        expect(BigInt(user1BalanceAfter)).toBe(BigInt(user1BalanceBefore) + BigInt(amountToTransfer));
    });

    test("Users can transfer tokens allowed from another user", async () => {
        const user1Client = await getUser1Client();
        const [user1Account] = await (await getUser1Wallet()).getAccounts();
        let user1AllowanceBefore;
        let user1AllowanceAfter;
        let user1BalanceBefore;
        let user1BalanceAfter;
        let amountToTransfer = "1000";

        const allowanceMsg = {
            "allowance": {
                "owner": validatorAccount.address,
                "spender": user1Account.address
            }
        };
        user1AllowanceBefore = (await user1Client.queryContractSmart(contractAddress, allowanceMsg)).allowance;
        console.log("User before allowance:", user1AllowanceBefore);

        const balanceMsg = {
            "balance": {
                "address": user1Account.address
            }
        };
        user1BalanceBefore = (await user1Client.queryContractSmart(contractAddress, balanceMsg)).balance;
        console.log("User before balance:", user1BalanceBefore);

        // send some native tokens to the user, so that they can call TransferFrom
        const nativeTokenTransfer = {
            denom: "nolus",
            amount: "2000000",
        };
        const fee = {
            amount: [{denom: "nolus", amount: "12"}],
            gas: "100000"
        };
        await validatorClient.sendTokens(validatorAccount.address, user1Account.address, [nativeTokenTransfer], fee, "Send transaction");

        const increaseAllowanceMsg = {
            "increase_allowance": {
                "spender": user1Account.address,
                "amount": amountToTransfer,
            }
        };
        await validatorClient.execute(validatorAccount.address, contractAddress, increaseAllowanceMsg, customFees.exec);

        user1AllowanceAfter = (await user1Client.queryContractSmart(contractAddress, allowanceMsg)).allowance;
        console.log("User after allowance:", user1AllowanceAfter);

        expect(BigInt(user1AllowanceAfter)).toBe(BigInt(user1AllowanceBefore) + BigInt(amountToTransfer));

        const transferFromMsg = {
            "transfer_from": {
                "owner": validatorAccount.address,
                "recipient": user1Account.address,
                "amount": amountToTransfer
            }
        };
        await user1Client.execute(user1Account.address, contractAddress, transferFromMsg, customFees.exec);

        user1BalanceAfter = (await user1Client.queryContractSmart(contractAddress, balanceMsg)).balance;
        console.log("User after balance:", user1BalanceAfter);
        console.log("User after transfer allowance:", (await user1Client.queryContractSmart(contractAddress, allowanceMsg)).allowance);

        expect(BigInt(user1BalanceAfter)).toBe(BigInt(user1BalanceBefore) + BigInt(amountToTransfer));
    }, 30000);
});
