# [lunie-ng](https://github.com/tendermint/lunie-ng)

## Pros
- web based wallet, stores wallet addresses on the user's machine
- can be built and hosted on a CDN
- supports the base functionality of a blockchain wallet (create wallet, restore using mnemonic, balance, history, transfer)
- no external servers/apis required except the blockchain node's API
- supports all native currencies that can enter our blockchain through IBC
- integration with keplr (using https://docs.keplr.app/api/suggest-chain.html)
- ledger support
- Apache 2.0 license, meaning we can fork and rebrand/modify as we like

## Cons
- has not received any updates in the last 6 months
- does not support CW20 tokens
- no mobile apps
- not a "universal" wallet - only supports working with a single network

## Usable features
- base wallet functionality
- ledger support (chrome only)
- keplr integration (chrome only)
- staking

## Features we might not need
- governance proposal voting

## Cosmos SDK module requirements
- auth
- bank
- distribution (used by the staking feature)
- gov (used by the governance feature)
- staking (used by the staking feature)
- ibc (to display external native currency balances)

## Additional development required, if picked
- CW20 token support
- mobile apps in the future
- migration from deprecated (removed in the latest version of cosmos sdk) transaction broadcast endpoint to the new one
- migration from deprecated IBC transfer endpoint to the new one
- possibly other minor issues due to the wallet being neglected for months

## Aditional notes
- written in JavaScript, using Vue.js and NuxtJS

## network.js used to test locally
```js
export default {
  id: 'nomo-private',
  name: 'Nomo Local Net',
  description:
    'Cosmos is a network of independent parallel blockchains, powered by BFT consensus algorithms like Tendermint.',
  logo: `logo.svg`,
  website: 'https://cosmos.network',
  // apiURL: 'http://localhost:8010/proxy',
  apiURL: 'http://localhost:1317',
  rpcURL: 'http://localhost:26657',
  minBlockHeight: 1, // actually 5200791, but it has the wrong block time.
  stakingDenom: 'nomo',
  coinLookup: [
    {
      viewDenom: 'nomo',
      chainDenom: 'nomo',
      chainToViewConversionFactor: 1e-6,
      icon: `currencies/atom.png`,
    },
  ],
  addressPrefix: 'nomo',
  validatorAddressPrefix: 'nomovaloper',
  validatorConsensusaddressPrefix: 'nomovalcons', // needed to map validators from staking queries to the validator set
  HDPath: `m/44'/118'/0'/0/0`,
  lockUpPeriod: `3 days`,
  fees: {
    default: {
      gasEstimate: 350000,
      feeOptions: [
        {
          denom: 'nomo',
          amount: 0.001,
        },
      ],
    },
  },
  // icon: `https://lunie.fra1.digitaloceanspaces.com/network-icons/cosmos.png`,
  localSigning: true, // this is only to be used as a developer tool - never deployed in production or for mainnet chains
}
```