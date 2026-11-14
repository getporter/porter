package mongodb

import (
	"context"
	"encoding/json"
	"fmt"

	"get.porter.sh/porter/pkg/storage/plugins"
	"get.porter.sh/porter/pkg/storage/plugins/protov4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ plugins.StorageBackend = &Backend{}

// Collection names. These match the names used by the v3 protocol so that
// existing databases remain readable with no data migration.
const (
	collInstallations = "installations"
	collRuns          = "runs"
	collResults       = "results"
	collOutputs       = "outputs"
	collCredentials   = "credentials"
	collParameters    = "parameters"
	collWorkflows     = "workflows"
	collConfig        = "config"
)

const logsOutputName = "io.cnab.outputs.invocationImageLogs"

// Backend implements the v4 domain-level storage protocol on top of the
// connection managed by Store. Documents are persisted in the same shape as
// the v3 protocol: the entity's JSON representation, stored as a bson document.
type Backend struct {
	*Store
}

// NewBackend creates a v4 backend that uses the specified mongodb store.
func NewBackend(s *Store) *Backend {
	return &Backend{Store: s}
}

func (b *Backend) Info(ctx context.Context, req *protov4.InfoRequest) (*protov4.InfoResponse, error) {
	return &protov4.InfoResponse{
		ProtocolVersion: plugins.PluginProtocolVersionV4,
		Backend:         "mongodb",
	}, nil
}

// Migrate ensures that the collections have the indexes that Porter queries
// rely on.
func (b *Backend) Migrate(ctx context.Context, req *protov4.MigrateRequest) (*protov4.MigrateResponse, error) {
	err := b.EnsureIndex(ctx, plugins.EnsureIndexOptions{Indices: []plugins.Index{
		// query installations by a namespace (list) or namespace + name (get)
		{Collection: collInstallations, Keys: bson.D{{Key: "namespace", Value: 1}, {Key: "name", Value: 1}}, Unique: true},
		// query runs by installation (list)
		{Collection: collRuns, Keys: bson.D{{Key: "namespace", Value: 1}, {Key: "installation", Value: 1}}},
		// query results by installation (delete or batch get)
		{Collection: collResults, Keys: bson.D{{Key: "namespace", Value: 1}, {Key: "installation", Value: 1}}},
		// query results by run (list)
		{Collection: collResults, Keys: bson.D{{Key: "runId", Value: 1}}},
		// query most recent outputs by run
		{Collection: collOutputs, Keys: bson.D{{Key: "namespace", Value: 1}, {Key: "installation", Value: 1}, {Key: "resultId", Value: -1}}},
		// query outputs by result (list)
		{Collection: collOutputs, Keys: bson.D{{Key: "resultId", Value: 1}, {Key: "name", Value: 1}}, Unique: true},
		// query most recent outputs by name for an installation
		{Collection: collOutputs, Keys: bson.D{{Key: "namespace", Value: 1}, {Key: "installation", Value: 1}, {Key: "name", Value: 1}, {Key: "resultId", Value: -1}}},
		// query credentials by namespace + name
		{Collection: collCredentials, Keys: bson.D{{Key: "namespace", Value: 1}, {Key: "name", Value: 1}}, Unique: true},
		// query parameters by namespace + name
		{Collection: collParameters, Keys: bson.D{{Key: "namespace", Value: 1}, {Key: "name", Value: 1}}, Unique: true},
		// query workflows by namespace (list) or namespace + id (get)
		{Collection: collWorkflows, Keys: bson.D{{Key: "namespace", Value: 1}, {Key: "_id", Value: 1}}, Unique: true},
	}})
	if err != nil {
		return nil, err
	}
	return &protov4.MigrateResponse{}, nil
}

func (b *Backend) Put(ctx context.Context, req *protov4.PutRequest) (*protov4.PutResponse, error) {
	if err := b.Connect(ctx); err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, b.timeout)
	defer cancel()

	switch req.Mode {
	case protov4.PutRequest_INSERT:
		// Group by collection, preserving order, so that each collection gets a single InsertMany.
		grouped := map[string][]interface{}{}
		for _, d := range req.Documents {
			coll, err := collectionFor(d.Kind)
			if err != nil {
				return nil, err
			}
			doc, err := jsonToDocument(d.Json)
			if err != nil {
				return nil, err
			}
			grouped[coll] = append(grouped[coll], doc)
		}
		for coll, docs := range grouped {
			if _, err := b.getCollection(coll).InsertMany(ctx, docs); err != nil {
				return nil, mapError(err)
			}
		}
	case protov4.PutRequest_UPDATE, protov4.PutRequest_UPSERT:
		upsert := req.Mode == protov4.PutRequest_UPSERT
		for _, d := range req.Documents {
			coll, err := collectionFor(d.Kind)
			if err != nil {
				return nil, err
			}
			doc, err := jsonToDocument(d.Json)
			if err != nil {
				return nil, err
			}
			res, err := b.getCollection(coll).ReplaceOne(ctx, keyFilter(d), doc, options.Replace().SetUpsert(upsert))
			if err != nil {
				return nil, mapError(err)
			}
			if !upsert && res.MatchedCount == 0 {
				return nil, status.Errorf(codes.NotFound, "%s not found", coll)
			}
		}
	default:
		return nil, status.Errorf(codes.InvalidArgument, "unsupported put mode %s", req.Mode)
	}
	return &protov4.PutResponse{}, nil
}

