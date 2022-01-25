# Choosing testnet explorer

- Status: accepted
- Deciders: the dev team
- Date: 2022-01-25
- Tags: testnet, infrastructure

## Context and Problem Statement

The blockchain explorer is one of the most important tools for crypto users as it allows them to observe the state of the blockchain without the need for setting up local interfaces for communication with the system [1]. 

There are multiple explorer solutions for Cosmos networks, especially already hosted ones, however as we are still in a pre-release state it would be more suitable for us to find a self-hosted one. In the current state of the ecosystem, this means choosing between BigDipper V1, BigDipper V2 and Ping-Pub (selfhosted).

## Decision Outcome

Based on the research on the state of the explorers: ping-pub, bigdipper v1 and bigdipper v2, we have chosen a custom fork of ping-pub, due to its minimalistic nature it would be easier to modify and deploy on our environment. It is also trivial to push testnet/mainnet configurations in the explorer project when they become available.

## References

1. [Blockchain explorer summary](https://www.gemini.com/cryptopedia/what-is-a-block-explorer-btc-bch-eth-ltc)
2. [Nolus Ping-Pub fork](https://gitlab-nomo.credissimo.net/nomo/ping-pub)
