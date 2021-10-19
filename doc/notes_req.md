19.10.2021
==========

- vesting - it seems the built-in module cover the usecase. Going to build and CI/CD integrate functional tests.
- dev network - going to prepare a Docker image and deploy it on GCP and AWS, total three validators. Integrate the deployment as a GitLab pipeline.
- fuzzy and integration testing - the team is researching on what is the state-of-the-art in Cosmos. The aim is to integrate them in CI/CD pipelines.
- wallet - Keplr and Lunie are being researched. It seems the former is more mature and easy to setup and use. In addition, it is well-known in the Cosmos community.
    - usecase *NewUser* - a rebranded wallet is provided to the user to install
    - usecase *ExistingCosmosUser* - connect her wallet to Nomo network
- bridge - discussed a few options. Going to compare them based on the anticipated capex and opex
    - build our bridges to BTC and ETH
    - build our bridge to ETH and use it for BTC as well relying on wrapped BTC in ETH
    - use the Gravity bridge
    - use ThorChain
- swap - list $NOMO in DEX-es after the main NOMO net is launched. The first version will not require exchanging since it will support using leasing and deposit products only using same currency as input and output


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