func (b *Backend) Get(ctx context.Context, req *protov4.GetRequest) (*protov4.GetResponse, error) {
	coll, err := collectionFor(req.Kind)
	if err != nil {
		return nil, err
	}

	q := query{collection: coll, limit: 1}
	if req.FallbackToGlobal {
		// Prefer the document in the requested namespace over the global one.
		q.filter = bson.M{"name": req.Name, "$or": []bson.M{{"namespace": ""}, {"namespace": req.Namespace}}}
		q.sort = bson.D{{Key: "namespace", Value: -1}}
	} else {
		q.filter = getFilter(req)
	}
	return b.findOne(ctx, req.Kind, q)
}

func (b *Backend) List(ctx context.Context, req *protov4.ListRequest) (*protov4.ListResponse, error) {
	coll, err := collectionFor(req.Kind)
	if err != nil {
		return nil, err
	}

	filter := bson.M{}
	if req.Namespace != "*" {
		filter["namespace"] = req.Namespace
	}
	if req.NameContains != "" {
		// Same behavior as the v3 protocol: the value is used as a regular expression.
		filter["name"] = bson.M{"$regex": req.NameContains}
	}
	for k, v := range req.Labels {
		filter["labels."+k] = v
	}

	sort := bson.D{{Key: "namespace", Value: 1}, {Key: "name", Value: 1}}
	switch req.Kind {
	case protov4.Kind_KIND_WORKFLOW:
		sort = bson.D{{Key: "namespace", Value: 1}, {Key: "_id", Value: 1}}
	case protov4.Kind_KIND_RUN, protov4.Kind_KIND_RESULT:
		sort = bson.D{{Key: "_id", Value: 1}}
	}
	return b.find(ctx, req.Kind, query{collection: coll, filter: filter, sort: sort, skip: req.Skip, limit: req.Limit})
}

func (b *Backend) Remove(ctx context.Context, req *protov4.RemoveRequest) (*protov4.RemoveResponse, error) {
	coll, err := collectionFor(req.Kind)
	if err != nil {
		return nil, err
	}
	if err := b.Connect(ctx); err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, b.timeout)
	defer cancel()

	// Removing a document that doesn't exist is not an error, same as v3.
	_, err = b.getCollection(coll).DeleteOne(ctx, getFilter(&protov4.GetRequest{
		Kind: req.Kind, Id: req.Id, Namespace: req.Namespace, Name: req.Name,
	}))
	if err != nil {
		return nil, mapError(err)
	}
	return &protov4.RemoveResponse{}, nil
}

// RemoveInstallation removes the installation and its runs, results and
// outputs. MongoDB has no multi-collection atomicity here, so a failure
// partway through can leave orphaned child documents, as in v3.
func (b *Backend) RemoveInstallation(ctx context.Context, req *protov4.RemoveInstallationRequest) (*protov4.RemoveInstallationResponse, error) {
	if err := b.Connect(ctx); err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, b.timeout)
	defer cancel()

	ref := req.Installation
	_, err := b.getCollection(collInstallations).DeleteOne(ctx, bson.M{"namespace": ref.Namespace, "name": ref.Installation})
	if err != nil {
		return nil, mapError(err)
	}

	children := bson.M{"namespace": ref.Namespace, "installation": ref.Installation}
	for _, coll := range []string{collRuns, collResults, collOutputs} {
		if _, err := b.getCollection(coll).DeleteMany(ctx, children); err != nil {
			return nil, mapError(err)
		}
	}
	return &protov4.RemoveInstallationResponse{}, nil
}

