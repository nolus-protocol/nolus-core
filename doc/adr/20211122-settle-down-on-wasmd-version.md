# Settle down on wasmd version

- Status: accepted
- Deciders: the Nolus dev team
- Date: 2021-11-22
- Tags:

## Context and Problem Statement

We cannot use wasm module, provided from the CosmWasm team via the library wasmd with the latest CosmosSdk versions +v0.43 - wasmd issue tracking number [#501](https://github.com/CosmWasm/wasmd/issues/501). Originally, CosmWasm were pushing to release wasmd version 1.0 in a state where it would have used a CosmWasm-VM version 1.0 while still being only CosmosSdk v0.42 compatible. However, after Cosmos Hub announced that they would migrate to v[0.44](https://github.com/cosmos/gaia/blob/main/docs/roadmap/cosmos-hub-roadmap-2.0.md) CosmWasm decided that they will release a version of wasmd which would be CosmosSDK v0.44 compatible and target a wasmd 1.0 release afterwards.

Based on these observations, we need to decide how to proceed with the CosmWasm integration. The main routes we can take are:

 - We can stay on CosmosSDK v0.42, use the official wasmd version and only migrate once CosmWasm's wasmd v0.22 is released.

 - Use a Provenance wasmd fork - we have a wasmd 19 (i.e. pre CosmWasm-VM 1.0-beta release) on the latest v0.44 CosmosSDK. Again, migrate to CosmWasm's wasmd once v0.22 is released.

 - Create our own fork that targets our desired compatability and migrate to an official module in a later stage.

## Decision Outcome

We have decided on staying with the Provenance wasmd fork and migration to wasmd v0.22 as soon as it gets released. Although the CosmWasm team haven't set a hard deadline for the v0.22 release, they have announced a that they target the end of the year on their [Community Call #32](https://vimeo.com/646566481). As this timeline would align with our project goals and would mean that we would not have to go in mainnet with the fork's version, this approach would require the least resources while still being robust in the long run.