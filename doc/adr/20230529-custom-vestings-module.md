# Accept fees in NLS only

- Status: accepted
- Deciders: the product owner, the dev team
- Date: 2023-05-29
- Tags: vestings module, vesting accounts, custom vesting start time

Technical Story:
We want to add vesting accounts with a custom start time which is currently not supported in the cosmos-sdk's native vesting module.
There is an ongoing discussion in the cosmos-sdk community about adding this feature/flag to the native vesting module. 
https://github.com/cosmos/cosmos-sdk/issues/4287

We decided to implement our own vesting module which will support custom vesting start time.

## Context and Problem Statement

We need to be able to create vesting accounts that have start-time different than the current block time after the chain has started.

There is a command for creating vesting accounts in the cosmos-sdk's native vesting module, but it does not support custom vesting start time. 
Instead it always uses the current block time as the vesting start time.

## Decision Drivers

- We need to be able to create vesting accounts with a custom start time after the chain has started

## Decision Outcome

The decision we took is to create a new module that will support adding a vesting account with custom vesting start time.
We've seen that there is a similar module developed for this purpose in stargaze network(alloc module), but it has additional features that we don't need.

### Positive Consequences

- We will be able to create vesting accounts with custom vesting start time after the chain has started


## Potential evolution

If a flag for custom start time is added to the cosmos-sdk's native vesting module, we can deprecate our custom vesting module and use the native one instead.