func (b *Backend) ListRuns(ctx context.Context, req *protov4.ListRunsRequest) (*protov4.ListRunsResponse, error) {
	res, err := b.find(ctx, protov4.Kind_KIND_RUN, query{
		collection: collRuns,
		filter:     installationFilter(req.Installation),
		sort:       bson.D{{Key: "_id", Value: 1}},
	})
	if err != nil {
		return nil, err
	}
	return &protov4.ListRunsResponse{Runs: res.Documents}, nil
}

func (b *Backend) GetLastRun(ctx context.Context, req *protov4.InstallationRef) (*protov4.GetResponse, error) {
	return b.findOne(ctx, protov4.Kind_KIND_RUN, query{
		collection: collRuns,
		filter:     installationFilter(req),
		sort:       bson.D{{Key: "_id", Value: -1}},
		limit:      1,
	})
}

// GetActiveRuns returns the runs that have not reached a terminal status,
// including runs that have no results yet.
func (b *Backend) GetActiveRuns(ctx context.Context, req *protov4.InstallationRef) (*protov4.ListResponse, error) {
	return b.aggregate(ctx, protov4.Kind_KIND_RUN, collRuns, []bson.D{
		{{Key: "$match", Value: installationFilter(req)}},
		{{Key: "$lookup", Value: bson.M{
			"from":         collResults,
			"localField":   "_id",
			"foreignField": "runId",
			"as":           "results",
		}}},
		// A run is active if none of its results (including zero results)
		// has reached a terminal status.
		{{Key: "$match", Value: bson.M{
			"results.status": bson.M{"$nin": []string{"succeeded", "failed", "canceled"}},
		}}},
		{{Key: "$project", Value: bson.M{"results": 0}}},
	}, nil)
}

func (b *Backend) ListResults(ctx context.Context, req *protov4.ListResultsRequest) (*protov4.ListResponse, error) {
	var filter bson.M
	switch scope := req.Scope.(type) {
	case *protov4.ListResultsRequest_RunId:
		filter = bson.M{"runId": scope.RunId}
	case *protov4.ListResultsRequest_Installation:
		filter = installationFilter(scope.Installation)
	default:
		return nil, status.Error(codes.InvalidArgument, "ListResults requires a run_id or installation")
	}
	return b.find(ctx, protov4.Kind_KIND_RESULT, query{
		collection: collResults,
		filter:     filter,
		sort:       bson.D{{Key: "_id", Value: 1}},
	})
}

func (b *Backend) InsertOutputs(ctx context.Context, req *protov4.InsertOutputsRequest) (*protov4.InsertOutputsResponse, error) {
	docs := make([]interface{}, len(req.Outputs))
	for i, o := range req.Outputs {
		doc, err := jsonToDocument(o.Json)
		if err != nil {
			return nil, err
		}
		docs[i] = doc
	}
	if len(docs) == 0 {
		return &protov4.InsertOutputsResponse{}, nil
	}
	if err := b.Connect(ctx); err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, b.timeout)
	defer cancel()
	if _, err := b.getCollection(collOutputs).InsertMany(ctx, docs); err != nil {
		return nil, mapError(err)
	}
	return &protov4.InsertOutputsResponse{}, nil
}

func (b *Backend) ListOutputs(ctx context.Context, req *protov4.ListOutputsRequest) (*protov4.ListResponse, error) {
	var filter bson.M
	switch scope := req.Scope.(type) {
	case *protov4.ListOutputsRequest_RunId:
		filter = bson.M{"runId": scope.RunId}
	case *protov4.ListOutputsRequest_ResultId:
		filter = bson.M{"resultId": scope.ResultId}
	default:
		return nil, status.Error(codes.InvalidArgument, "ListOutputs requires a run_id or result_id")
	}
	return b.find(ctx, protov4.Kind_KIND_OUTPUT, query{
		collection: collOutputs,
		filter:     filter,
		sort:       bson.D{{Key: "name", Value: 1}},
	})
}

func (b *Backend) GetOutput(ctx context.Context, req *protov4.GetOutputRequest) (*protov4.GetResponse, error) {
	switch scope := req.Scope.(type) {
	case *protov4.GetOutputRequest_RunId:
		return b.findOne(ctx, protov4.Kind_KIND_OUTPUT, query{
			collection: collOutputs,
			filter:     bson.M{"runId": scope.RunId, "name": req.Name},
			limit:      1,
		})
	case *protov4.GetOutputRequest_Installation:
		filter := installationFilter(scope.Installation)
		filter["name"] = req.Name
		return b.findOne(ctx, protov4.Kind_KIND_OUTPUT, query{
			collection: collOutputs,
			filter:     filter,
			sort:       bson.D{{Key: "resultId", Value: -1}},
			limit:      1,
		})
	default:
		return nil, status.Error(codes.InvalidArgument, "GetOutput requires a run_id or installation")
	}
}

