package storage

import (
	"context"
	"errors"

	"get.porter.sh/porter/pkg/storage/plugins/protov4"
	"get.porter.sh/porter/pkg/tracing"
)

const (
	CollectionInstallations = "installations"
	CollectionRuns          = "runs"
	CollectionResults       = "results"
	CollectionOutputs       = "outputs"
)

// logsOutputName is the name of the output that holds the logs of a run.
const logsOutputName = "io.cnab.outputs.invocationImageLogs"

var _ InstallationProvider = InstallationStore{}

// InstallationStore is a persistent store for installation documents.
type InstallationStore struct {
	client  BackendClient
	encrypt EncryptionHandler
	decrypt EncryptionHandler
}

// NewInstallationStore creates a persistent store for installations using the specified
// backing datastore.
func NewInstallationStore(client BackendClient) InstallationStore {
	return InstallationStore{
		client:  client,
		encrypt: noOpEncryptionHandler,
		decrypt: noOpEncryptionHandler,
	}
}

func (s InstallationStore) ListInstallations(ctx context.Context, listOptions ListOptions) ([]Installation, error) {
	_, log := tracing.StartSpan(ctx)
	defer log.EndSpan()

	return listDocuments[Installation](ctx, s.client, listOptions.toListRequest(protov4.Kind_KIND_INSTALLATION))
}

func (s InstallationStore) ListRuns(ctx context.Context, namespace string, installation string) ([]Run, map[string][]Result, error) {
	ref := &protov4.InstallationRef{Namespace: namespace, Installation: installation}

	runsResp, err := s.client.backend.ListRuns(ctx, &protov4.ListRunsRequest{Installation: ref})
	if err != nil {
		return nil, nil, mapBackendError(err, protov4.Kind_KIND_RUN, "")
	}
	runs, err := unmarshalDocuments[Run](runsResp.Runs)
	if err != nil {
		return nil, nil, err
	}

	resultsResp, err := s.client.backend.ListResults(ctx, &protov4.ListResultsRequest{
		Scope: &protov4.ListResultsRequest_Installation{Installation: ref},
	})
	if err != nil {
		return runs, nil, mapBackendError(err, protov4.Kind_KIND_RESULT, "")
	}
	results, err := unmarshalDocuments[Result](resultsResp.Documents)
	if err != nil {
		return runs, nil, err
	}

	resultsMap := make(map[string][]Result, len(runs))
	for _, run := range runs {
		resultsMap[run.ID] = []Result{}
	}
	for _, res := range results {
		if _, ok := resultsMap[res.RunID]; ok {
			resultsMap[res.RunID] = append(resultsMap[res.RunID], res)
		}
	}

	return runs, resultsMap, nil
}

// GetActiveRuns returns runs for the installation that have not yet reached
// a terminal status (succeeded, failed, canceled), for example a run with no
// result yet, or whose last result is pending/running/unknown.
func (s InstallationStore) GetActiveRuns(ctx context.Context, namespace string, installation string) ([]Run, error) {
	resp, err := s.client.backend.GetActiveRuns(ctx, &protov4.InstallationRef{Namespace: namespace, Installation: installation})
	if err != nil {
		return nil, mapBackendError(err, protov4.Kind_KIND_RUN, "")
	}
	return unmarshalDocuments[Run](resp.Documents)
}

func (s InstallationStore) ListResults(ctx context.Context, runID string) ([]Result, error) {
	resp, err := s.client.backend.ListResults(ctx, &protov4.ListResultsRequest{
		Scope: &protov4.ListResultsRequest_RunId{RunId: runID},
	})
	if err != nil {
		return nil, mapBackendError(err, protov4.Kind_KIND_RESULT, "")
	}
	return unmarshalDocuments[Result](resp.Documents)
}

func (s InstallationStore) ListOutputs(ctx context.Context, resultID string) ([]Output, error) {
	resp, err := s.client.backend.ListOutputs(ctx, &protov4.ListOutputsRequest{
		Scope: &protov4.ListOutputsRequest_ResultId{ResultId: resultID},
	})
	if err != nil {
		return nil, mapBackendError(err, protov4.Kind_KIND_OUTPUT, "")
	}
	return unmarshalDocuments[Output](resp.Documents)
}

