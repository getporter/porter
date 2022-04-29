//go:generate protoc pkg/storage/plugins/proto/storage_protocol.proto --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative --proto_path=.

// Package proto is the protobuf definition for the StorageProtocol
package proto
