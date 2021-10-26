25.10.2021
==========

- [in progress] vesting - no progress since the last week, need verification with functional tests of the CosmosSDK module
- [in progress] dev network - researching the options to launch a network and set up its genesis, validator set, and p2p nodes
- [in progress] fuzzy and integration testing - a framework for the latter integrated into our codebase, pending CI integration
- wallet - we stick to Keplr
    - [open question] use case *NewUser* - a rebranded wallet is provided to the user to install
    - [in progress] use case *ExistingCosmosUser* - connect her wallet to Nomo network
- web-app - it's time to start working on it
- [in progress] bridge - see below for details, added another option
    - use Terra to reach Wormhole
    - nevertheless, which we chose, the end-customer will not have to deal with it. Nomo is going to ask (s)he to transfer BTC to our address. Then NOMO is going to buy wBTC and use the bridge to move the coins into Nomo
- launching dev/test network - a one-click setup on some Cloud providers is nice to have
- explorer - to research how/if we can apply for integration into some of the existing explorers

19.10.2021
==========

- vesting - it seems the built-in module covers the use case. Going to build and integrate functional tests into CI/CD.
- dev network - going to prepare a Docker image and deploy it on GCP and AWS, a total of three validators. Integrate the deployment as a GitLab pipeline.
- fuzzy and integration testing - the team is researching what is the state-of-the-art in Cosmos. The aim is to integrate them into CI/CD pipelines.
- wallet - Keplr and Lunie are being researched. It seems the former is more mature and easy to setup and use. In addition, it is well-known in the Cosmos community
    - usecase *NewUser* - a rebranded wallet is provided to the user to install
    - usecase *ExistingCosmosUser* - connect her wallet to Nomo network
- bridge - discussed a few options. Going to compare them based on the anticipated CAPEX and OPEX
    - build our bridges to BTC and ETH
    - build our bridge to ETH and use it for BTC as well relying on wrapped BTC in ETH
    - use the Gravity bridge
    - use ThorChain
- swap - list $NOMO in DEX-es after the main NOMO net is launched. The first version will not require exchanging since it will support using leasing and deposit products only using the same currency as input and output

12.10.2021
==========

- $NOMO tokenomics to be finalized soon and implemented in the genesys
- rewards are going to be part of the tokenomics
- presale is going to be done on the test net, then build them in the genesys of the main net
- token validators / replica will be known before main net launch
- research the available wallets in the Cosmos ecosystem and decide on whether to use or embed one
- [simmilarly to as the above item] research the available BTC and ETH bridges and decide on whether to use or instantiate/implement our own
- research external market data providers and approaches for integration, chainlink?
- an exit from the leasing product may either be completing the lease or automatic liquidation due to imbalance of asset values, due value vs. collateral + payments
- the development of front-ends is going on par with the backend of the platform. [P0] a Web app with support for mobiles. [P1] - native apps, Android, iOS


11.10.2021
==========

- [P0] 50% token to distribute to given customers - private sell - via contract, and wallet
- [P0] paramerized vesting contract and the tokens issued to customers bought during the private sell and got locked into
- off-chain governance of the smart contracts
- allow token owners to stake them or run a node
- ramps will have liquidity pools fed with NOMO and the customers will be able to buy/sell NOMO sending and withdrawing fiat
- gas will be calculated
- when an IBC-compatible currency is sent to another IBC then the amount is kept in the same currency on the other subnet, not wrapped in the native currency
- leasing amount for collateral of the same currency
- the leasing contract buys instantly the lended amount with stable coins from an account hold by a custodian?