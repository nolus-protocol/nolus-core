# Run priviledged oracle instances as a solution for feeding market data

- Status: accepted
- Deciders: the product owner, the dev team
- Date: 2021-12-16
- Tags: oracle market-data

## Context and Problem Statement

Blockchains progress their state reliably and securely enabled by the ability of every participant to verify state transitions and receive deterministic results. To achieve that, Blockchains live in isolated environments where access to external data is limited.

To implement Nolus's core business logic we need secure and reliable access to up-to-date external data like market feeds and global time to name a few.

## Decision Drivers <!-- optional -->

- reliable - failure resistant, high-availability
- accurate - provided data reflects the real outside-world data. This implies resistance to malicious behavior.
- time to market - how much time it will take to implement
- optimal - avoid feeding same or insignificantly changed data. The significance is defined as a % change from the last reported value.

 ## Considered Options

- In-house solution

|Oracle <---> Data Source |Crypto Exchanges (Binance, Coinbase, Huobi, UniSwap, ...) | Data Aggregators (CoinGecko, Kaiko, Amberdata, ...)
---|---|---
|**Centralized**, a trusted external service instance owning a Nolus privileged account collects, aggregates, and pushes data into a smart contract. It is a simple read/write register. | [-] single-point-of-failure<p> [-] vulnerable to attacks<p> [-] less accurate data<p> [+] free data<p> [+] shorter time-to-market | [-] single-point-of-failure<p> [-] vulnerable to attacks<p> [+] more accurate data<p> [-] paid data<p> [+] shorter time-to-market
|**Fault-Tolerant Centralized**, a set of trusted external service instances, each owning a Nolus privileged account, collect, aggregate, and push data into a smart contract. It validates and aggregates received observations in rounds. | [+] reliable<p> [+] resistent to attacks<p> [-] less accurate data<p> [+] free data<p> [+] shorter time-to-market | [+] reliable<p> [+] resistent to attacks<p> [+] more accurate data<p> [-] paid data<p> [+] shorter time-to-market
|**Decentralized**, multiple untrusted service instances, each depositing some amount, collect, aggregate, and push data into a smart contract. It aggregates received observations in rounds and compensates or penalizes oracles.| [+] reliable<p> [+] resistent to attacks<p> [-] less accurate data<p> [+] free data<p> [-] longer time-to-market| [+] reliable<p> [+] resistent to attacks<p> [+] more accurate data<p> [-] paid data<p> [-] longer time-to-market

Demo code for implementing an aggregator in Rust for CosmWasm on Terra [here](https://github.com/smartcontractkit/chainlink-terra-feeds-demo) and [by Hack.bg](https://github.com/hackbg/chainlink-terra-cosmwasm-contracts).

- Chainlink's Decentralized Oracle Network, [DON](https://chain.link/education/blockchain-oracles#decentralized-oracles), [src](https://github.com/smartcontractkit/chainlink)

"A Decentralized Oracle Network, or DON for short, combines multiple independent oracle node operators and multiple reliable data sources to establish end-to-end decentralization." [1]

DON is an implementation of a decentralized oracle solution targeting EVM networks only. Chainlink develops Node Operator software in Go and a Solidity Smart Contract.

- Integration with [Bandchain](https://bandprotocol.com/)

Bandchain is a CosmosSDK-based chain providing data via IBC channels on requests sent from the counterparty's network. Bandchain data retrieval and aggregation logic is implemented in [Data Sources](https://docs.bandchain.org/custom-script/data-source/introduction.html) and [Oracle Scripts](https://docs.bandchain.org/custom-script/oracle-script/introduction.html). The former are Python scripts executed off-chain by validators whereas the latter are wasm modules executed on-chain to process further the data collected by data sources. Both are deployed onto the network by anyone needing them.

## Decision Outcome

The product owner has confirmed the dev team's proposal to start with "an in-house developed **Fault-Tolerant Centralized** oracle based on data aggregators", because
* external solutions either do not support Cosmos chains or provide market data only on demand,
* the security of the decentralized in-house solutions rely on enough staked amounts that fall aside the tokenomics, and
* the single instance centralized solution does not provide enough availability and is not failure nor vulnerability resistant

## Potential evolution

* engage validators and their stake

## Links

1. [ChainLink's overview of Blockchain Oracles](https://chain.link/education/blockchain-oracles)
2. [What is the Blockchain Oracle Problem](https://blog.chain.link/what-is-the-blockchain-oracle-problem/)