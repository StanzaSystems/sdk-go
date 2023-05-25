#!/bin/bash

sudo apt-get update
DEBIAN_FRONTEND=noninteractive \
    sudo apt-get -y install --no-install-recommends \
    apt-utils dialog protobuf-compiler

go install github.com/bufbuild/buf/cmd/buf@latest
go install github.com/bufbuild/connect-go/cmd/protoc-gen-connect-go@latest
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install github.com/grpc-ecosystem/grpc-health-probe@latest
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

# Because I'm tired of remembering to do this everytime I rebuild my devcontainer
echo '-w "\\n"' > ~/.curlrc
echo 'alias ll="ls -l"' >> ~/.bashrc
source ~/.bashrc