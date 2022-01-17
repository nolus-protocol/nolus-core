import {Query, QueryClientImpl} from "../util/codec/nolus/suspend/v1beta1/query";
import {Tendermint34Client} from "@cosmjs/tendermint-rpc";
import {createProtobufRpcClient, QueryClient} from "@cosmjs/stargate";


// Documentation on custom queries - https://github.com/cosmos/cosmjs/blob/main/packages/stargate/CUSTOM_PROTOBUF_CODECS.md#step-3b-instantiate-a-query-client-using-your-custom-query-service
export async function getSuspendQueryClient(nodeUrl: string): Promise<Query> {
    const tendermintClient = await Tendermint34Client.connect(nodeUrl);

    const queryClient = new QueryClient(tendermintClient);

    const rpcClient = createProtobufRpcClient(queryClient);

    return new QueryClientImpl(rpcClient);

}