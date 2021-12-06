# Use XYZ as an oracle solution for feeding external data

- Status: draft
- Deciders:
- Date: 2021-11-30
- Tags: oracle market-data

## Context and Problem Statement

Blockchains progress their state reliably and securely enabled by the ability of every participant to verify state transitions and receive deterministic results. To achieve that, Blockchains live in isolated environments where access to external data is limited.

To implement Nolus's core business logic we need secure and reliable access to up-to-date external data like market feeds and global time to name a few.

## Decision Drivers <!-- optional -->

- reliable - failure resistant, high-availability
- accurate - provided data reflects the real outside-world data. This implies resistance to malicious behavior.
- optimal - avoid feeding same or insignificantly changed data. The significance is defined as a % change from the last reported value.

 ## Considered Options

- In-house solution

|Oracle <---> Data Source |Crypto Exchanges (Binance, Coinbase, Huobi, UniSwap, ...) | Data Aggregators (CoinGecko, Kaiko, Amberdata, ...)
---|---|---
|**Centralized**, a trusted external service instance owning a Nolus priviledged account collects, aggregates and pushes data into a smart contract. It is a simple read/write register. | [+] free data<p> [-] unreliable<p> [-] vulnerable to attacks<p> [-] less accurate data| [-] paid data<p> [-] reliable<p> [-] resistent to attacks<p> [+] more accurate data
|**Decentralized**, multiple untrusted service instances depositing some amount collect, aggregate and push data into a smart contract. It aggregates received observations in rounds and compensates or penalizes oracles.| [+] free data<p> [+] reliable<p> [+] resistent to attacks<p> [-] less accurate data| [-] paid data<p> [+] reliable<p> [+] resistent to attacks<p> [+] more accurate data

Demo code for implementing an aggregator in Rust for CosmWasm on Terra [here](https://github.com/smartcontractkit/chainlink-terra-feeds-demo) and [by Hack.bg](https://github.com/hackbg/chainlink-terra-cosmwasm-contracts).

- Chainlink's Decentralized Oracle Network, [DON](https://chain.link/education/blockchain-oracles#decentralized-oracles), [src](https://github.com/smartcontractkit/chainlink)

"A Decentralized Oracle Network, or DON for short, combines multiple independent oracle node operators and multiple reliable data sources to establish end-to-end decentralization." [1]

DON is an implementation of a decentralized oracle solution targeting EVM networks only. Chainlink develops Node Operator software in Go and a Solidity Smart Contract.

- Integration with [Bandchain](https://bandprotocol.com/)

Bandchain is a CosmosSDK-based chain providing data via IBC channels on requests sent from the counterparty's network. Bandchain data retrieval and aggregation logic is implemented in [Data Sources](https://docs.bandchain.org/custom-script/data-source/introduction.html) and [Oracle Scripts](https://docs.bandchain.org/custom-script/oracle-script/introduction.html). The former are Python scripts executed off-chain by validators whereas the latter are wasm modules executed on-chain to process further the data collected by data sources. Both are deployed onto the network by anyone needing them.

## Decision Outcome

Chosen option: "[option 1]", because [justification. e.g., only option, which meets k.o. criterion decision driver | which resolves force force | … | comes out best (see below)].

### Positive Consequences <!-- optional -->

- [e.g., improvement of quality attribute satisfaction, follow-up decisions required, …]
- …

### Negative Consequences <!-- optional -->

- [e.g., compromising quality attribute, follow-up decisions required, …]
- …

## Pros and Cons of the Options <!-- optional -->

### Centralized Oracle

[example | description | pointer to more information | …] <!-- optional -->

- Good, because [argument a]
- Good, because [argument b]
- [cons] single point of failure, because a single node/service provides input to the smart contracts to work properly
- [cons] centralization, because a single entity controlling the node/service is provided with an excessive power to influence the core business logic implemented by smart contracts
- … <!-- numbers of pros and cons can vary -->

### [option 2]

[example | description | pointer to more information | …] <!-- optional -->

- Good, because [argument a]
- Good, because [argument b]
- Bad, because [argument c]
- … <!-- numbers of pros and cons can vary -->

### [option 3]

[example | description | pointer to more information | …] <!-- optional -->

- Good, because [argument a]
- Good, because [argument b]
- Bad, because [argument c]
- … <!-- numbers of pros and cons can vary -->

## Links <!-- optional -->

1. [ChainLink's overview of Blockchain Oracles](https://chain.link/education/blockchain-oracles)
2. [What is the Blockchain Oracle Problem](https://blog.chain.link/what-is-the-blockchain-oracle-problem/)