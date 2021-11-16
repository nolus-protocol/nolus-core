# Getting Started
For environment set-up and contract deployment refer to the [Getting Started](https://docs.cosmwasm.com/docs/0.16/getting-started/intro) section of the CosmWasm docs.

# Overview
Pieces related to CosmWasm contracts:
- Rust->WASM provided by [Wasmer](https://github.com/wasmerio/wasmer)
- [CosmWasmVM](https://github.com/CosmWasm/cosmwasm/tree/main/packages/vm) which is a Rust wrapper around Wasmer. This provides a WASM runtime(virtual machine) to execute smart contracts.
- [CosmWasmStd](https://github.com/CosmWasm/cosmwasm/tree/main/packages/std) which is the "standard" library which is compiled with the smart contract.
- The [wasmvm](https://github.com/CosmWasm/wasmvm) Go package, which is a Go wrapper around CosmWasmVM
- [wasmd](https://github.com/CosmWasm/wasmd) which contains Cosmos SDK modules for running CosmWasm smart contracts in `x/wasm`.

In Ethereum the arguments to the contructor of a contract are passed alongside the code deploy transaction. This results in a contract instance being created.
This is not the case in CosmWasm where code deployment and contract instantiation are separate.

Blockchain messages related to smart contracts are the following:
- `MsgStoreCode` - Sent to deploy a contract. `CodeId` is created for this contract's code. No contract state and address is created.
- `MsgInstantiateContract` - Sent to deploy an instance of a previously uploaded contract code.
Contract address is created along with some initial contract state.
- `MsgExecuteContract` - Sent to make a call to a smart contract.
- `MsgMigrateContract` - Sent by an admin that was stored in the contract on instantiation to upgrade or downgrade a contract. This message contains a `CodeId` of the new contract code. The contract's `migrate` function will perform any needed storage data transformations.

Message definitions and description can be found [here](https://github.com/CosmWasm/wasmd/blob/master/proto/cosmwasm/wasm/v1/tx.proto).

Message handling before calling into the CosmWasmVM can be found [here](https://github.com/CosmWasm/wasmd/blob/master/x/wasm/keeper/keeper.go).

The interface of the CosmWasmVM can be found [here](https://github.com/CosmWasm/wasmd/blob/master/x/wasm/types/wasmer_engine.go).

CosmWasm contracts expose the following functions:
```
// signal for 1.0 compatibility
extern "C" fn interface_version_8() -> () {}

// copy memory to/from host, so we can pass in/out Vec<u8>
extern "C" fn allocate(size: usize) -> u32;

extern "C" fn deallocate(pointer: u32);

// main contract entry points
extern "C" fn instantiate(env_ptr: u32, info_ptr: u32, msg_ptr: u32) -> u32;

extern "C" fn execute(env_ptr: u32, info_ptr: u32, msg_ptr: u32) -> u32;

extern "C" fn query(env_ptr: u32, msg_ptr: u32) -> u32;

// in-place contract migrations
extern "C" fn migrate(env_ptr: u32, info_ptr: u32, msg_ptr: u32) -> u32;

// support submessage callbacks
extern "C" fn reply(env_ptr: u32, msg_ptr: u32) -> u32;

// expose privileged entry points to Cosmos SDK modules, not external accounts
extern "C" fn sudo(env_ptr: u32, msg_ptr: u32) -> u32;

// and to write an IBC application as a contract, implement these:
extern "C" fn ibc_channel_open(env_ptr: u32, msg_ptr: u32) -> u32;

extern "C" fn ibc_channel_connect(env_ptr: u32, msg_ptr: u32) -> u32;

extern "C" fn ibc_channel_close(env_ptr: u32, msg_ptr: u32) -> u32;

extern "C" fn ibc_packet_receive(env_ptr: u32, msg_ptr: u32) -> u32;

extern "C" fn ibc_packet_ack(env_ptr: u32, msg_ptr: u32) -> u32;

extern "C" fn ibc_packet_timeout(env_ptr: u32, msg_ptr: u32) -> u32;
```

Message `MsgExecuteContract` results in a call the `execute` contract function.
Message `QuerySmartContractStateRequest`
results in a call to the `query` contract function.
Both messages carry a JSON msg to be passed to the smart contract function(entry_point).

A CosmWasm contract can return a list of messages to be executed in the same transaction. This means that a contract can request a send to happen after it has finished, or call into another contract. If any message in this list fails then the entire transaction reverts, including updates to the contract's state.

With the 0.8 CosmWasm release synchronous queries have been added allowing a contract to call another contract directly or an underlying Cosmos SDK module. These Queries only have access to a read-only database snapshot and be unable to modify state or send messages to other modules.

# Contract Layout
`src/schema/*.json` - Contains generated JSON files(ABI) to be used by clients of the smart contract. These are usually front-end applications.
`src/msg.rs` - Types to represent received JSON Msg objects.
`src/state.rs` - Types for the contract storage.
`src/error.rs` - Error enums.
`src/contract.rs` - Main contract implementation file. Contains processing of queries and messages. Defines contract entry points. 

# Execute & Query
The `execute` and `query` functions have the following signatures:
```
pub fn execute(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    msg: ExecuteMsg,
) -> Result<Response, ContractError>

pub fn query(
    deps: Deps,
    _env: Env,
    msg: QueryMsg)
-> StdResult<Binary>
```

where:
```
pub struct DepsMut<'a> {
    pub storage: &'a mut dyn Storage,
    pub api: &'a dyn Api,
    pub querier: QuerierWrapper<'a>,
}

pub trait Storage {
    fn get(&self, key: &[u8]) -> Option<Vec<u8>>;
    fn range<'a>(
        &'a self, 
        start: Option<&[u8]>, 
        end: Option<&[u8]>, 
        order: Order
    ) -> Box<dyn Iterator<Item = Pair> + 'a>;
    fn set(&mut self, key: &[u8], value: &[u8]);
    fn remove(&mut self, key: &[u8]);
}

pub trait Api {
    fn addr_validate(&self, human: &str) -> StdResult<Addr>;
    fn addr_canonicalize(&self, human: &str) -> StdResult<CanonicalAddr>;
    fn addr_humanize(&self, canonical: &CanonicalAddr) -> StdResult<Addr>;
    fn secp256k1_verify(
        &self, 
        message_hash: &[u8], 
        signature: &[u8], 
        public_key: &[u8]
    ) -> Result<bool, VerificationError>;
    fn secp256k1_recover_pubkey(
        &self, 
        message_hash: &[u8], 
        signature: &[u8], 
        recovery_param: u8
    ) -> Result<Vec<u8>, RecoverPubkeyError>;
    fn ed25519_verify(
        &self, 
        message: &[u8], 
        signature: &[u8], 
        public_key: &[u8]
    ) -> Result<bool, VerificationError>;
    fn ed25519_batch_verify(
        &self, 
        messages: &[&[u8]], 
        signatures: &[&[u8]], 
        public_keys: &[&[u8]]
    ) -> Result<bool, VerificationError>;
    fn debug(&self, message: &str);
}

pub struct Env {
    pub block: BlockInfo,
    pub contract: ContractInfo,
}

pub struct BlockInfo {
    pub height: u64,
    pub time: Timestamp,
    pub chain_id: String,
}

pub struct ContractInfo {
    pub address: Addr,
}
```

# Contract Calls
If a contract's `execute` call is successful the following response object is returned:
```
pub struct Response<T = Empty>
where
    T: Clone + fmt::Debug + PartialEq + JsonSchema,
{
    /// Optional list of "subcalls" to make. These will be executed in order
    /// (and this contract's subcall_response entry point invoked)
    /// *before* any of the "fire and forget" messages get executed.
    pub submessages: Vec<SubMsg<T>>,
    /// After any submessages are processed, these are all dispatched in the host blockchain.
    /// If they all succeed, then the transaction is committed. If any fail, then the transaction
    /// and any local contract state changes are reverted.
    pub messages: Vec<CosmosMsg<T>>,
    /// The attributes that will be emitted as part of a "wasm" event
    pub attributes: Vec<Attribute>,
    pub data: Option<Binary>,
}
```
It contains:
- a list of submessages, which are calls to modules or contracts for which the initial contract expects a reply. The reply is handled in the contract's `reply` function. If any fail the transaction is NOT necessarily reverted.
- a list of messages, which are calls to modules or contract for which the initial contract does not expect a reply. If any fail the transaction IS reverted.
- a list of attributes, which are a list of `{key, value}` pairs for the default event.
- optional data field for the transaction

```
pub struct SubMsg<T = Empty>
where
    T: Clone + fmt::Debug + PartialEq + JsonSchema,
{
    pub id: u64,
    pub msg: CosmosMsg<T>,
    pub gas_limit: Option<u64>,
    pub reply_on: ReplyOn,
}

pub enum ReplyOn {
    /// Always perform a callback after SubMsg is processed
    Always,
    /// Only callback if SubMsg returned an error, no callback on success case
    Error,
    /// Only callback if SubMsg was successful, no callback on error case
    Success,
}
```

```
pub enum CosmosMsg<T = Empty>
where
    T: Clone + fmt::Debug + PartialEq + JsonSchema,
{
    Bank(BankMsg),
    /// This can be defined by each blockchain as a custom extension
    Custom(T),
    Staking(StakingMsg),
    Distribution(DistributionMsg),
    Stargate {
        type_url: String,
        value: Binary,
    },
    Ibc(IbcMsg),
    Wasm(WasmMsg),
}
```

If a CosmWasm contract wants to call another smart contract it needs to return a message with a call to the other contract that will be executed in the same transaction.
[Example](https://github.com/certusone/wormhole/blob/dev.v2/terra/contracts/token-bridge/src/contract.rs#L687):
```
...
let mut messages = vec![CosmosMsg::Wasm(WasmMsg::Execute {
    contract_addr: contract_addr.clone(),
    msg: to_binary(&WrappedMsg::Mint {
        recipient: recipient.to_string(),
        amount: Uint128::from(amount),
    })?,
    funds: vec![],
})];
...
Ok(Response::new()
    .add_messages(messages)
...
```

If a CosmWasm contract wants to query data from another contract it can do so synchronously.
[Example](https://github.com/certusone/wormhole/blob/dev.v2/terra/contracts/token-bridge/src/contract.rs#L726):
```
...
let token_info: TokenInfoResponse =
    deps.querier.query(&QueryRequest::Wasm(WasmQuery::Smart {
        contract_addr: contract_addr.to_string(),
        msg: to_binary(&TokenQuery::TokenInfo {})?,
    }))?;
...
```

If a CosmWasm contract wants to call a native Cosmos SDK module for which a message is defined in CosmosMsg.
It can do so as this [example](https://github.com/CosmWasm/cosmwasm/blob/71f643f577184a23b2f1f122531c944f0de94c34/contracts/reflect/src/msg.rs#L30-L64):
```
...
let mut messages = vec![CosmosMsg::Bank(BankMsg::Send {
    to_address: recipient.to_string(),
    amount: coins_after_tax(deps.branch(), vec![coin(amount, &denom)])?,
})];
...
Ok(Response::new()
    .add_messages(messages)
...
```

If a CosmWasm contract wants to call a native Cosmos SDK module for which a message in not defined in CosmosMsg,
it can define a custom message similar to this example:
https://github.com/CosmWasm/cosmwasm/blob/main/contracts/reflect/src/msg.rs#L71

# Other
- A key difference to Ethereum is that sending tokens directly to a contract via SendMsg is possible but will not trigger contract code. This is design decision as it makes all contract execution be explicitly requested.

- Contract returned messages are executed depth-first. This means if contract A returns AM1 (WasmMsg::Execute) and AM2 (BankMsg::Send), and contract B (from the WasmMsg::Execute) returns BM1 and BM2 (eg. StakingMsg and DistributionMsg), the order of execution will be AM1, BM1, BM2, AM2.

- In order to enable better integrations with the native blockchain, a set of standardized module interfaces is provided. The most basic one is to the `Bank` module, which provides access to the underlying native tokens. This allows a contract to return messages `BankMsg::Send`, `BankQuery::Balance` and `BankQuery::AllBalances`.
Another standard module is staking.
