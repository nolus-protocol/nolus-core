# Upgrade cosmos-sdk to v0.53 and ibc-go to v10

- Status: Accepted
- Deciders: Product Owner, Development Team
- Date: 2025-05-09
- Tags: cosmos-sdk, ibc-go, upgrading

Technical Story - upgrade to:

 cosmos-sdk v0.53 -  introducing unordered transactions, performance improvements, and additional modules (not currently in use). [Changelog](https://github.com/cosmos/cosmos-sdk/blob/v0.53.0/UPGRADING.md)

 ibc-go v10 – enables v2 transfers, deprecates the 29-fee middleware, improves interchain fee abstraction, and includes a series of quality-of-life updates.
[Migration Guide](https://ibc.cosmos.network/v10/migrations/v8_1-to-v10/)

## Context and Problem Statement

The decision to upgrade our cosmos-sdk and ibc-go versions is driven by the desire to access the latest versions providing new features, security fixes, performance enhancements, and forward compatibility with emerging interchain standards. These are strategic for future integrations and ecosystem alignment.
Cosmos-sdk/x/gov module doesn't yet support removed type messages, and there were some removed proposals which we've used in the past with older ibc-go versions.

## Decision Drivers

- Align with the latest supported versions of core libraries
- Enable future support for unordered transactions
- Deprecate legacy and now-unsupported components (e.g., 29-fee middleware)
- Reduce technical debt and ease of maintenance
- Prepare for enhanced IBC integrations and interchain features

## Decision Outcome

The decision we took is to upgrade cosmos-sdk from v0.50 to v0.53 and to upgrade ibc-go from v8 to v10.
We decided to create and keep a `legacycodec` folder containing all the deprecated message types that we need for querying our proposals. Once there is a more elegant solution available in the cosmos-sdk/x/gov, we will use it. 

### Positive Consequences

- Increased compatibility with upstream IBC-enabled chains and tooling
- Modernized codebase with enhanced maintainability
- Performance improvements from core updates in cosmos-sdk
- Reduced reliance on deprecated components and middleware
- Enables opt-in adoption of unordered transaction processing

## Potential evolution

Potential evolution is enabling ibc-go transfer V2 functionality which will open the opportunity to connect with other non cosmos native chains if they have
a compatible ibc light client. We could also utilize some new modules from the sdk. Overall the upgrade sets us in a good position to continue aligning with the core libraries and develop the project, accessing the latest functionalities.