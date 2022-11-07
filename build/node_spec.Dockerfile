FROM alpine:3.14
ARG ARTIFACT_BIN=nolus.tar.gz
ARG CUSTOM_MONIKER=docker_generated_node
ARG ACCESS_TOKEN=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

COPY $ARTIFACT_BIN /tmp/
RUN tar -xvf /tmp/$ARTIFACT_BIN --directory /usr/bin/
RUN rm /tmp/$ARTIFACT_BIN

RUN wget -O genesis.json --header="Authorization: Token $ACCESS_TOKEN" https://raw.githubusercontent.com/Nolus-Protocol/Wiki/main/testnet-rila/genesis.json
RUN wget -O persistent_peers.txt --header="Authorization: Token $ACCESS_TOKEN" https://raw.githubusercontent.com/Nolus-Protocol/Wiki/main/testnet-rila/persistent_peers.txt

COPY /scripts/docker-node.sh docker-node.sh
RUN chmod +x docker-node.sh

# tendermint p2p
EXPOSE 26656

ENTRYPOINT /docker-node.sh ; nolusd start