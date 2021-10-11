# Read-only replica node used for quering pusposes

- Separation of queries from command functionalities thus offloading validators
- Use events to implement event sourcing feeding the replica with blockchain data
- Employ GraphQL to offer clients dynamic query execution on collected data

# Enhance smart contract safety by introducing Resource-Oriented Programming

A fundamental property of digital assets is that they cannot be copied nor implicitly discarded. We have witnessed many instances of thefts in the crypto world. The majority of the cases have been possible due to bugs in the code of the smart contracts treating digital assets as copyable or discardable values.
We aspire to enhance the safety of smart contracts in the Cosmos ecosystem by bringing the concept of resources for representing digital assets to life.

References:
1. [Move language](https://developers.diem.com/docs/technical-papers/move-paper)
2. [Cadence language](https://www.onflow.org/post/resources-programming-ownership)
