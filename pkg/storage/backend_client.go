package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"get.porter.sh/porter/pkg/storage/plugins"
	"get.porter.sh/porter/pkg/storage/plugins/protov4"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const collectionConfig = "config"

// BackendClient adapts the domain-level protocol v4 (plugins.StorageBackend),
// which exchanges protobuf documents, to the typed documents used by Porter.
//
// It converts entities to and from their json representation along with the key
// fields that backends use to index them, and maps protocol errors, such as a
// NotFound gRPC status, to Porter's errors, such as ErrNotFound.
type BackendClient struct {
	backend plugins.StorageBackend
}

// NewBackendClient wraps the specified storage backend.
func NewBackendClient(backend plugins.StorageBackend) BackendClient {
	return BackendClient{backend: backend}
}

// Backend returns the wrapped storage backend, for operations that don't
// involve a generic document.
func (c BackendClient) Backend() plugins.StorageBackend {
	return c.backend
}

// Close the backend, when it holds resources.
func (c BackendClient) Close() error {
	if closer, ok := c.backend.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

// put the documents with the specified mode, converting each document to its
// protocol representation first.
func (c BackendClient) put(ctx context.Context, mode protov4.PutRequest_Mode, values ...interface{}) error {
	docs := make([]*protov4.Document, len(values))
	for i, v := range values {
		doc, err := toDocument(v)
		if err != nil {
			return err
		}
		docs[i] = doc
	}

	_, err := c.backend.Put(ctx, &protov4.PutRequest{Mode: mode, Documents: docs})
	if len(docs) > 0 {
		err = mapBackendError(err, docs[0].Kind, docs[0].Name)
	}
	return err
}

// getDocument retrieves a single document as the specified type.
// Returns ErrNotFound when the document doesn't exist.
func getDocument[T any](ctx context.Context, c BackendClient, req *protov4.GetRequest) (T, error) {
	var out T
	resp, err := c.backend.Get(ctx, req)
	if err != nil {
		return out, mapBackendError(err, req.Kind, req.Name)
	}
	err = unmarshalDocument(resp.Document, &out)
	return out, err
}

// listDocuments retrieves documents as the specified type.
func listDocuments[T any](ctx context.Context, c BackendClient, req *protov4.ListRequest) ([]T, error) {
	resp, err := c.backend.List(ctx, req)
	if err != nil {
		return nil, mapBackendError(err, req.Kind, "")
	}
	return unmarshalDocuments[T](resp.Documents)
}

func unmarshalDocument(doc *protov4.Document, out interface{}) error {
	if doc == nil {
		return fmt.Errorf("the storage backend returned an empty document")
	}
	if err := json.Unmarshal(doc.Json, out); err != nil {
		return fmt.Errorf("could not unmarshal document of type %T: %w", out, err)
	}
	return nil
}

func unmarshalDocuments[T any](docs []*protov4.Document) ([]T, error) {
	out := make([]T, len(docs))
	for i, doc := range docs {
		if err := unmarshalDocument(doc, &out[i]); err != nil {
			return nil, err
		}
	}
	return out, nil
}

// toDocument converts a Porter document to its protocol representation, by way
// of the json representation of the type. That ensures any custom marshaling
// logic, such as a calculated _id, is applied when persisting the document.
func toDocument(v interface{}) (*protov4.Document, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("could not marshal %T: %w", v, err)
	}

	doc := &protov4.Document{Json: data}
	switch t := v.(type) {
	case Installation:
		doc.Kind, doc.Namespace, doc.Name = protov4.Kind_KIND_INSTALLATION, t.Namespace, t.Name
	case Run:
		doc.Kind, doc.Id, doc.Namespace, doc.Installation = protov4.Kind_KIND_RUN, t.ID, t.Namespace, t.Installation
	case Result:
		doc.Kind, doc.Id, doc.Namespace, doc.Installation, doc.RunId = protov4.Kind_KIND_RESULT, t.ID, t.Namespace, t.Installation, t.RunID
	case Output:
		doc.Kind, doc.Name, doc.Namespace, doc.Installation = protov4.Kind_KIND_OUTPUT, t.Name, t.Namespace, t.Installation
		doc.RunId, doc.ResultId = t.RunID, t.ResultID
	case CredentialSet:
		doc.Kind, doc.Namespace, doc.Name = protov4.Kind_KIND_CREDENTIAL_SET, t.Namespace, t.Name
	case ParameterSet:
		doc.Kind, doc.Namespace, doc.Name = protov4.Kind_KIND_PARAMETER_SET, t.Namespace, t.Name
	case Workflow:
		doc.Kind, doc.Id, doc.Namespace = protov4.Kind_KIND_WORKFLOW, t.ID, t.Namespace
	case Schema:
		doc.Kind, doc.Id = protov4.Kind_KIND_SCHEMA, t.ID
	default:
		return nil, fmt.Errorf("unsupported document type %T", v)
	}
	return doc, nil
}

// collectionName is the name used when describing a kind of document in errors.
func collectionName(kind protov4.Kind) string {
	switch kind {
	case protov4.Kind_KIND_INSTALLATION:
		return CollectionInstallations
	case protov4.Kind_KIND_RUN:
		return CollectionRuns
	case protov4.Kind_KIND_RESULT:
		return CollectionResults
	case protov4.Kind_KIND_OUTPUT:
		return CollectionOutputs
	case protov4.Kind_KIND_CREDENTIAL_SET:
		return CollectionCredentials
	case protov4.Kind_KIND_PARAMETER_SET:
		return CollectionParameters
	case protov4.Kind_KIND_WORKFLOW:
		return CollectionWorkflows
	case protov4.Kind_KIND_SCHEMA:
		return collectionConfig
	default:
		return ""
	}
}

// mapBackendError converts a protocol error into the error that Porter's
// callers expect, for example a NotFound status into ErrNotFound.
func mapBackendError(err error, kind protov4.Kind, item string) error {
	if err == nil {
		return nil
	}
	if status.Code(err) == codes.NotFound {
		return ErrNotFound{Collection: collectionName(kind), Item: item}
	}
	return err
}