// GetLastOutputs returns the newest output for each output name.
func (b *Backend) GetLastOutputs(ctx context.Context, req *protov4.InstallationRef) (*protov4.ListResponse, error) {
	return b.aggregate(ctx, protov4.Kind_KIND_OUTPUT, collOutputs, []bson.D{
		{{Key: "$match", Value: installationFilter(req)}},
		// Newest on top, within each output name
		{{Key: "$sort", Value: bson.D{
			{Key: "namespace", Value: 1},
			{Key: "installation", Value: 1},
			{Key: "name", Value: 1},
			{Key: "resultId", Value: -1},
		}}},
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$name"},
			{Key: "lastOutput", Value: bson.M{"$first": "$$ROOT"}},
		}}},
	}, func(raw bson.Raw) (bson.Raw, error) {
		last, ok := raw.Lookup("lastOutput").DocumentOK()
		if !ok {
			return nil, status.Error(codes.Internal, "aggregated output is missing lastOutput")
		}
		return last, nil
	})
}

func (b *Backend) FindInstallationsReferencing(ctx context.Context, req *protov4.FindReferencingRequest) (*protov4.ListResponse, error) {
	var filter bson.M
	switch target := req.Target.(type) {
	case *protov4.FindReferencingRequest_CredentialSet:
		filter = bson.M{"credentialSets": target.CredentialSet}
	case *protov4.FindReferencingRequest_ParameterSet:
		filter = bson.M{"parameterSets": target.ParameterSet}
	case *protov4.FindReferencingRequest_Installation:
		filter = bson.M{"status.references.installation": target.Installation}
	default:
		return nil, status.Error(codes.InvalidArgument, "FindInstallationsReferencing requires a target")
	}
	return b.find(ctx, protov4.Kind_KIND_INSTALLATION, query{collection: collInstallations, filter: filter})
}

func (b *Backend) FindDependencyCandidates(ctx context.Context, req *protov4.FindDependencyCandidatesRequest) (*protov4.ListResponse, error) {
	filter := bson.M{
		"uninstalled":       bson.M{"$ne": true},
		"bundle.repository": req.Repository,
		"$or":               []bson.M{{"namespace": ""}, {"namespace": req.Namespace}},
	}
	if req.Digest != "" {
		filter["status.bundleDigest"] = req.Digest
	}
	return b.find(ctx, protov4.Kind_KIND_INSTALLATION, query{
		collection: collInstallations,
		filter:     filter,
		// prefer the local namespace match over the global ("") one
		sort: bson.D{{Key: "namespace", Value: -1}},
	})
}

type query struct {
	collection string
	filter     bson.M
	sort       bson.D
	skip       int64
	limit      int64
}

func (b *Backend) find(ctx context.Context, kind protov4.Kind, q query) (*protov4.ListResponse, error) {
	if err := b.Connect(ctx); err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, b.timeout)
	defer cancel()

	opts := options.Find()
	if q.sort != nil {
		opts.SetSort(q.sort)
	}
	if q.skip > 0 {
		opts.SetSkip(q.skip)
	}
	if q.limit > 0 {
		opts.SetLimit(q.limit)
	}
	if q.filter == nil {
		q.filter = bson.M{}
	}

	cur, err := b.getCollection(q.collection).Find(ctx, q.filter, opts)
	if err != nil {
		return nil, mapError(err)
	}
	defer cur.Close(ctx)

	resp := &protov4.ListResponse{}
	for cur.Next(ctx) {
		doc, err := rawToDocument(kind, cur.Current)
		if err != nil {
			return nil, err
		}
		resp.Documents = append(resp.Documents, doc)
	}
	return resp, mapError(cur.Err())
}

func (b *Backend) findOne(ctx context.Context, kind protov4.Kind, q query) (*protov4.GetResponse, error) {
	q.limit = 1
	res, err := b.find(ctx, kind, q)
	if err != nil {
		return nil, err
	}
	if len(res.Documents) == 0 {
		return nil, status.Errorf(codes.NotFound, "%s not found", q.collection)
	}
	return &protov4.GetResponse{Document: res.Documents[0]}, nil
}

