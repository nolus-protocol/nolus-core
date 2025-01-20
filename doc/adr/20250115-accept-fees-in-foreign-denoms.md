# Upgrade to accept fees in foreign denoms

- Status: Accepted
- Deciders: Product Owner, Development Team
- Date: (2023-12-15)
- Updated : 2025-01-15
- Tags: foreign fees, cosmos-sdk, refactoring

Technical Story
We aim to allow the nolus protocol to accept fees paid in foreign denoms and not only in NLS. This will involve a custom fee checker functionality which was introduced in cosmos-sdk@v47+, executed with the help of ante handlers. The custom fee checker will be implemented as a part of the x/tax module.

## Context and Problem Statement

The decision to enable paying in foreign denoms (to the nolus network) is driven by the desire to scale our protocol and make it easier to use for new users. The protocol won't support all kinds of foreign denoms out of the box but will be controlled with parameters for the custom nolus' x/gov module.

## Decision Drivers


Having the ability to pay fees in denoms other than NLS will allow us to not require every user to own NLS to use the network. This will make it easier for newcomers to try out the protocol with their existing funds.

## Decision Outcome

The decision we took is to implement a custom TxFeeChecker functionality executed by the ante handler - cosmos-sdk/x/auth/ante deductFee decorator.The custom txFeeChecker is an optional parameter to the deductFee decorator. We forked the default implementation of the txFeecheker and built additional logic on it to support our use case. 

OLD *We used our custom wasm contracts to query for prices and then calculate the fees needed to be paid. Our tax module has new parameters with the following proto format:*
UPDATED: We use custom parameters from the tax module where we select each denom that we want to accept and set a fraction of what is the minimum accepted fee per gas. 
The calculation is similar to the default implementation for the fee in base asset. Check the proto files of the tax module to see the new parameters.

For each DEX that we work with, there will be a separate profit address. The denoms that we want to accept as fees will be defined for each DEX. For example, ATOM transferred from dex1 to nolus will have one denom, and ATOM transferred from dex2 to nolus will have another denom. The custom txFeeChecker will compare the denom that the user paid with the accepted denoms. When there is a match, we will know what fraction to use for the minimum required fees calculation. The deducted tax, not paid in the base asset, will be sent to the corresponding profit address from the parameters configuration.

### Positive Consequences

- Easier Onboarding for New Users: New users can participate in the nolus network without the necessity to acquire NLS initially. This reduces the barrier to entry and encourages experimentation with the protocol using existing funds.

- Scalability Enhancement: The flexibility to accept fees in various denoms promotes scalability. As the network evolves, additional foreign denoms can be incorporated through parameter adjustments in the custom nolus' x/gov module.

## Potential evolution

- Monitoring and Analytics: Implement monitoring and analytics tools to track the usage of different denoms and fee structures. This data can inform future adjustments to parameters and provide insights into user behavior within the nolus network.

- Security Audits and Compliance: Conduct regular security audits to ensure the robustness of the implemented fee calculation logic. Stay abreast of regulatory developments and ensure compliance with evolving standards related to decentralized finance (DeFi) and blockchain protocols.

- Cross-Chain Compatibility: Explore possibilities for cross-chain compatibility, allowing the nolus protocol to interact seamlessly with assets and protocols on other blockchains. This could open up new avenues for liquidity and user adoption.

This decision is documented to provide clarity on the rationale behind the upgrade and serves as a reference for future discussions and evaluations related to our blockchain development efforts.