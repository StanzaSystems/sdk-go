# https://github.com/devcontainers/images/blob/main/src/go
FROM mcr.microsoft.com/devcontainers/go:1.19

# [Optional] Uncomment this section to install additional OS packages.
RUN apt-get update && export DEBIAN_FRONTEND=noninteractive \
    && apt-get -y install --no-install-recommends apt-utils dialog dnsutils protobuf-compiler

# [Optional] Uncomment this section to install anything else you need
USER vscode
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2
RUN go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
RUN go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest
RUN go install github.com/fullstorydev/grpcurl/cmd/grpcurl@v1.8.7
RUN go install github.com/bufbuild/buf/cmd/buf@v1.11.0

# [Optional] Uncomment this line to install global node packages.
# RUN su vscode -c "source /usr/local/share/nvm/nvm.sh && npm install -g <your-package-here>" 2>&1