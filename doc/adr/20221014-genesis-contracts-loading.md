# [Initialization of nolus chain with appropriate wasm settings]

- Status: accepted
- Deciders: the Nolus dev team
- Published: 2022-10-14
- Last updated: 2024-08-08
- Tags: genesis, wasm, infrastructure, smart-contract

Technical Story: https://app.clickup.com/t/1yyn7rm

## Context and Problem Statement

### Context

In the genesis you can set global settings in the wasm module for who can upload and instantiate contracts. The options that we have for access are:
`Everybody`, `Nobody`, `OnlyAddress` (current wasmd version 0.27.0). If we upgrade the wasmd version, there will be also one more option - `OnlyAddresses`.
We can also set scoped instantiation permission for each code-upload(contract). This could conflict with the global settings, for example, if we have a contract with instantiation permissions set to X address, and in the global scope we have set instantiation to `OnlyAddress` with address Y, there will be conflict.
Each smart contract has an optional admin role. If a contract instance has an admin(address) set, then this admin could interact with the contract directly(for example invoking migrate-contract) rather than with gov proposals.

### Problem statement

We want to run our chain as a permissioned one. We also want to upload and instantiate smart-contracts at the genesis.
The decision that most suited both of the above requirements was to have one address as the only permissioned address which can upload and instantiate contracts. This address will also be owner of the contracts. We would use an administrator contract for this purpose on each of our networks or to have one address as the only permissioned address which can upload and instantiate contracts (we refer to it as contracts_owner).

## Decision Drivers

- If we were to have all of our contracts instantiated with an admin address, we would have to keep it's private key. And if anyone
broke through and got the private key, he might have malicious intentions and damage our contracts.
- Using the Admin contract's address as contracts owner has limited options in terms of what one might do(can execute functions only defined in the contract) which would make automated testing more difficult, but it gives security.

## Considered Options

- [option 1] Having messages for uploading and instantiating contracts in the wasm->gen_msgs subsection in the genesis, where each instance will have instantiation permissions for contracts_owner.
Global wasm permissions:
    code-upload: OnlyAddress - contracts_owner
    instantiation: Everybody

- [option 2] Wasm section also has two subsections codes[] and contracts[] in the genesis. We considered having all of our contracts respectively uploaded and instantiated via appending them to the codes[] and contracts[] subsection.

- [option 3] Starting the chain as permissionless then making a gov proposal to change to permissioned.

- [option 4] Having messages for uploading and instantiating contracts in the wasm->gen_msgs subsection in the genesis. We would set the global permissions for uploading and instantiating contracts to only one address - the address of the Leaser contract.

- [option 5] Having messages for uploading and instantiating contracts in the wasm->gen_msgs subsection in the genesis, where each instance will have instantiation permissions for Admin contract.
Global wasm permissions:
    code-upload: OnlyAddress - admin contract address
    instantiation: Everybody

- [option 6] Upload and instantiate contracts after the chain has started, where each contract instance will have a specific address with instantiation permissions, selected when uploading the contracts.
Global wasm permissions:
    code-upload: OnlyAddress - admin contract address
    instantiation: Everybody

## Decision Outcome

We decided to go with option-6 because it is the most safe approach and we won't need to overengineer the genesis creation.
We also decided to decuple from neutron wasm fork and we won't be using the gen_msgs anymore as it is not supported in the basic wasmd functionlities. We've created
a wasmd-nolus fork which includes code optimization and ibc events filtering which are required for neutron interchain-txs and interchain-queries' modules.

In the future if we need faster contracts execution, we can do code pinning [https://docs.cosmwasm.com/docs/1.0/smart-contracts/code-pinning/]
We can also have DAO governed smart contracts, adding another layer to voting [https://docs.cosmwasm.com/dev-academy/dao-governance/what-is-a-dao#dao-governed-smart-contracts]

If we consider option-2, we would have to add additional scripts for retrieving code hash.
Also we would need to copy contract-states for each instance(this option seems good for hard forks)

If we consider option-3, we would have a short window where everyone could be able to upload
and instantiate contracts, which is not acceptable.
