import { isBroadcastTxFailure } from "@cosmjs/stargate";
import { getValidatorClient, getValidatorWallet } from "../util/clients";
import { TextProposal } from "cosmjs-types/cosmos/gov/v1beta1/gov";
import { ParameterChangeProposal } from "cosmjs-types/cosmos/params/v1beta1/params";
import { CommunityPoolSpendProposal } from "cosmjs-types/cosmos/distribution/v1beta1/distribution";
import { SoftwareUpgradeProposal, CancelSoftwareUpgradeProposal } from "cosmjs-types/cosmos/upgrade/v1beta1/upgrade";
import { UpgradeProposal, ClientUpdateProposal } from "../util/proposals"
import { ClientState } from "cosmjs-types/ibc/lightclients/tendermint/v1/tendermint"
import Long from "long";

describe('proposal submission', () => {

    test('validator cannot submit a Text proposal', async () => {
        const client = await getValidatorClient();
        const wallet = await getValidatorWallet();
        const [firstAccount] = await wallet.getAccounts();
        const msg = {
            typeUrl: "/cosmos.gov.v1beta1.MsgSubmitProposal",
            value: {
                content: {
                    typeUrl: "/cosmos.gov.v1beta1.TextProposal",
                    value: TextProposal.encode({
                        description: "This proposal proposes to test whether this proposal passes",
                        title: "Test Proposal",
                    }).finish(),
                },
                proposer: firstAccount.address,
                initialDeposit: [{denom: "nomo", amount: "12"}],
            }
        };

        const fee = {
            amount: [{denom: "nomo", amount: "12"}],
            gas: "100000"
        }

        const result = await client.signAndBroadcast(firstAccount.address, [msg], fee);
        expect(isBroadcastTxFailure(result)).toBeTruthy();
        expect(result.rawLog).toEqual("failed to execute message; message index: 0: gov: no handler exists for proposal type");
    })

    test('validator cannot submit a CommunityPoolSpend proposal', async () => {
        const client = await getValidatorClient();
        const wallet = await getValidatorWallet();
        const [firstAccount] = await wallet.getAccounts();
        const msg = {
            typeUrl: "/cosmos.gov.v1beta1.MsgSubmitProposal",
            value: {
                content: {
                    typeUrl: "/cosmos.distribution.v1beta1.CommunityPoolSpendProposal",
                    value: CommunityPoolSpendProposal.encode({
                        description: "This proposal proposes to test whether this proposal passes",
                        title: "Test Proposal",
                        recipient: firstAccount.address,
                        amount: [{denom: "nomo", amount: "1000000"}]
                    }).finish(),
                },
                proposer: firstAccount.address,
                initialDeposit: [{denom: "nomo", amount: "12"}],
            }
        };

        const fee = {
            amount: [{denom: "nomo", amount: "12"}],
            gas: "100000"
        }

        const result = await client.signAndBroadcast(firstAccount.address, [msg], fee);
        expect(isBroadcastTxFailure(result)).toBeTruthy();
        expect(result.rawLog).toEqual("failed to execute message; message index: 0: distribution: no handler exists for proposal type");
    })

    test('validator cannot submit a ParameterChange proposal', async () => {
        const client = await getValidatorClient();
        const wallet = await getValidatorWallet();
        const [firstAccount] = await wallet.getAccounts();
        const msg = {
            typeUrl: "/cosmos.gov.v1beta1.MsgSubmitProposal",
            value: {
                content: {
                    typeUrl: "/cosmos.params.v1beta1.ParameterChangeProposal",
                    value: ParameterChangeProposal.encode({
                        description: "This proposal proposes to test whether this proposal passes",
                        title: "Test Proposal",
                        changes: [{
                            subspace: "subspace",
                            key: "key",
                            value: "value"
                        }]
                    }).finish(),
                },
                proposer: firstAccount.address,
                initialDeposit: [{denom: "nomo", amount: "12"}],
            }
        };

        const fee = {
            amount: [{denom: "nomo", amount: "12"}],
            gas: "100000"
        }

        const result = await client.signAndBroadcast(firstAccount.address, [msg], fee);
        expect(isBroadcastTxFailure(result)).toBeTruthy();
        expect(result.rawLog).toEqual("failed to execute message; message index: 0: params: no handler exists for proposal type");
    })

    test('validator cannot submit a SoftwareUpgrade proposal', async () => {
        const client = await getValidatorClient();
        const wallet = await getValidatorWallet();
        const [firstAccount] = await wallet.getAccounts();
        const msg = {
            typeUrl: "/cosmos.gov.v1beta1.MsgSubmitProposal",
            value: {
                content: {
                    typeUrl: "/cosmos.upgrade.v1beta1.SoftwareUpgradeProposal",
                    value: SoftwareUpgradeProposal.encode({
                        description: "This proposal proposes to test whether this proposal passes",
                        title: "Test Proposal",
                        plan: {
                            name: "Upgrade 1",
                            height: Long.fromInt(10000),
                            info: ""
                        }
                    }).finish(),
                },
                proposer: firstAccount.address,
                initialDeposit: [{denom: "nomo", amount: "12"}],
            }
        };

        const fee = {
            amount: [{denom: "nomo", amount: "12"}],
            gas: "100000"
        }

        const result = await client.signAndBroadcast(firstAccount.address, [msg], fee);
        expect(isBroadcastTxFailure(result)).toBeTruthy();
        expect(result.rawLog).toEqual("failed to execute message; message index: 0: upgrade: no handler exists for proposal type");
    })

    test('validator cannot submit a CancelSoftwareUpgrade proposal', async () => {
        const client = await getValidatorClient();
        const wallet = await getValidatorWallet();
        const [firstAccount] = await wallet.getAccounts();
        const msg = {
            typeUrl: "/cosmos.gov.v1beta1.MsgSubmitProposal",
            value: {
                content: {
                    typeUrl: "/cosmos.upgrade.v1beta1.CancelSoftwareUpgradeProposal",
                    value: CancelSoftwareUpgradeProposal.encode({
                        description: "This proposal proposes to test whether this proposal passes",
                        title: "Test Proposal",
                    }).finish(),
                },
                proposer: firstAccount.address,
                initialDeposit: [{denom: "nomo", amount: "12"}],
            }
        };

        const fee = {
            amount: [{denom: "nomo", amount: "12"}],
            gas: "100000"
        }

        const result = await client.signAndBroadcast(firstAccount.address, [msg], fee);
        expect(isBroadcastTxFailure(result)).toBeTruthy();
        expect(result.rawLog).toEqual("failed to execute message; message index: 0: upgrade: no handler exists for proposal type");
    })

    test('validator cannot submit an IBC Upgrade proposal', async () => {
        const client = await getValidatorClient();
        const wallet = await getValidatorWallet();
        const [firstAccount] = await wallet.getAccounts();
        const msg = {
            typeUrl: "/cosmos.gov.v1beta1.MsgSubmitProposal",
            value: {
                content: {
                    typeUrl: "/ibc.core.client.v1.UpgradeProposal",
                    value: UpgradeProposal.encode({
                        description: "This proposal proposes to test whether this proposal passes",
                        title: "Test Proposal",
                        plan: {
                            name: "Upgrade 1",
                            height: Long.fromInt(10000),
                            info: ""
                        },
                        upgradedClientState: {
                            typeUrl: "/ibc.lightclients.tendermint.v1.ClientState",
                            value: ClientState.encode({
                                chainId: "nomo-private",
                                proofSpecs: [{minDepth: 0, maxDepth: 0}],
                                upgradePath: ["upgrade", "upgradedIBCState"],
                                allowUpdateAfterExpiry: true,
                                allowUpdateAfterMisbehaviour: true
                            }).finish()
                        }
                    }).finish(),
                },
                proposer: firstAccount.address,
                initialDeposit: [{denom: "nomo", amount: "12"}],
            }
        };

        const fee = {
            amount: [{denom: "nomo", amount: "12"}],
            gas: "200000"
        }

        const result = await client.signAndBroadcast(firstAccount.address, [msg], fee);
        expect(isBroadcastTxFailure(result)).toBeTruthy();
        expect(result.rawLog).toEqual("failed to execute message; message index: 0: client: no handler exists for proposal type");
    })

    test('validator cannot submit a ClientUpgrade proposal', async () => {
        const client = await getValidatorClient();
        const wallet = await getValidatorWallet();
        const [firstAccount] = await wallet.getAccounts();
        const msg = {
            typeUrl: "/cosmos.gov.v1beta1.MsgSubmitProposal",
            value: {
                content: {
                    typeUrl: "/ibc.core.client.v1.ClientUpdateProposal",
                    value: ClientUpdateProposal.encode({
                        description: "This proposal proposes to test whether this proposal passes",
                        title: "Test Proposal",
                        subjectClientId: "tendermint-07",
                        substituteClientId: "tendermint-08"
                    }).finish(),
                },
                proposer: firstAccount.address,
                initialDeposit: [{denom: "nomo", amount: "12"}],
            }
        };

        const fee = {
            amount: [{denom: "nomo", amount: "12"}],
            gas: "100000"
        }

        const result = await client.signAndBroadcast(firstAccount.address, [msg], fee);
        expect(isBroadcastTxFailure(result)).toBeTruthy();
        expect(result.rawLog).toEqual("failed to execute message; message index: 0: client: no handler exists for proposal type");
    })
})