// aggregate runs a pipeline. The optional unwrap function selects the document
// to return from each pipeline result.
func (b *Backend) aggregate(ctx context.Context, kind protov4.Kind, collection string, pipeline []bson.D, unwrap func(bson.Raw) (bson.Raw, error)) (*protov4.ListResponse, error) {
	if err := b.Connect(ctx); err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, b.timeout)
	defer cancel()

	cur, err := b.getCollection(collection).Aggregate(ctx, pipeline)
	if err != nil {
		return nil, mapError(err)
	}
	defer cur.Close(ctx)

	resp := &protov4.ListResponse{}
	for cur.Next(ctx) {
		raw := cur.Current
		if unwrap != nil {
			if raw, err = unwrap(raw); err != nil {
				return nil, err
			}
		}
		doc, err := rawToDocument(kind, raw)
		if err != nil {
			return nil, err
		}
		resp.Documents = append(resp.Documents, doc)
	}
	return resp, mapError(cur.Err())
}

func collectionFor(kind protov4.Kind) (string, error) {
	switch kind {
	case protov4.Kind_KIND_INSTALLATION:
		return collInstallations, nil
	case protov4.Kind_KIND_RUN:
		return collRuns, nil
	case protov4.Kind_KIND_RESULT:
		return collResults, nil
	case protov4.Kind_KIND_OUTPUT:
		return collOutputs, nil
	case protov4.Kind_KIND_CREDENTIAL_SET:
		return collCredentials, nil
	case protov4.Kind_KIND_PARAMETER_SET:
		return collParameters, nil
	case protov4.Kind_KIND_WORKFLOW:
		return collWorkflows, nil
	case protov4.Kind_KIND_SCHEMA:
		return collConfig, nil
	default:
		return "", status.Errorf(codes.InvalidArgument, "unsupported document kind %s", kind)
	}
}

// keyFilter identifies the stored document that a document replaces.
func keyFilter(d *protov4.Document) bson.M {
	switch d.Kind {
	case protov4.Kind_KIND_RUN, protov4.Kind_KIND_RESULT, protov4.Kind_KIND_SCHEMA:
		return bson.M{"_id": d.Id}
	case protov4.Kind_KIND_WORKFLOW:
		return bson.M{"namespace": d.Namespace, "_id": d.Id}
	case protov4.Kind_KIND_OUTPUT:
		return bson.M{"resultId": d.ResultId, "name": d.Name}
	default: // installations, credential sets and parameter sets
		return bson.M{"namespace": d.Namespace, "name": d.Name}
	}
}

func getFilter(req *protov4.GetRequest) bson.M {
	switch {
	case req.Kind == protov4.Kind_KIND_WORKFLOW:
		return bson.M{"namespace": req.Namespace, "_id": req.Id}
	case req.Id != "":
		return bson.M{"_id": req.Id}
	default:
		return bson.M{"namespace": req.Namespace, "name": req.Name}
	}
}

func installationFilter(ref *protov4.InstallationRef) bson.M {
	return bson.M{"namespace": ref.GetNamespace(), "installation": ref.GetInstallation()}
}

// jsonToDocument converts the json representation of a Porter document into
// something that can be stored. Going through json ensures any custom
// marshaling logic (e.g. calculated _id fields) is applied, same as v3.
func jsonToDocument(data []byte) (map[string]interface{}, error) {
	var doc map[string]interface{}
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "document is not valid json: %v", err)
	}
	return doc, nil
}

// rawToDocument converts a stored bson document into a protocol document with
// its key fields populated and its json representation set.
func rawToDocument(kind protov4.Kind, raw bson.Raw) (*protov4.Document, error) {
	var m bson.M
	if err := bson.Unmarshal(raw, &m); err != nil {
		return nil, status.Errorf(codes.Internal, "could not unmarshal stored document: %v", err)
	}
	data, err := json.Marshal(m)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not convert stored document to json: %v", err)
	}

	str := func(key string) string {
		if v, ok := m[key].(string); ok {
			return v
		}
		return ""
	}
	return &protov4.Document{
		Kind:         kind,
		Id:           str("_id"),
		Namespace:    str("namespace"),
		Name:         str("name"),
		Installation: str("installation"),
		RunId:        str("runId"),
		ResultId:     str("resultId"),
		Json:         data,
	}, nil
}

// mapError converts mongodb errors to gRPC status errors.
func mapError(err error) error {
	switch {
	case err == nil:
		return nil
	case mongo.IsDuplicateKeyError(err):
		return status.Error(codes.AlreadyExists, err.Error())
	case err == mongo.ErrNoDocuments:
		return status.Error(codes.NotFound, err.Error())
	default:
		if _, ok := status.FromError(err); ok {
			return err
		}
		return fmt.Errorf("mongodb: %w", err)
	}
}
