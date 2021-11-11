# mintscan
https://www.mintscan.io/cosmos  
service, closed source, has most networks

contact for integration  
business@cosmostation.io  
support@cosmostation.io

# big-dipper
https://cosmos.bigdipper.live/  
service, open source, has some networks

https://github.com/forbole/big-dipper

contact for integration  
info@forbole.com

- Nomo configuration for Big dipper:
```js
{
    "public":{
        "chainName": "Nomo",
        "chainId": "nomo-private",
        "gtm": "{Add your Google Tag Manager ID here}",
        "slashingWindow": 10000,
        "uptimeWindow": 250,
        "initialPageSize": 30,
        "secp256k1": false,
        "bech32PrefixAccAddr": "nomo",
        "bech32PrefixAccPub": "nomopub",
        "bech32PrefixValAddr": "nomovaloper",
        "bech32PrefixValPub": "nomovaloperpub",
        "bech32PrefixConsAddr": "nomovalcons",
        "bech32PrefixConsPub": "nomovalconspub",
        "bondDenom": "nomo",
        "powerReduction": 1000000,
        "coins": [
            {
                "denom": "nomo",
                "displayName": "NOMO",
                "fraction": 1000000
            }
        ],
        "ledger":{
            "coinType": 118,
            "appName": "Cosmos",
            "appVersion": "2.16.0",
            "gasPrice": 0.02
        },
        "modules": {
            "bank": true,
            "supply": true,
            "minting": false,
            "gov": false,
            "distribution": false
        },
        "coingeckoId": "cosmos",
        "networks": "https://gist.githubusercontent.com/kwunyeung/8be4598c77c61e497dfc7220a678b3ee/raw/bd-networks.json",
        "banners": false
    },
    "remote":{
        "rpc":"http://127.0.0.1:26657",
        "api":"http://127.0.0.1:1317"
    },
    "debug": {
        "startTimer": true
    },
    "params":{
        "startHeight": 0,
        "defaultBlockTime": 5000,
        "validatorUpdateWindow": 300,
        "blockInterval": 15000,
        "transactionsInterval": 18000,
        "keybaseFetchingInterval": 18000000,
        "consensusInterval": 1000,
        "statusInterval":7500,
        "signingInfoInterval": 1800000,
        "proposalInterval": 5000,
        "missedBlocksInterval": 60000,
        "delegationInterval": 900000
    }
}



```

# hubble
https://hubble.figment.io/  
service, open source, has some networks

https://github.com/figment-networks/hubble

contact for integration  
support@figment.io

# Other Explorers
### terra block explorer
own explorer  
https://github.com/terra-money/finder

### token view
https://atom.tokenview.com/  
closed source

### aneka
https://cosmos.aneka.io  
closed source

### stake id
https://stake.id/#/  
https://terra.stake.id/#/  
closed source

### atom scan
https://atomscan.com/  
closed source

### cosmos scan
https://cosmoscan.io  
https://github.com/everstake/cosmoscan-front  
open source, limited functionality(does not show blocks?)

### anthem
https://github.com/ChorusOne/anthem  
discontinued

### stargazer
https://stargazer.certus.one/  
discontinued

### coris
https://coris.network/  
discontinued
