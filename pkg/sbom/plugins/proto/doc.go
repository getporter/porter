//go:generate protoc pkg/sbom/plugins/proto/sbom_generator_protocol.proto --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative --proto_path=.

// Package proto is the protobuf definition for the SBOMGeneratorProtocol
package proto
