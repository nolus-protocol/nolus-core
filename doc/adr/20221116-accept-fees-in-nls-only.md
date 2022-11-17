# Accept fees in NLS only

- Status: accepted
- Deciders: the product owner, the dev team
- Date: 2022-11-10
- Tags: tax module, fees, unls, nls, base denom

Technical Story:
After our initial audit of the protocol an issue was found on how we deduct the tax - In x/tax/keeper/taxdecorator.go:76-81, the ApplyTax function called from the tax module’s AnteHandler is performing an unbounded iteration over the fee Coins provided by users.

An attacker could craft a message with a significant number of Coins with the intention of
slowing down the block production, which in extreme cases may lead to Tendermint’s
propose timeout to be surpassed.

This triggered a refactoring of the tax ante handler and the unit tests connected to it. Initially we agreed that it is a good time to prepare our code base for accepting fees in different denoms approved by governance.

## Context and Problem Statement

During implementation we found out the limitations we have to accept fees in other then our base denom - unls.

The process to accept fee, in atom for example, would be the following:
1. Calculate the minimum required fee in our base denom.
2. Calculate the amount of provided atom in unls by the spot price of atom/unls pair.
3. Compare minimum required fee with the unls amount.
4. If the fee amount is sufficient we proceed to collecting the fee.
5. Collect the fee by actually swapping the atom to unls and send it the fee collector 

The problem comes from that we don't have the actual spot price and the capability of swapping on our blockchain. This results in non atomic operation which is something we don't want.

## Decision Drivers

- simple and robust mechanism to deduct tax
- do not introduce dependencies to third party for swapping and price feeding
- atomic transactions

## Decision Outcome

The decision we took is to restrict our tax module to accept single fee coin in our base denom.

### Positive Consequences

- Remove the issue of the tax module’s AnteHandler to performing an unbounded iteration over the fee Coins. 
- Remove the need to have real time spot prices of coins
- Remove the need to implement coin swap on our blockchain 
