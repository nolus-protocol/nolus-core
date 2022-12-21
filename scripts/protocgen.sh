#!/usr/bin/env bash

#== Requirements ==
#
## make sure your `go env GOPATH` is in the `$PATH`
## Install:
## + latest buf (v1.0.0-rc11 or later)
## + protobuf v3
#
## All protoc dependencies must be installed not in the module scope
## currently we must use grpc-gateway v1
# cd ~
# go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
# go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
# go install github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway@v1.16.0
# go install github.com/cosmos/cosmos-proto/cmd/protoc-gen-go-pulsar@latest
# go get github.com/regen-network/cosmos-proto@latest # doesn't work in install mode
# go get github.com/regen-network/cosmos-proto/protoc-gen-gocosmos@v0.3.1

set -eo pipefail

echo "Generating gogo proto code"
cd proto
buf mod update
cd ..
buf generate

# move proto files to the right places
cp -r ./github.com/Nolus-Protocol/nolus-core/x/* x/  
rm -rf ./github.com/Nolus-Protocol

go mod tidy

# TODO
# **/*.pb.gw.go"github.com/golang/protobuf/descriptor" is deprecated: See the "google.golang.org/protobuf/reflect/protoreflect" package for how to obtain an EnumDescriptor or MessageDescriptor in order to programatically interact with the protobuf type system.  (SA1019)
# **/*.pb.gw.go"github.com/golang/protobuf/proto" is deprecated: Use the "google.golang.org/protobuf/proto" package instead.  (SA1019)
# **/*.pb.gw.godescriptor.ForMessage is deprecated: Not all concrete message types satisfy the Message interface. Use MessageDescriptorProto instead. If possible, the calling code should be rewritten to use protobuf reflection instead. See package "google.golang.org/protobuf/reflect/protoreflect" for details.  (SA1019)
# Temporary adding lint ignore
sed -i '1s/^/\/\/lint:file-ignore SA1019 Ignoring due to failing pipeline.\n/' ./x/mint/types/query.pb.gw.go
sed -i '1s/^/\/\/lint:file-ignore SA1019 Ignoring due to failing pipeline.\n/' ./x/tax/types/query.pb.gw.go
