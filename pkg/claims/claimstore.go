package claims

import (
	"context"

	"get.porter.sh/porter/pkg/storage"
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

var _ Provider = ClaimStore{}

// ClaimStore is a persistent store for claims.
type ClaimStore struct {
	store   storage.Store
	encrypt EncryptionHandler
	decrypt EncryptionHandler
}

// NewClaimStore creates a persistent store for claims using the specified
// backing datastore.
func NewClaimStore(datastore storage.Store) ClaimStore {
	return ClaimStore{
		store:   datastore,
		encrypt: noOpEncryptionHandler,
		decrypt: noOpEncryptionHandler,
	}
}

func (s ClaimStore) Initialize() error {
	opts := storage.EnsureIndexOptions{
		Indices: []storage.Index{
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

	return s.store.EnsureIndex(opts)
}

func (s ClaimStore) ListInstallations(ctx context.Context, namespace string, name string, labels map[string]string) ([]Installation, error) {
	_, log := tracing.StartSpan(ctx)
	defer log.EndSpan()

	var out []Installation
	findOpts := storage.FindOptions{
		Sort:   []string{"namespace", "name"},
		Filter: storage.CreateListFiler(namespace, name, labels),
	}

	err := s.store.Find(CollectionInstallations, findOpts, &out)
	return out, err
}

func (s ClaimStore) ListRuns(namespace string, installation string) ([]Run, map[string][]Result, error) {
	var runs []Run
	var err error
	var results []Result

	opts := storage.FindOptions{
		Sort: []string{"_id"},
		Filter: bson.M{
			"namespace":    namespace,
			"installation": installation,
		},
	}
	err = s.store.Find(CollectionRuns, opts, &runs)
	if err != nil {
		return nil, nil, err
	}

	err = s.store.Find(CollectionResults, opts, &results)
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

func (s ClaimStore) ListResults(runID string) ([]Result, error) {
	var out []Result
	opts := storage.FindOptions{
		Sort: []string{"_id"},
		Filter: bson.M{
			"runId": runID,
		},
	}
	err := s.store.Find(CollectionResults, opts, &out)
	return out, err
}

func (s ClaimStore) ListOutputs(resultID string) ([]Output, error) {
	var out []Output
	opts := storage.FindOptions{
		Sort: []string{"resultId", "name"},
		Filter: bson.M{
			"resultId": resultID,
		},
	}
	err := s.store.Find(CollectionOutputs, opts, &out)
	return out, err
}

func (s ClaimStore) GetInstallation(namespace string, name string) (Installation, error) {
	var out Installation
	opts := storage.FindOptions{
		Filter: bson.M{
			"namespace": namespace,
			"name":      name,
		},
	}
	err := s.store.FindOne(CollectionInstallations, opts, &out)
	return out, err
}

func (s ClaimStore) GetRun(id string) (Run, error) {
	var out Run
	opts := storage.GetOptions{ID: id}
	err := s.store.Get(CollectionRuns, opts, &out)
	return out, err
}

func (s ClaimStore) GetResult(id string) (Result, error) {
	var out Result
	opts := storage.GetOptions{ID: id}
	err := s.store.Get(CollectionResults, opts, &out)
	return out, err
}

func (s ClaimStore) GetLastRun(namespace string, installation string) (Run, error) {
	var out []Run
	opts := storage.FindOptions{
		Sort:  []string{"-_id"},
		Limit: 1,
		Filter: bson.M{
			"namespace":    namespace,
			"installation": installation,
		},
	}
	err := s.store.Find(CollectionRuns, opts, &out)
	if err != nil {
		return Run{}, err
	}
	if len(out) == 0 {
		return Run{}, storage.ErrNotFound{Collection: CollectionRuns}
	}
	return out[0], err
}

func (s ClaimStore) GetLastOutput(namespace string, installation string, name string) (Output, error) {
	var out Output
	opts := storage.FindOptions{
		Sort:  []string{"-_id"},
		Limit: 1,
		Filter: bson.M{
			"namespace":    namespace,
			"installation": installation,
			"name":         name,
		},
	}
	err := s.store.FindOne(CollectionOutputs, opts, &out)
	return out, err
}

func (s ClaimStore) GetLastOutputs(namespace string, installation string) (Outputs, error) {
	var groupedOutputs []struct {
		ID         string `json:"_id"`
		LastOutput Output `json:"lastOutput"`
	}
	opts := storage.AggregateOptions{
		Pipeline: []bson.M{
			// List outputs by installation
			{"$match": bson.M{
				"namespace":    namespace,
				"installation": installation,
			}},
			// Reverse sort them (newest on top)
			{"$sort": bson.M{
				"namespace":    1,
				"installation": 1,
				"name":         1,
				"resultId":     -1,
			}},
			// Group them by output name and select the last value for each output
			{"$group": bson.M{
				"_id":        "$name",
				"lastOutput": bson.M{"$first": "$$ROOT"},
			}},
		},
	}
	err := s.store.Aggregate(CollectionOutputs, opts, &groupedOutputs)

	lastOutputs := make([]Output, len(groupedOutputs))
	for i, groupedOutput := range groupedOutputs {
		lastOutputs[i] = groupedOutput.LastOutput
	}

	return NewOutputs(lastOutputs), err
}

func (s ClaimStore) GetLogs(runID string) (string, bool, error) {
	var out Output
	opts := storage.FindOptions{
		Sort: []string{"resultId"}, // get logs from the last result for a run
		Filter: bson.M{
			"runId": runID,
			"name":  "io.cnab.outputs.invocationImageLogs",
		},
		Limit: 1,
	}
	err := s.store.FindOne(CollectionOutputs, opts, &out)
	if errors.Is(err, storage.ErrNotFound{}) {
		return "", false, nil
	}
	return string(out.Value), err == nil, err
}

func (s ClaimStore) GetLastLogs(namespace string, installation string) (string, bool, error) {
	var out Output
	opts := storage.FindOptions{
		Sort: []string{"-resultId"}, // get logs from the last result for a run
		Filter: bson.M{
			"namespace":    namespace,
			"installation": installation,
			"name":         "io.cnab.outputs.invocationImageLogs",
		},
		Limit: 1,
	}
	err := s.store.FindOne(CollectionOutputs, opts, &out)
	if errors.Is(err, storage.ErrNotFound{}) {
		return "", false, nil
	}
	return string(out.Value), err == nil, err
}

func (s ClaimStore) InsertInstallation(installation Installation) error {
	installation.SchemaVersion = SchemaVersion
	opts := storage.InsertOptions{
		Documents: []interface{}{installation},
	}
	return s.store.Insert(CollectionInstallations, opts)
}

func (s ClaimStore) InsertRun(run Run) error {
	opts := storage.InsertOptions{
		Documents: []interface{}{run},
	}
	return s.store.Insert(CollectionRuns, opts)
}

func (s ClaimStore) InsertResult(result Result) error {
	opts := storage.InsertOptions{
		Documents: []interface{}{result},
	}
	return s.store.Insert(CollectionResults, opts)
}

func (s ClaimStore) InsertOutput(output Output) error {
	opts := storage.InsertOptions{
		Documents: []interface{}{output},
	}
	return s.store.Insert(CollectionOutputs, opts)
}

func (s ClaimStore) UpdateInstallation(installation Installation) error {
	installation.SchemaVersion = SchemaVersion
	opts := storage.UpdateOptions{
		Document: installation,
	}
	return s.store.Update(CollectionInstallations, opts)
}

func (s ClaimStore) UpsertRun(run Run) error {
	opts := storage.UpdateOptions{
		Upsert:   true,
		Document: run,
	}
	return s.store.Update(CollectionRuns, opts)
}

func (s ClaimStore) UpsertInstallation(installation Installation) error {
	installation.SchemaVersion = SchemaVersion
	opts := storage.UpdateOptions{
		Upsert:   true,
		Document: installation,
	}
	return s.store.Update(CollectionInstallations, opts)
}

// RemoveInstallation and all associated data.
func (s ClaimStore) RemoveInstallation(namespace string, name string) error {
	removeInstallation := storage.RemoveOptions{
		Filter: bson.M{
			"namespace": namespace,
			"name":      name,
		},
	}
	err := s.store.Remove(CollectionInstallations, removeInstallation)
	if err != nil {
		return err
	}

	// Find associated documents
	removeChildDocs := storage.RemoveOptions{
		Filter: bson.M{
			"namespace":    namespace,
			"installation": name,
		},
		All: true,
	}

	// Delete runs
	err = s.store.Remove(CollectionRuns, removeChildDocs)
	if err != nil {
		return err
	}

	// Delete results
	err = s.store.Remove(CollectionResults, removeChildDocs)
	if err != nil {
		return err
	}

	// Delete outputs
	err = s.store.Remove(CollectionOutputs, removeChildDocs)
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
