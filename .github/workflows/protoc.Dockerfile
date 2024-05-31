FROM golang:1.22
RUN apt-get update && apt-get -y install protobuf-compiler
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2
WORKDIR /proto
