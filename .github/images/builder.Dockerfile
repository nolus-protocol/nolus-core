FROM golang:1.21-alpine

ARG LEDGER_ENABLED
ENV LEDGER_ENABLED=${LEDGER_ENABLED:-false}

ARG BUILD_TAGS
ENV BUILD_TAGS=${BUILD_TAGS:-muslc}

ARG LINK_STATICALLY
ENV LINK_STATICALLY=${LINK_STATICALLY:-true}

#credit goes to https://github.com/CosmWasm/wasmd/blob/v0.27.0/Dockerfile for details on muslc build
RUN set -eux; apk add --no-cache ca-certificates build-base;

RUN apk add git

ARG WASMVM_VERSION="v1.2.6"
ARG WASMVM_LIB="libwasmvm_muslc.x86_64.a"
ARG WASMVM_BASE_URL="https://github.com/CosmWasm/wasmvm/releases/download/$WASMVM_VERSION"
ARG WASMVM_URL="$WASMVM_BASE_URL/$WASMVM_LIB"
ARG WASMVM_REL_DIR=".wasmvm"
ARG WASMVM_DIR=/go/"$WASMVM_REL_DIR"
# pointing the linker to the dir the library is stored
ENV WASMVM_DIR=${WASMVM_DIR}
ARG WASMVM_LIB_LOCAL="libwasmvm_muslc.a"
ARG WASMVM_LOCAL_PATH="$WASMVM_DIR/$WASMVM_LIB_LOCAL"
ARG WASMVM_CHECKSUM_URL="$WASMVM_BASE_URL/checksums.txt"

RUN mkdir -p $WASMVM_DIR
RUN wasmvm_lib_checksum=$(wget -O - "$WASMVM_CHECKSUM_URL")
RUN wget -O $WASMVM_LOCAL_PATH $WASMVM_URL
RUN echo "$(sha256sum "$WASMVM_LOCAL_PATH")" | grep "$wasmvm_lib_checksum"

CMD ["/bin/sh"]