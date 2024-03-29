# Bridge Options
## Wormhole Integrated Chain <--Wormhole--> Nomo
- Wormhole V1 is on mainnet.
- Nomo needs to approach Wormhole DAO and convince them to add Nomo as a supported chain.
- Transfer transaction fees are paid by the user in ETH and Nomo.
- Users will be able to bridge between ETH, BSC, Polygon, Solana, Terra and Nomo.
- The tokens will be minted/burned directly on Nomo.
- Nomo needs to operate IBC relays to transfer wrapped tokens to/from another chain.

## ETH <--Own Wormhole--> Nomo
- Nomo can deploy their own version of Wormhole.
- Likely Ethereum <-> Nomo only support.
- Transfer transaction fees are paid by the user in ETH and Nomo.
- Nomo will need to operate the Wormhole validators.
- Each validator runs a client to observe events on each network.
- Nomo will probably own all validators(centralization).
- Nomo will need to protect the validator private keys which is a security risk.
- Nomo needs to operate IBC relays to transfer wrapped tokens to/from another chain.

## Wormhole Integrated Chain <--Wormhole--> Terra <--IBC--> Nomo
- Wormhole V1 is on mainnet.
- Transfer transaction fees are paid by the user in ETH, Nomo and Terra.
- Depends on the continuous operation and existance of Terra.
- Nomo needs to operate IBC relays to transfer tokens from/to Terra.
- Users will be able to bridge between ETH, BSC, Polygon, Solana, Terra and Nomo.
- Nomo may want to hide from users that the tokens pass through Terra.  
This can be done by deploying a forked version of the [Wormhole webapp](https://wormholebridge.com/#/transfer).  
The webapp can make the user sign a transaction that calls the Wormhole smart contract with Nomo's Terra address as recipient instead of the user's Terra address.
Then Nomo will need to operate an observer service that will initiate IBC transfers from Terra to Nomo and back.

## ETH <--Gravity--> GravityChain <--IBC--> Nomo
- Gravity does not currently have a mainnet.
- Nomo depends on the continuous existance of GravityChain.
- Transfer transaction fees are paid by the user in ETH.
- The user is minted wrapped tokens on GravityChain.
- The user can transfer tokens from/to Nomo using IBC.
- On GravityChain to transfer back to ETH the user needs to specify zero or non-zero token amount that will be subtracted from their unlocked amount and awarded to submitter of the Ethereum transaction. The submitter can be the user or a Gravity validator.
- Nomo needs to operate IBC relays to transfer wrapped tokens to/from another chain.

## ETH <--Own Gravity--> Nomo
- Gravity does not currently have a mainnet.
- Transfer transaction fees are paid by the user in ETH.
- Nomo operates a set of Gravity validators.
- Nomo secures Gravity validator private keys.
- The user is minted wrapped tokens directly on Nomo.
- Nomo operates a set of IBC relays to allow transfering tokens to another chain and back.

# Bridge fees
## ETH --Wormhole--> Terra
ETH -> Terra Deposit:  
- ERC20's `approve` - 46352 gas(0.00211 Ether/8.81 USD)  
https://etherscan.io/tx/0x9f9d129b484d1a45cf554f9b4ee33f4e29d7bcb29c937001156d6102240e9d1b  
- `wrapAndTransferETH` - 80209 gas(0.0047 Ether/19.62 USD)  
https://etherscan.io/tx/0x2b9f94089e892dda95226797a65a3ddc533792c1da1b7cb75e72ba7645de90c7  
- User receives on Terra - 0.048021 Luna  
https://finder.terra.money/columbus-5/tx/773C28FACD5368F7B656C8AEE72B35E3DFF8D27F939F6B5182806640553376C1

BSC -> Terra Deposit:  
- `approve` - 44406 gas(0.00022203 BNB/0.106840 USD)  
https://bscscan.com/tx/0xe0cb576aa114493cb6e3fe6f799be28300b1d9e7fccd2328f9c76d28e5f470d7  
- `transferTokens` - 81623 gas(0.000408115 BNB/0.20 USD)  
https://bscscan.com/tx/0x02e91b6f61f7c68a6caf5606efba02d70d2bfc276ee4b25745b58be17ace7a6f  
- User receives on Terra - 0.480287 Luna  
https://finder.terra.money/columbus-5/tx/DA626FEA7938C87FE829A79F45DD59A810C9D99BC44A553F716D151E20725483

## ETH --Gravity--> GravityChain
ETH -> GravityChain:  
- ERC20's `approve` - 46604 gas(i.e if gas is 100 gwei,  0.0046604 Ether/19.876 USD)  
- `sendToCosmos` - based on the implementation, the cost is a bit larger but similar to the above cost.  
https://github.com/cosmos/gravity-bridge/blob/main/solidity/contracts/Gravity.sol#L524

GravityChain -> ETH:  
- `MsgSendToEthereum` - user sends a transaction to the non-Ethereum chain to burn the wrapped(non-native) token.  
Gas cost is cheap because transaction is done on a Cosmos chain.  
(NOTE:  
`MsgSendToEthereum` contains `Erc20Fee` which is the amount of tokens the user wishes to offer to validators to relay to Ethereum.
This fee amount can be zero if the user will submit the batch containing this transaction to Ethereum.  
Or the fee can be some non-zero value if the user wants the validators to submit the batch containing the transaction.)
- `submitBatch` - user or a Gravity validator pay for broadcasting the batch transaction to the Gravity Ethereum contract (very large Ethereum fee).

# Wormhole
Wormhole V2 is under development, V1 is deployed to mainnet.  
In V1 Wormhole validators store the signatures of all chains' observed deposits on Solana.  
In V2 this is changed and Wormhole validators store and exchange these signatures between themselves.  
In V2 the Wormhole validators expose an API which provides the user with the signatures.  
The user can then call the Wormhole smart contract with the signatures paying the network fee.

The Wormhole bridge consists of:
- a smart contract or Cosmos module that implements  
the Wormhole protocol running on each integrated chain
- validators that sign observed deposits and withdrawals.  
These validators run a light-node to observe a chain.  
For some networks like Solana a full node is required.

Flow:  
The user initiates a `transfer` by calling either of:
- `wrapAndTransferETH` - a function that receives the caller's Ether, locks it in the contract.  
If the user wants to send the validator signatures to the other chain's smart contract, they specify a zero fee.  
If the user wants a validator to send the validator signatures to the other chain's smart contract, they specify a non-zero fee.  
A `LogMessagePublished` event is emitted.
- `transferTokens` - a function that either locks tokens (if they are native) or burns wrapped tokens (if they are not native). A `LogMessagePublished` event is emitted.

The Wormhole validators observe the smart contract for `LogMessagePublished` events.  
They sign the event's payload and exchange it between themselves using the so called "gossip" network.

If the transaction has zero fee, the user polls a Wormhole validator's API to get 2/3 validator signatures.  
Then they call the other chain's `completeTransfer` function providing the signatures.

If the transaction has non-zero fee, a Wormhole validator will call `completeTransfer` sending the signatures and claiming the fee.

The `completeTransfer` function will do either:
- mint wrapped tokens if the `transfer` was a native token deposit on the source chain
- or unlock native tokens if the `transfer` was a burn on the source chain

### DAO
If business desires to integrate Nomo with the already existing version of [Wormhole](https://wormholenetwork.com/en/),
they would need to approach the Wormhole DAO explaining why this integration should happen.  
To read more on the integration [here](https://github.com/certusone/wormhole/blob/dev.v2/CONTRIBUTING.md#contributions-faq).

# Gravity
The Gravity bridge is under development and undergoing audits.

The Gravity bridge consists of:
- an Ethereum smart contract to lock native tokens or burn wrapped tokens
- a Gravity Cosmos SDK module to integrate in the non-Ethereum chain
- A set of validators containing an `Orchestrator` and `Relayer` components
- The `Orchestrator` is responsible for observing and signing deposit events then storing their signature on the non-Ethereum chain.
- The `Relayer` creates batches of (withdraw) transactions if they decide the batch fees are favourable for them.  

Deposit Flow:  
The user first calls the token contract's `approve` function.  
Then the user calls the Gravity Ethereum smart contract's `sendToCosmos` function.  
This locks the user's tokens in the contract.  
A `SendToCosmosEvent` event is emitted.

An `Orchestrator` will observe, sign and transmit the deposit event to the non-Ethereum chain.
Once the chain receives a majority of signatures, tokens are minted to the user by the Gravity module.

Withdraw Flow:  
The user sends a `MsgSendToEthereum` transaction to the Cosmos chain.  
Specifying their receiver Ethereum address and the submitter fee.  
The user or a `Relayer` can include this transaction in a transaction batch.  
The submitter(user or `Relayer`) calls the Gravity Ethereum contract with the batch paying for the transaction but claiming sum of submitter fees for themselves.  
If the submitter wishes to sustain this process they can exchange the submitter fees back to Ether allowing them to submit new transaction batches.

# Sifchain Peggy
https://github.com/Sifchain/sifnode/blob/develop/docs/  
Peggy is a bridge that moves assets from Ethereum to Sifchain and back.  
Only transfers for whitelisted tokens are allowed.

Ethereum -> Sifchain:  
A user calls the Peggy Ethereum smart contract to lock a native token or burn a wrapped Sifchain token.  
A relayer listens for the Peggy Ethereum smart contract's `LogLock` and `LogBurn` events.  
Once such an event is observed by a relayer,  
it waits 50 blocks to ensure that the transaction is still valid,  
then submits a new `ProphecyClaim` to Sifchain.  
Once enough other relayers submit the `ProphecyClaim` to Sifchain,  
the assets are unlocked or minted to the Sifchain recipient.

Sifchain -> Ethereum:  
A user calls the Sifchain Cosmos SDK module to lock a native token or burn a wrapped Ethereum token.  
A relayer listens for Sifchain's `MsgLock` and `MsgBurn` events.  
Once such an event is observed by a relayer,  
a new `ProphecyClaim` is submitted to the Ethereum smart contract.  
Once enough other relayers submit the `ProphecyClaim` to Ethereum,  
the funds are unlocked or minted to the Ethereum recipient.

# PlantUML
[PlantUML](https://plantuml.com/) is used to generate images from text files of sequence and state diagrams.

- [PlantUML Online](http://www.plantuml.com/plantuml).

- Download the `.jar` file from [Download](https://plantuml.com/download).  
Place the `.jar` file and text files in the same directory.
Executing the `.jar` file will attempt to generate an image for each text file.

- Download the [VSCode Extension](https://marketplace.visualstudio.com/items?itemName=jebbs.plantuml).  
`Alt + D` to preview, `Ctrl + Shift + P` -> `Export Current Diagram` to generate an image from a text file.
