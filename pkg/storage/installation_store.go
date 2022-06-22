package storage

import (
	"context"

	"get.porter.sh/porter/pkg/tracing"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
)

const (
	CollectionInstallations = "installations"
	CollectionRuns          = "runs"
	CollectionResults       = "results"
	CollectionOutputs       = "outputs"
)

var _ InstallationProvider = InstallationStore{}

// InstallationStore is a persistent store for installation documents.
type InstallationStore struct {
	store   Store
	encrypt EncryptionHandler
	decrypt EncryptionHandler
}

// NewInstallationStore creates a persistent store for installations using the specified
// backing datastore.
func NewInstallationStore(datastore Store) InstallationStore {
	return InstallationStore{
		store:   datastore,
		encrypt: noOpEncryptionHandler,
		decrypt: noOpEncryptionHandler,
	}
}

// EnsureInstallationIndices created indices on the installations collection.
func EnsureInstallationIndices(ctx context.Context, store Store) error {
	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()

	span.Debug("Initializing installation collection indices")

	opts := EnsureIndexOptions{
		Indices: []Index{
			// query installations by a namespace (list) or namespace + name (get)
			{Collection: CollectionInstallations, Keys: []string{"namespace", "name"}, Unique: true},
			// query runs by installation (list)
			{Collection: CollectionRuns, Keys: []string{"namespace", "installation"}},
			// query results by installation (delete or batch get)
			{Collection: CollectionResults, Keys: []string{"namespace", "installation"}},
			// query results by run (list)
			{Collection: CollectionResults, Keys: []string{"runId"}},
			// query most recent outputs by run (porter installation run show, when we list outputs)
			{Collection: CollectionOutputs, Keys: []string{"namespace", "installation", "-resultId"}},
			// query outputs by result (list)
			{Collection: CollectionOutputs, Keys: []string{"resultId", "name"}, Unique: true},
			// query most recent outputs by name for an installation
			{Collection: CollectionOutputs, Keys: []string{"namespace", "installation", "name", "-resultId"}},
		},
	}

	err := store.EnsureIndex(ctx, opts)
	return span.Error(err)
}

func (s InstallationStore) ListInstallations(ctx context.Context, listOptions ListOptions) ([]Installation, error) {
	_, log := tracing.StartSpan(ctx)
	defer log.EndSpan()

	var out []Installation
	err := s.store.Find(ctx, CollectionInstallations, listOptions.ToFindOptions(), &out)
	return out, err
}

func (s InstallationStore) ListRuns(ctx context.Context, namespace string, installation string) ([]Run, map[string][]Result, error) {
	var runs []Run
	var err error
	var results []Result

	opts := FindOptions{
		Sort: []string{"_id"},
		Filter: bson.M{
			"namespace":    namespace,
			"installation": installation,
		},
	}
	err = s.store.Find(ctx, CollectionRuns, opts, &runs)
	if err != nil {
		return nil, nil, err
	}

	err = s.store.Find(ctx, CollectionResults, opts, &results)
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

	return runs, resultsMap, err
}

func (s InstallationStore) ListResults(ctx context.Context, runID string) ([]Result, error) {
	var out []Result
	opts := FindOptions{
		Sort: []string{"_id"},
		Filter: bson.M{
			"runId": runID,
		},
	}
	err := s.store.Find(ctx, CollectionResults, opts, &out)
	return out, err
}

func (s InstallationStore) ListOutputs(ctx context.Context, resultID string) ([]Output, error) {
	var out []Output
	opts := FindOptions{
		Sort: []string{"resultId", "name"},
		Filter: bson.M{
			"resultId": resultID,
		},
	}
	err := s.store.Find(ctx, CollectionOutputs, opts, &out)
	return out, err
}

func (s InstallationStore) FindInstallations(ctx context.Context, findOpts FindOptions) ([]Installation, error) {
	_, log := tracing.StartSpan(ctx)
	defer log.EndSpan()

	var out []Installation
	err := s.store.Find(ctx, CollectionInstallations, findOpts, &out)
	return out, err
}

func (s InstallationStore) GetInstallation(ctx context.Context, namespace string, name string) (Installation, error) {
	var out Installation

	opts := FindOptions{
		Filter: bson.M{
			"namespace": namespace,
			"name":      name,
		},
	}
	err := s.store.FindOne(ctx, CollectionInstallations, opts, &out)
	return out, err
}

func (s InstallationStore) GetRun(ctx context.Context, id string) (Run, error) {
	var out Run
	opts := GetOptions{ID: id}
	err := s.store.Get(ctx, CollectionRuns, opts, &out)
	return out, err
}

func (s InstallationStore) GetResult(ctx context.Context, id string) (Result, error) {
	var out Result
	opts := GetOptions{ID: id}
	err := s.store.Get(ctx, CollectionResults, opts, &out)
	return out, err
}

func (s InstallationStore) GetLastRun(ctx context.Context, namespace string, installation string) (Run, error) {
	var out []Run
	opts := FindOptions{
		Sort:  []string{"-_id"},
		Limit: 1,
		Filter: bson.M{
			"namespace":    namespace,
			"installation": installation,
		},
	}
	err := s.store.Find(ctx, CollectionRuns, opts, &out)
	if err != nil {
		return Run{}, err
	}
	if len(out) == 0 {
		return Run{}, ErrNotFound{Collection: CollectionRuns}
	}
	return out[0], err
}

func (s InstallationStore) GetLastOutput(ctx context.Context, namespace string, installation string, name string) (Output, error) {
	var out Output
	opts := FindOptions{
		Sort:  []string{"-_id"},
		Limit: 1,
		Filter: bson.M{
			"namespace":    namespace,
			"installation": installation,
			"name":         name,
		},
	}
	err := s.store.FindOne(ctx, CollectionOutputs, opts, &out)
	return out, err
}

