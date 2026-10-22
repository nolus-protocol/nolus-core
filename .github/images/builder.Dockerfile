# Use alpine:3.24 as base image and manually download golang temporarily
# TODO: once we migrate to wasmvm 2.1.0, we can go back to using base golang image
FROM alpine:3.24

ENV PATH=/usr/local/go/bin:$PATH
COPY --from=golang:1.27.3 /usr/local/go /usr/local/go

ARG LEDGER_ENABLED
ENV LEDGER_ENABLED=${LEDGER_ENABLED:-false}

ARG BUILD_TAGS
ENV BUILD_TAGS=${BUILD_TAGS:-muslc}

ARG LINK_STATICALLY
ENV LINK_STATICALLY=${LINK_STATICALLY:-true}

#credit goes to https://github.com/CosmWasm/wasmd/blob/v0.27.0/Dockerfile for details on muslc build
RUN set -eux; apk add --no-cache ca-certificates build-base;

RUN apk add git

ARG WASMVM_VERSION="v3.0.8"
ARG WASMVM_LIB="libwasmvm_muslc.x86_64.a"
ARG WASMVM_BASE_URL="https://github.com/CosmWasm/wasmvm/releases/download/$WASMVM_VERSION"
ARG WASMVM_URL="$WASMVM_BASE_URL/$WASMVM_LIB"
ARG WASMVM_REL_DIR=".wasmvm"
ARG WASMVM_DIR=/go/"$WASMVM_REL_DIR"
# pointing the linker to the dir the library is stored
ENV WASMVM_DIR=${WASMVM_DIR}
ARG WASMVM_LIB_LOCAL="libwasmvm_muslc.x86_64.a"
ARG WASMVM_LOCAL_PATH="$WASMVM_DIR/$WASMVM_LIB_LOCAL"
ARG WASMVM_SHA256="b2299c85d49faccf3dcbb84984f30f55e8870111df98c10f017f86204d007470"

RUN mkdir -p $WASMVM_DIR
RUN set -eux; \
    wget -O "$WASMVM_LOCAL_PATH" "$WASMVM_URL"; \
    echo "${WASMVM_SHA256}  ${WASMVM_LOCAL_PATH}" | sha256sum -c -

CMD ["/bin/sh"]