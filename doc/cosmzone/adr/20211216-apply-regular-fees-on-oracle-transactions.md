# Apply regular fees on oracle transactions

- Status: accepted
- Deciders: the product owner, the dev team
- Date: 2021-12-16
- Tags: oracle transactions fees

## Context and Problem Statement

Do we need to treat transactions initiated by oracles as special ones and grant them `free tickets` for the network? How much would they cost in operations versus the added cost of developing, testing, and maintaining a `no-fees` case in the existing Cosmos-SDK modules, including the `auth` ante handler?

## Decision Drivers

- security - censorship
- cost - initial and operational, e.g. transaction fees, complexity in code maintenance, potential bugs, etc.

## Considered Options

- classify oracle transactions as free and apply no fees
- apply regular fees

## Decision Outcome

Chosen option: `"apply regular fees"`, because:

- the added complexity and maintenance cost in addition to potential liveness issues related to censorship from validators would outweight the transaction fees, and
- the added 40% flat tax on each transaction would collect the necessary amount to cover the oracle transactions' fees