func (s InstallationStore) GetLastOutputs(ctx context.Context, namespace string, installation string) (Outputs, error) {
	var groupedOutputs []struct {
		ID         string `json:"_id"`
		LastOutput Output `json:"lastOutput"`
	}
	opts := AggregateOptions{
		Pipeline: []bson.D{
			// List outputs by installation
			{{Key: "$match", Value: bson.M{
				"namespace":    namespace,
				"installation": installation,
			}}},
			// Reverse sort them (newest on top)
			{{Key: "$sort", Value: bson.D{
				{Key: "namespace", Value: 1},
				{Key: "installation", Value: 1},
				{Key: "name", Value: 1},
				{Key: "resultId", Value: -1},
			}}},
			// Group them by output name and select the last value for each output
			{{Key: "$group", Value: bson.D{
				{Key: "_id", Value: "$name"},
				{Key: "lastOutput", Value: bson.M{"$first": "$$ROOT"}},
			}}},
		},
	}
	err := s.store.Aggregate(ctx, CollectionOutputs, opts, &groupedOutputs)

	lastOutputs := make([]Output, len(groupedOutputs))
	for i, groupedOutput := range groupedOutputs {
		lastOutputs[i] = groupedOutput.LastOutput
	}

	return NewOutputs(lastOutputs), err
}

func (s InstallationStore) GetLogs(ctx context.Context, runID string) (string, bool, error) {
	var out Output
	opts := FindOptions{
		Sort: []string{"resultId"}, // get logs from the last result for a run
		Filter: bson.M{
			"runId": runID,
			"name":  "io.cnab.outputs.invocationImageLogs",
		},
		Limit: 1,
	}
	err := s.store.FindOne(ctx, CollectionOutputs, opts, &out)
	if errors.Is(err, ErrNotFound{}) {
		return "", false, nil
	}
	return string(out.Value), err == nil, err
}

func (s InstallationStore) GetLastLogs(ctx context.Context, namespace string, installation string) (string, bool, error) {
	var out Output
	opts := FindOptions{
		Sort: []string{"-resultId"}, // get logs from the last result for a run
		Filter: bson.M{
			"namespace":    namespace,
			"installation": installation,
			"name":         "io.cnab.outputs.invocationImageLogs",
		},
		Limit: 1,
	}
	err := s.store.FindOne(ctx, CollectionOutputs, opts, &out)
	if errors.Is(err, ErrNotFound{}) {
		return "", false, nil
	}
	return string(out.Value), err == nil, err
}

func (s InstallationStore) InsertInstallation(ctx context.Context, installation Installation) error {
	installation.SchemaVersion = InstallationSchemaVersion
	opts := InsertOptions{
		Documents: []interface{}{installation},
	}
	return s.store.Insert(ctx, CollectionInstallations, opts)
}

func (s InstallationStore) InsertRun(ctx context.Context, run Run) error {
	opts := InsertOptions{
		Documents: []interface{}{run},
	}
	return s.store.Insert(ctx, CollectionRuns, opts)
}

func (s InstallationStore) InsertResult(ctx context.Context, result Result) error {
	opts := InsertOptions{
		Documents: []interface{}{result},
	}
	return s.store.Insert(ctx, CollectionResults, opts)
}

func (s InstallationStore) InsertOutput(ctx context.Context, output Output) error {
	opts := InsertOptions{
		Documents: []interface{}{output},
	}
	return s.store.Insert(ctx, CollectionOutputs, opts)
}

func (s InstallationStore) UpdateInstallation(ctx context.Context, installation Installation) error {
	installation.SchemaVersion = InstallationSchemaVersion
	opts := UpdateOptions{
		Document: installation,
	}
	return s.store.Update(ctx, CollectionInstallations, opts)
}

func (s InstallationStore) UpsertRun(ctx context.Context, run Run) error {
	opts := UpdateOptions{
		Upsert:   true,
		Document: run,
	}
	return s.store.Update(ctx, CollectionRuns, opts)
}

func (s InstallationStore) UpsertInstallation(ctx context.Context, installation Installation) error {
	installation.SchemaVersion = InstallationSchemaVersion
	opts := UpdateOptions{
		Upsert:   true,
		Document: installation,
	}
	return s.store.Update(ctx, CollectionInstallations, opts)
}

// RemoveInstallation and all associated data.
func (s InstallationStore) RemoveInstallation(ctx context.Context, namespace string, name string) error {
	removeInstallation := RemoveOptions{
		Filter: bson.M{
			"namespace": namespace,
			"name":      name,
		},
	}
	err := s.store.Remove(ctx, CollectionInstallations, removeInstallation)
	if err != nil {
		return err
	}

	// Find associated documents
	removeChildDocs := RemoveOptions{
		Filter: bson.M{
			"namespace":    namespace,
			"installation": name,
		},
		All: true,
	}

	// Delete runs
	err = s.store.Remove(ctx, CollectionRuns, removeChildDocs)
	if err != nil {
		return err
	}

	// Delete results
	err = s.store.Remove(ctx, CollectionResults, removeChildDocs)
	if err != nil {
		return err
	}

	// Delete outputs
	err = s.store.Remove(ctx, CollectionOutputs, removeChildDocs)
	if err != nil {
		return err
	}

	return nil
}

// EncryptionHandler is a function that transforms data by encrypting or decrypting it.
type EncryptionHandler func([]byte) ([]byte, error)

// noOpEncryptHandler is used when no handler is specified.
var noOpEncryptionHandler = func(data []byte) ([]byte, error) {
	return data, nil
}
