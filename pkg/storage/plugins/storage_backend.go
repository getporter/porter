package plugins

import (
	"context"

	"get.porter.sh/porter/pkg/storage/plugins/protov4"
)

// PluginProtocolVersionV4 is the version of the domain-level storage protocol
// defined by protov4.
const PluginProtocolVersionV4 = 4

// StorageBackend is the interface that storage plugins implement. It has no
// query language: each method is a domain operation that a backend implements
// natively (a pipeline in MongoDB, a join in PostgreSQL).
//
// The request and response types are the protov4 messages, so the Go contract
// and the wire contract cannot drift. Methods return a gRPC status error with
// codes.NotFound, codes.AlreadyExists, etc. See ErrNotFound handling in the
// pluginstore client.
//
// The method set mirrors the protov4.Storage service. See storage_v4.proto for
// the semantics of each operation.
type StorageBackend interface {
	// Info is the handshake.
	Info(ctx context.Context, req *protov4.InfoRequest) (*protov4.InfoResponse, error)
	Migrate(ctx context.Context, req *protov4.MigrateRequest) (*protov4.MigrateResponse, error)

	Put(ctx context.Context, req *protov4.PutRequest) (*protov4.PutResponse, error)
	Get(ctx context.Context, req *protov4.GetRequest) (*protov4.GetResponse, error)
	List(ctx context.Context, req *protov4.ListRequest) (*protov4.ListResponse, error)
	Remove(ctx context.Context, req *protov4.RemoveRequest) (*protov4.RemoveResponse, error)

	RemoveInstallation(ctx context.Context, req *protov4.RemoveInstallationRequest) (*protov4.RemoveInstallationResponse, error)
	ListRuns(ctx context.Context, req *protov4.ListRunsRequest) (*protov4.ListRunsResponse, error)
	GetLastRun(ctx context.Context, req *protov4.InstallationRef) (*protov4.GetResponse, error)
	GetActiveRuns(ctx context.Context, req *protov4.InstallationRef) (*protov4.ListResponse, error)
	ListResults(ctx context.Context, req *protov4.ListResultsRequest) (*protov4.ListResponse, error)
	InsertOutputs(ctx context.Context, req *protov4.InsertOutputsRequest) (*protov4.InsertOutputsResponse, error)
	ListOutputs(ctx context.Context, req *protov4.ListOutputsRequest) (*protov4.ListResponse, error)
	GetOutput(ctx context.Context, req *protov4.GetOutputRequest) (*protov4.GetResponse, error)
	GetLastOutputs(ctx context.Context, req *protov4.InstallationRef) (*protov4.ListResponse, error)

	FindInstallationsReferencing(ctx context.Context, req *protov4.FindReferencingRequest) (*protov4.ListResponse, error)
	FindDependencyCandidates(ctx context.Context, req *protov4.FindDependencyCandidatesRequest) (*protov4.ListResponse, error)
}

// The generated gRPC server interface must stay in lockstep with StorageBackend:
// every StorageServer is usable as a StorageBackend.
var _ StorageBackend = protov4.StorageServer(nil)