// FindInstallationsUsingCredentialSet finds the installations that use the
// credential set with the specified name, optionally limited to a namespace.
func (s InstallationStore) FindInstallationsUsingCredentialSet(ctx context.Context, namespace string, name string) ([]Installation, error) {
	return s.findReferencing(ctx, &protov4.FindReferencingRequest{
		Target:    &protov4.FindReferencingRequest_CredentialSet{CredentialSet: name},
		Namespace: namespace,
	})
}

// FindInstallationsUsingParameterSet finds the installations that use the
// parameter set with the specified name, optionally limited to a namespace.
func (s InstallationStore) FindInstallationsUsingParameterSet(ctx context.Context, namespace string, name string) ([]Installation, error) {
	return s.findReferencing(ctx, &protov4.FindReferencingRequest{
		Target:    &protov4.FindReferencingRequest_ParameterSet{ParameterSet: name},
		Namespace: namespace,
	})
}

// FindInstallationsReferencing finds the installations that depend on the
// specified installation, identified by its "namespace/name".
func (s InstallationStore) FindInstallationsReferencing(ctx context.Context, installation string) ([]Installation, error) {
	return s.findReferencing(ctx, &protov4.FindReferencingRequest{
		Target: &protov4.FindReferencingRequest_Installation{Installation: installation},
	})
}

func (s InstallationStore) findReferencing(ctx context.Context, req *protov4.FindReferencingRequest) ([]Installation, error) {
	_, log := tracing.StartSpan(ctx)
	defer log.EndSpan()

	resp, err := s.client.backend.FindInstallationsReferencing(ctx, req)
	if err != nil {
		return nil, mapBackendError(err, protov4.Kind_KIND_INSTALLATION, "")
	}
	return unmarshalDocuments[Installation](resp.Documents)
}

// FindDependencyCandidates finds the installations that may be shared with a
// dependent bundle. Installations in the specified namespace sort before
// installations in the global namespace.
func (s InstallationStore) FindDependencyCandidates(ctx context.Context, namespace string, repository string, digest string) ([]Installation, error) {
	_, log := tracing.StartSpan(ctx)
	defer log.EndSpan()

	resp, err := s.client.backend.FindDependencyCandidates(ctx, &protov4.FindDependencyCandidatesRequest{
		Namespace:  namespace,
		Repository: repository,
		Digest:     digest,
	})
	if err != nil {
		return nil, mapBackendError(err, protov4.Kind_KIND_INSTALLATION, "")
	}
	return unmarshalDocuments[Installation](resp.Documents)
}

func (s InstallationStore) GetInstallation(ctx context.Context, namespace string, name string) (Installation, error) {
	return getDocument[Installation](ctx, s.client, &protov4.GetRequest{
		Kind:      protov4.Kind_KIND_INSTALLATION,
		Namespace: namespace,
		Name:      name,
	})
}

func (s InstallationStore) GetRun(ctx context.Context, id string) (Run, error) {
	return getDocument[Run](ctx, s.client, &protov4.GetRequest{Kind: protov4.Kind_KIND_RUN, Id: id})
}

func (s InstallationStore) GetResult(ctx context.Context, id string) (Result, error) {
	return getDocument[Result](ctx, s.client, &protov4.GetRequest{Kind: protov4.Kind_KIND_RESULT, Id: id})
}

func (s InstallationStore) GetLastRun(ctx context.Context, namespace string, installation string) (Run, error) {
	resp, err := s.client.backend.GetLastRun(ctx, &protov4.InstallationRef{Namespace: namespace, Installation: installation})
	if err != nil {
		return Run{}, mapBackendError(err, protov4.Kind_KIND_RUN, "")
	}

	var out Run
	err = unmarshalDocument(resp.Document, &out)
	return out, err
}

func (s InstallationStore) GetLastOutput(ctx context.Context, namespace string, installation string, name string) (Output, error) {
	return s.getOutput(ctx, &protov4.GetOutputRequest{
		Scope: &protov4.GetOutputRequest_Installation{Installation: &protov4.InstallationRef{Namespace: namespace, Installation: installation}},
		Name:  name,
	})
}

