//go:generate protoc pkg/signing/plugins/proto/signing_protocol.proto --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative --proto_path=.

// Package proto is the protobuf definition for the SigningProtocol
package proto
