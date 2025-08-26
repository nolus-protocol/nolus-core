# Enable EIP-191 signing and upgrade to latest wasmd

- Status: Accepted
- Deciders: Product Owner, Development Team
- Date: 2025-08-25
- Tags: metamask, ethereum, wasmd, ibc-v2

Technical Story - 

 Enable EIP-191 signing.

 Upgrade to latest wasmd v0.61 -  introducing ibc V2 support, wasmvm 3.0 and other quality of life improvements. [Changelog](https://github.com/CosmWasm/wasmd/blob/v0.61.2/CHANGELOG.md)


## Context and Problem Statement

The decision to upgrade our wasmd version is driven by the desire to access the latest versions providing new features, security fixes, performance enhancements, and forward compatibility with emerging interchain standards. These are strategic for future integrations and ecosystem alignment.
Enabling EIP-191 signing allows us to provide an easier onboarding proccess for users that are used to ethereum wallets like metamask. With this approach we don't require creating and using a whole new wallet like keplr (if they are not familiar/uncomfortable).

## Decision Drivers

- Align with the latest supported versions of core libraries
- Enable EIP-191 signing for easier onboarding
- Patch security reports from various libraries
- Prepare for potential IBC V2 usage

## Decision Outcome

The decision we took is to upgrade wasmd from v0.60 to v0.61 and to enable EIP-191 signing.
Enabling EIP-191 also resulted in some additional changes/code in our cosmos-sdk fork.

### Positive Consequences

- Easier user onboarding
- Security improvements from updates in cosmos-sdk, wasmd and other libraries

## Potential evolution

Potential evolution is enabling ibc-go transfer V2 functionality which will open the opportunity to connect with other non cosmos native chains if they have
a compatible ibc light client. We could follow the same model from EIP-191 to enable transactions signing for other wallets like the solana based phantom wallet
which will broaden the user base that we could attract.