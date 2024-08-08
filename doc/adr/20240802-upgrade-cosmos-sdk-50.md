# Upgrade cosmos-sdk Library to v0.50+

- Status: Accepted
- Deciders: Product Owner, Development Team
- Date: 2024-08-02
- Tags: cosmos-sdk, refactorings, upgrade

Technical Story
We aim to upgrade the cosmos-sdk library to version v0.50+, involving a comprehensive update of associated dependencies. Key components to be upgraded include cometbft, ibc-go (to v8), cosmos-sdk (to v0.50), iavl (to v1.2.0) wasmd (to v0.51+) and wasmvm (to v2.0+) .

This major upgrade has been an ongoing effort spanning several months, requiring substantial code refactoring and bug fixes to accommodate the changes introduced by the upgrade.

## Context and Problem Statement

The decision to upgrade the cosmos-sdk library to v0.50+ is driven by the desire to access the latest features and bug fixes. 

Notable features and fixes provided by the upgrade include:

The upgrade of the core libraries opens a range of new features and enhancements that could be implemented in the future such as
IBC callbacks, direct price feeds from validators, and more from vote extensions. 

## Decision Drivers

Unlock new functionalities with the latest version of our core libraries.

## Decision Outcome

The decision we took is to upgrade cosmos-sdk from v0.47 to v0.50.7

### Positive Consequences

- Access to Latest Features: The upgrade ensures access to the most recent features and enhancements offered by the core libraries, enhancing the overall capabilities of our system.

- Bug Fixes and Stability: Incorporating the upgrades brings numerous fixes from the cosmos-sdk, contributing to the stability and resilience of our blockchain infrastructure.

## Potential evolution

Keep an eye on the newest versions of the core libraries that we use and upgrade them as needed.
Familiarize more with vote extensions and weigh their potential impact and benefits to our system. 

This decision is documented to provide clarity on the rationale behind the upgrade and serves as a reference for future discussions and evaluations related to our blockchain development efforts.