# [Keplr](https://github.com/chainapsis/keplr-extension)

## Pros
- web (Chrome and Brave extension) and mobile based wallet (mobile version contains only Cosmos, Osmosis and Regen tokens at the moment)
- Can be built and hosted on a CDN.
- Supports the base functionality of a blockchain wallet (create wallet, restore using mnemonic, balance, history, transfer, IBC transfer **(Chrome extension)** and staking).
- You can also use your Google and Apple (mobile only) account to create and restore your wallet.
- No external servers/apis required except the blockchain node's API.
- Supports all native currencies that can enter our blockchain through IBC.
- Ledger support.
- Apache 2.0 license, meaning we can fork and rebrand/modify as we like.
- It is updated relatively often.
- Working with different networks.
- CW20 token support

## Usable features
- base wallet functionality
- ledger support (chrome and mobile)
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

## Aditional notes
- written in TypeScript, using React and ReactNative

## Run on your local:
-Go to `packages/extension/src/config.ts` and add:
-   **NOMO** variables to imports
```js
  NOMO_REST_CONFIG,
  NOMO_REST_ENDPOINT,
  NOMO_RPC_CONFIG,
  NOMO_RPC_ENDPOINT,
```
-   Also add **NOMO** configuration to list with all networks:
```js
{
    rpc: NOMO_RPC_ENDPOINT,
    rpcConfig: NOMO_RPC_CONFIG,
    rest: NOMO_REST_ENDPOINT,
    restConfig: NOMO_REST_CONFIG,
    chainId: "nomo-private",
    chainName: "Nomo",
    stakeCurrency: {
      coinDenom: "nomo",
      coinMinimalDenom: "nomo",
      coinDecimals: 6,
      coinGeckoId: "nomo",
    },
    walletUrl:
      process.env.NODE_ENV === "production"
        ? "https://wallet.keplr.app/#/nomohub/stake"
        : "http://localhost:8080/#/nomohub/stake",
    walletUrlForStaking:
      process.env.NODE_ENV === "production"
        ? "https://wallet.keplr.app/#/nomohub/stake"
        : "http://localhost:8080/#/nomohub/stake",
    bip44: {
      coinType: 118,
    },
    bech32Config: Bech32Address.defaultBech32Config("nomo"),
    currencies: [
      {
        coinDenom: "nomo",
        coinMinimalDenom: "nomo",
        coinDecimals: 6,
        coinGeckoId: "cosmos",
      },
    ],
    feeCurrencies: [
      {
        coinDenom: "nomo",
        coinMinimalDenom: "nomo",
        coinDecimals: 6,
        coinGeckoId: "cosmos",
      },
    ],
    coinType: 118,
    features: ["stargate", "ibc-transfer", "no-legacy-stdTx"],
  },
```
-Go to `packages/extension/src/` folder and copy these two files: `config.var.ts` and `config.ui.var.ts` from `config.ui.var.example.ts` and `config.var.example.ts`. Then you have to add the **NOMO** RPC and HTTP urls to `config.var.ts`:
```js
export const NOMO_RPC_ENDPOINT = "http://127.0.0.1:26657/";
export const NOMO_RPC_CONFIG: AxiosRequestConfig | undefined = undefined;
export const NOMO_REST_ENDPOINT = "http://127.0.0.1:1317";
export const NOMO_REST_CONFIG: AxiosRequestConfig | undefined = undefined;
```
When you complete these steps please build and run your Google chrome extension (using terminal): <br/>
1.  Go to `keplr-extension` root directory and run: `yarn clean && yarn install --frozen-lockfile`
2.  After installing the packages run: `lerna run build` (**Note:** if you don't have `lerna` installed please run: `npm install -G lerna`)
3.  When the build is successfully completed run: `lerna run dev --parallel` (**Note:** this process never ends). After this message you can import your wallet (check this tutorial: https://webkul.com/blog/how-to-install-the-unpacked-extension-in-chrome/):
```
@keplr-wallet/extension: No type errors found
@keplr-wallet/extension: Version: typescript 4.1.5
@keplr-wallet/extension: Time: 626900ms
@keplr-wallet/extension: Hash: 6511ca42d0f605782cb6
@keplr-wallet/extension: Version: webpack 4.44.2
@keplr-wallet/extension: Time: 732618ms
@keplr-wallet/extension: Built at: 10/31/2021 8:16:16 PM
@keplr-wallet/extension:                              Asset       Size                Chunks                   Chunk Names
@keplr-wallet/extension:        assets/NanumBarunGothic.ttf   3.99 MiB                        [emitted]
@keplr-wallet/extension:    assets/NanumBarunGothicBold.ttf   4.21 MiB                        [emitted]
@keplr-wallet/extension:   assets/NanumBarunGothicLight.ttf   4.69 MiB                        [emitted]
@keplr-wallet/extension:                  assets/atom-o.svg  687 bytes                        [emitted]
@keplr-wallet/extension:             assets/broken-link.svg   7.18 KiB                        [emitted]
@keplr-wallet/extension:        assets/export-to-mobile.svg   26.5 KiB                        [emitted]
@keplr-wallet/extension:           assets/fa-brands-400.eot    134 KiB                        [emitted]
@keplr-wallet/extension:           assets/fa-brands-400.svg    730 KiB                        [emitted]
@keplr-wallet/extension:           assets/fa-brands-400.ttf    133 KiB                        [emitted]
@keplr-wallet/extension:          assets/fa-brands-400.woff     90 KiB                        [emitted]
@keplr-wallet/extension:         assets/fa-brands-400.woff2   76.6 KiB                        [emitted]
@keplr-wallet/extension:            assets/fa-solid-900.eot    200 KiB                        [emitted]
@keplr-wallet/extension:            assets/fa-solid-900.svg    896 KiB                        [emitted]
@keplr-wallet/extension:            assets/fa-solid-900.ttf    200 KiB                        [emitted]
@keplr-wallet/extension:           assets/fa-solid-900.woff    102 KiB                        [emitted]
@keplr-wallet/extension:          assets/fa-solid-900.woff2   78.4 KiB                        [emitted]
@keplr-wallet/extension:                assets/icon-128.png   18.5 KiB                        [emitted]
@keplr-wallet/extension:                 assets/icon-16.png    2.3 KiB                        [emitted]
@keplr-wallet/extension:                 assets/icon-48.png   4.86 KiB                        [emitted]
@keplr-wallet/extension:           assets/icons8-cancel.svg  851 bytes                        [emitted]
@keplr-wallet/extension:          assets/icons8-checked.svg  899 bytes                        [emitted]
@keplr-wallet/extension:             assets/icons8-lock.svg   5.77 KiB                        [emitted]
@keplr-wallet/extension:              assets/icons8-pen.svg   3.27 KiB                        [emitted]
@keplr-wallet/extension:        assets/icons8-test-tube.svg   2.02 KiB                        [emitted]
@keplr-wallet/extension:        assets/icons8-trash-can.svg   1.46 KiB                        [emitted]
@keplr-wallet/extension:            assets/icons8-usb-2.svg  373 bytes                        [emitted]
@keplr-wallet/extension:               assets/info-mark.svg  618 bytes                        [emitted]
@keplr-wallet/extension:               assets/logo-temp.png   13.9 KiB                        [emitted]
@keplr-wallet/extension:            assets/nucleo-icons.eot   18.1 KiB                        [emitted]
@keplr-wallet/extension:            assets/nucleo-icons.svg    123 KiB                        [emitted]
@keplr-wallet/extension:            assets/nucleo-icons.ttf   17.9 KiB                        [emitted]
@keplr-wallet/extension:           assets/nucleo-icons.woff   9.98 KiB                        [emitted]
@keplr-wallet/extension:          assets/nucleo-icons.woff2   8.38 KiB                        [emitted]
@keplr-wallet/extension:               assets/temp-icon.svg   3.25 KiB                        [emitted]
@keplr-wallet/extension:                   assets/trash.svg   6.46 KiB                        [emitted]
@keplr-wallet/extension:               background.bundle.js    4.3 MiB            background  [emitted]        background
@keplr-wallet/extension:           background.bundle.js.map   4.13 MiB            background  [emitted] [dev]  background
@keplr-wallet/extension:                browser-polyfill.js   36.7 KiB                        [emitted]
@keplr-wallet/extension:           contentScripts.bundle.js    131 KiB        contentScripts  [emitted]        contentScripts
@keplr-wallet/extension:       contentScripts.bundle.js.map    139 KiB        contentScripts  [emitted] [dev]  contentScripts
@keplr-wallet/extension:           injectedScript.bundle.js   99.9 KiB        injectedScript  [emitted]        injectedScript
@keplr-wallet/extension:       injectedScript.bundle.js.map    109 KiB        injectedScript  [emitted] [dev]  injectedScript
@keplr-wallet/extension:                      manifest.json  953 bytes                        [emitted]
@keplr-wallet/extension:                    popup.bundle.js   12.7 MiB                 popup  [emitted]        popup
@keplr-wallet/extension:                popup.bundle.js.map   11.4 MiB                 popup  [emitted] [dev]  popup
@keplr-wallet/extension:                         popup.html  375 bytes                        [emitted]
@keplr-wallet/extension:             reactChartJS.bundle.js   15.5 KiB          reactChartJS  [emitted]        reactChartJS
@keplr-wallet/extension:         reactChartJS.bundle.js.map  393 bytes          reactChartJS  [emitted] [dev]  reactChartJS
@keplr-wallet/extension:     vendors~reactChartJS.bundle.js   1.28 MiB  vendors~reactChartJS  [emitted]        vendors~reactChartJS
@keplr-wallet/extension: vendors~reactChartJS.bundle.js.map   1.48 MiB  vendors~reactChartJS  [emitted] [dev]  vendors~reactChartJS
@keplr-wallet/extension: Entrypoint popup = popup.bundle.js popup.bundle.js.map
@keplr-wallet/extension: Entrypoint background = background.bundle.js background.bundle.js.map
@keplr-wallet/extension: Entrypoint contentScripts = contentScripts.bundle.js contentScripts.bundle.js.map
@keplr-wallet/extension: Entrypoint injectedScript = injectedScript.bundle.js injectedScript.bundle.js.map
@keplr-wallet/extension:  [0] multi ./src/index.tsx 28 bytes {popup} [built]
@keplr-wallet/extension: [13] multi ./src/background/background.ts 28 bytes {background} [built]
@keplr-wallet/extension: [14] multi ./src/content-scripts/content-scripts.ts 28 bytes {contentScripts} [built]
@keplr-wallet/extension: [15] multi ./src/content-scripts/inject/injected-script.ts 28 bytes {injectedScript} [built]
@keplr-wallet/extension:  [../background/build/index.js] 5.7 KiB {popup} {background} [built]
@keplr-wallet/extension:  [../common/build/index.js] 796 bytes {popup} {background} [built]
@keplr-wallet/extension:  [../provider/build/index.js] 796 bytes {popup} {contentScripts} {injectedScript} [built]
@keplr-wallet/extension:  [../router-extension/build/index.js] 797 bytes {popup} {background} {contentScripts} [built]
@keplr-wallet/extension:  [../router/build/index.js] 993 bytes {popup} {background} {contentScripts} {injectedScript} [built]
@keplr-wallet/extension:  [./src/background/background.ts] 2.01 KiB {background} [built]
@keplr-wallet/extension:  [./src/config.ts] 34.7 KiB {popup} {background} [built]
@keplr-wallet/extension:  [./src/config.ui.ts] 1.86 KiB {popup} [built]
@keplr-wallet/extension:  [./src/content-scripts/content-scripts.ts] 911 bytes {contentScripts} [built]
@keplr-wallet/extension:  [./src/content-scripts/inject/injected-script.ts] 390 bytes {injectedScript} [built]
@keplr-wallet/extension:  [./src/index.tsx] 7.57 KiB {popup} [built]
@keplr-wallet/extension:     + 2192 hidden modules
@keplr-wallet/extension: Child html-webpack-plugin for "popup.html":
@keplr-wallet/extension:      1 asset
@keplr-wallet/extension:     Entrypoint undefined = popup.html
@keplr-wallet/extension:     [./node_modules/html-webpack-plugin/lib/loader.js!./src/index.html] 525 bytes {0} [built]
@keplr-wallet/extension:     [./node_modules/webpack/buildin/global.js] (webpack)/buildin/global.js 472 bytes {0} [built]
@keplr-wallet/extension:     [./node_modules/webpack/buildin/module.js] (webpack)/buildin/module.js 497 bytes {0} [built]
@keplr-wallet/extension:         + 1 hidden module
```