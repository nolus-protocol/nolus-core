# Upgrade cosmos-sdk Library to v0.47+

- Status: Accepted
- Deciders: Product Owner, Development Team
- Date: 2023-12-15
- Tags: cosmos-sdk, refactorings, upgrade

Technical Story
We aim to upgrade the cosmos-sdk library to version v0.47+, involving a comprehensive update of associated dependencies. Key components to be upgraded include tendermint (to become cometbft), ibc-go (to v7), cosmos-sdk (to v0.47+), and wasmd (to v0.45+).

This major upgrade has been an ongoing effort spanning several months, requiring substantial code refactoring and bug fixes to accommodate the changes introduced by the upgrade.

## Context and Problem Statement

The decision to upgrade the cosmos-sdk library to v0.47+ is driven by the desire to access the latest features and bug fixes. Collaboration with other blockchains such as Neutron and Osmosis necessitates a coordinated upgrade strategy to ensure seamless functionality across platforms.

Notable features and fixes provided by the upgrade include:

IBC-Go v7, addressing a critical fix for ICS27 channels, preventing unexpected closures on timeout.
Multiple fixes from the cosmos-sdk, contributing to overall stability and reliability.
CometBFT - a fork of Tendermint.


## Decision Drivers

Old versions of cosmos-sdk become stale after time and we don't want to use libraries without up to date support. 
Also we want to be able to use the latest features and bug fixes.

## Decision Outcome

The decision we took is to upgrade cosmos-sdk from v0.45 to v0.47, skipping v0.46 which reportedly was not very stable.

### Positive Consequences

- Access to Latest Features: The upgrade ensures access to the most recent features and enhancements offered by the cosmos-sdk library, enhancing the overall capabilities of our system.

- Bug Fixes and Stability: Incorporating the upgrades brings numerous fixes from the cosmos-sdk, contributing to the stability and resilience of our blockchain infrastructure.

- Interoperability: Coordinating upgrades with other blockchains such as Neutron and Osmosis ensures interoperability and prevents potential disruptions in functionality.

## Potential evolution

The upgrade opens avenues for potential evolution, allowing for ongoing improvements and adaptability to emerging technologies in the blockchain ecosystem. Continuous monitoring and adaptation to evolving standards will be crucial for the sustained success of our system. We will monitor the cosmos-sdk library to see how stable v0.50+ versions are and eventually upgrade. 

This decision is documented to provide clarity on the rationale behind the upgrade and serves as a reference for future discussions and evaluations related to our blockchain development efforts.