# Introduce "suspend/resume" transaction processing in case of detected blockchain bugs

- Status: accepted
- Deciders: the Nolus dev&bisiness team
- Date: 2021-12-02
- Tags:

## Context and Problem Statement

If for some reason we need to pause processing of ALL transactions, reported bug in the consensus layer or CosmosSDK, or detected attack, we have to be able quickly to pause the network until analyse the case and assess its severity and exploitability.

It is worth to note, that this functionality is intended to be used only in cases when the network or most of the apps would be affected, not an isolated app. In the latter case we provide a simmilar feature on finer granularity, for example selected apps or contracts.

## Decision Drivers <!-- optional -->

- ability to enter the pausing mode immediately after the Nolus operations team receive a notification
- ability to get back to the fully operational mode once the Nolus operations team decides

## Considered Options

- Kill the binary process coordinated with the other validators
- Send a transaction to enter the pausing mode

## Decision Outcome

Chosen option: "Send a transaction to enter the pausing mode" because it could be executed quickly albeit its centralization nature.
