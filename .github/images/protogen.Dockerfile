FROM ghcr.io/cosmos/proto-builder:0.16.0

USER root

RUN apk update && apk add --no-cache \
  nodejs \
  npm \
  git \
  make \
  python3 \
  jq \
  bash

RUN npm install -g swagger-combine
RUN npm install -g swagger-merger
RUN npm install -g swagger2openapi

COPY --from=golang:1.23-alpine /usr/local/go/ /usr/local/go/

ENV PATH="/usr/local/go/bin:${PATH}"

# Create a non-root user and switch to it
ARG USERNAME=builder
ARG USER_UID=10001
ARG USER_GID=10001

RUN addgroup -g ${USER_GID} ${USERNAME} && \
    adduser -D -u ${USER_UID} -G ${USERNAME} ${USERNAME}

USER ${USERNAME}