func (s InstallationStore) GetLastOutputs(ctx context.Context, namespace string, installation string) (Outputs, error) {
	resp, err := s.client.backend.GetLastOutputs(ctx, &protov4.InstallationRef{Namespace: namespace, Installation: installation})
	if err != nil {
		return Outputs{}, mapBackendError(err, protov4.Kind_KIND_OUTPUT, "")
	}

	outputs, err := unmarshalDocuments[Output](resp.Documents)
	return NewOutputs(outputs), err
}

func (s InstallationStore) GetOutputs(ctx context.Context, runID string) (Outputs, error) {
	resp, err := s.client.backend.ListOutputs(ctx, &protov4.ListOutputsRequest{
		Scope: &protov4.ListOutputsRequest_RunId{RunId: runID},
	})
	if err != nil {
		return Outputs{}, mapBackendError(err, protov4.Kind_KIND_OUTPUT, "")
	}

	outputs, err := unmarshalDocuments[Output](resp.Documents)
	return NewOutputs(outputs), err
}

func (s InstallationStore) GetOutput(ctx context.Context, runID string, name string) (Output, error) {
	return s.getOutput(ctx, &protov4.GetOutputRequest{
		Scope: &protov4.GetOutputRequest_RunId{RunId: runID},
		Name:  name,
	})
}

func (s InstallationStore) getOutput(ctx context.Context, req *protov4.GetOutputRequest) (Output, error) {
	resp, err := s.client.backend.GetOutput(ctx, req)
	if err != nil {
		return Output{}, mapBackendError(err, protov4.Kind_KIND_OUTPUT, req.Name)
	}

	var out Output
	err = unmarshalDocument(resp.Document, &out)
	return out, err
}

func (s InstallationStore) GetLogs(ctx context.Context, runID string) (string, bool, error) {
	out, err := s.GetOutput(ctx, runID, logsOutputName)
	if errors.Is(err, ErrNotFound{}) {
		return "", false, nil
	}
	return string(out.Value), err == nil, err
}

func (s InstallationStore) GetLastLogs(ctx context.Context, namespace string, installation string) (string, bool, error) {
	out, err := s.GetLastOutput(ctx, namespace, installation, logsOutputName)
	if errors.Is(err, ErrNotFound{}) {
		return "", false, nil
	}
	return string(out.Value), err == nil, err
}

func (s InstallationStore) InsertInstallation(ctx context.Context, installation Installation) error {
	installation.SchemaVersion = DefaultInstallationSchemaVersion
	return s.client.put(ctx, protov4.PutRequest_INSERT, installation)
}

func (s InstallationStore) InsertRun(ctx context.Context, run Run) error {
	return s.client.put(ctx, protov4.PutRequest_INSERT, run)
}

func (s InstallationStore) InsertResult(ctx context.Context, result Result) error {
	return s.client.put(ctx, protov4.PutRequest_INSERT, result)
}

func (s InstallationStore) InsertOutput(ctx context.Context, output Output) error {
	return s.client.put(ctx, protov4.PutRequest_INSERT, output)
}

func (s InstallationStore) UpdateInstallation(ctx context.Context, installation Installation) error {
	installation.SchemaVersion = DefaultInstallationSchemaVersion
	return s.client.put(ctx, protov4.PutRequest_UPDATE, installation)
}

func (s InstallationStore) UpsertRun(ctx context.Context, run Run) error {
	return s.client.put(ctx, protov4.PutRequest_UPSERT, run)
}

func (s InstallationStore) UpsertInstallation(ctx context.Context, installation Installation) error {
	installation.SchemaVersion = DefaultInstallationSchemaVersion
	return s.client.put(ctx, protov4.PutRequest_UPSERT, installation)
}

// RemoveInstallation and all associated data.
func (s InstallationStore) RemoveInstallation(ctx context.Context, namespace string, name string) error {
	_, err := s.client.backend.RemoveInstallation(ctx, &protov4.RemoveInstallationRequest{
		Installation: &protov4.InstallationRef{Namespace: namespace, Installation: name},
	})
	return err
}

// EncryptionHandler is a function that transforms data by encrypting or decrypting it.
type EncryptionHandler func([]byte) ([]byte, error)

// noOpEncryptHandler is used when no handler is specified.
var noOpEncryptionHandler = func(data []byte) ([]byte, error) {
	return data, nil
}
