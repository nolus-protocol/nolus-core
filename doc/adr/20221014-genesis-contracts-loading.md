# [Initialization of nolus chain with appropriate wasm settings]

- Status: accepted
- Deciders: the Nolus dev team
- Date: 2022-10-14
- Tags: genesis, wasm, infrastructure, smart-contract

Technical Story: https://app.clickup.com/t/1yyn7rm

## Context and Problem Statement

### Context
In the genesis you can set global settings in the wasm module for who can upload and instantiate contracts. The options that we have for access are:
`Everybody`, `Nobody`, `OnlyAddress` (current wasmd version 0.27.0). If we upgrade the wasmd version, there will be also one more option - `OnlyAddresses`.
We can also set scoped instantiation permission for each code-upload(contract). This could conflict with the global settings, for example, if we have a contract with instantiation permissions set to X address, and in the global scope we have set instantiation to `OnlyAddress` with address Y, there will be conflict.
Each smart contract has an optional admin role. If a contract instance has an admin(address) set, then this admin could interact with the contract directly(for example invoking migrate-contract) rather than with gov proposals.
There is a `owner/privileged` contracts user that is only authorized to perform some administrative tasks, this user is set by the contracts to be the user who instantiated the contracts. 

### Problem statement
We want to run our chain as a permissioned one. We also want to upload and instantiate smart-contracts at the genesis.
The decision that most suited both of the above requirements was to have the Leaser contract's address as the only permissioned address which
can upload and instantiate contracts, where each instance will be without an admin.

## Decision Drivers

- If we were to have all of our contracts instantiated with an admin address, we would have to keep it's private key. And if anyone 
broke through and got the private key, he might have malicious intentions and damage our contracts. 
- Using the Leaser contract's address has limited options in terms of what one might do(can execute functions only defined in the contract).
- 

## Considered Options

- [option 1] Having messages for uploading and instantiating contracts in the wasm->gen_msgs subsection in the genesis. We would set the global permissions for uploading and instantiating contracts to only one address - the address of the Leaser contract.
- [option 2] Wasm section also has two subsections codes[] and contracts[] in the genesis. We considered having all of our contracts respectively uploaded and instantiated via appending them to the codes[] and contracts[] subsection. 
- [option 3] Starting the chain as permissionless then making a gov proposal to change to permissioned

## Decision Outcome

We decided to go with option-1 because it is the most straight forward and safe approach and we won't need to overengineer the genesis creation.
In the future if we need faster contracts execution, we can do code pinning [https://docs.cosmwasm.com/docs/1.0/smart-contracts/code-pinning/]
We can also have DAO governed smart contracts, adding another layer to voting [https://docs.cosmwasm.com/dev-academy/dao-governance/what-is-a-dao#dao-governed-smart-contracts]

If we consider option-2, we would have to add additional scripts for retrieving code hash.
Also we would need to copy contract-states for each instance(this option seems good for hard forks)

If we consider option-3, we would have a short window where everyone could be able to upload 
and instantiate contracts, which is not acceptable.
