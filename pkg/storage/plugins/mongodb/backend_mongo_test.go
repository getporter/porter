package mongodb_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/pkg/storage/plugins"
	"get.porter.sh/porter/pkg/storage/plugins/protov4"
	"get.porter.sh/porter/pkg/storage/plugins/testplugin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// These tests exercise the mongodb Backend against a real MongoDB, which the
// test plugin starts in a container when one isn't already running.

func newTestBackend(t *testing.T) plugins.StorageBackend {
	tc := portercontext.NewTestContext(t)
	backend := testplugin.NewTestStoragePlugin(tc)
	t.Cleanup(func() { backend.Close() })

	_, err := backend.Migrate(context.Background(), &protov4.MigrateRequest{})
	require.NoError(t, err, "Migrate failed")
	return backend
}

// doc builds a protocol document from a set of fields. The id, namespace, name,
// installation, runId and resultId fields are also used as the document keys.
func doc(kind protov4.Kind, fields map[string]interface{}) *protov4.Document {
	str := func(k string) string {
		v, _ := fields[k].(string)
		return v
	}
	data, err := json.Marshal(fields)
	if err != nil {
		panic(err)
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
	}
}

func put(t *testing.T, b plugins.StorageBackend, mode protov4.PutRequest_Mode, docs ...*protov4.Document) error {
	_, err := b.Put(context.Background(), &protov4.PutRequest{Mode: mode, Documents: docs})
	return err
}

func mustPut(t *testing.T, b plugins.StorageBackend, mode protov4.PutRequest_Mode, docs ...*protov4.Document) {
	require.NoError(t, put(t, b, mode, docs...))
}

func names(docs []*protov4.Document) []string {
	out := make([]string, len(docs))
	for i, d := range docs {
		out[i] = d.Name
	}
	return out
}

func ids(docs []*protov4.Document) []string {
	out := make([]string, len(docs))
	for i, d := range docs {
		out[i] = d.Id
	}
	return out
}

func inst(namespace, name string, extra map[string]interface{}) *protov4.Document {
	fields := map[string]interface{}{"namespace": namespace, "name": name}
	for k, v := range extra {
		fields[k] = v
	}
	return doc(protov4.Kind_KIND_INSTALLATION, fields)
}

func TestBackend_Migrate_IsIdempotent(t *testing.T) {
	b := newTestBackend(t)
	_, err := b.Migrate(context.Background(), &protov4.MigrateRequest{})
	require.NoError(t, err)
}

func TestBackend_Put(t *testing.T) {
	ctx := context.Background()
	b := newTestBackend(t)

	t.Run("insert", func(t *testing.T) {
		mustPut(t, b, protov4.PutRequest_INSERT, inst("dev", "a", map[string]interface{}{"v": 1}))
		got, err := b.Get(ctx, &protov4.GetRequest{Kind: protov4.Kind_KIND_INSTALLATION, Namespace: "dev", Name: "a"})
		require.NoError(t, err)
		assert.Equal(t, "a", got.Document.Name)
		assert.Equal(t, "dev", got.Document.Namespace)
	})

	t.Run("insert duplicate", func(t *testing.T) {
		err := put(t, b, protov4.PutRequest_INSERT, inst("dev", "a", nil))
		assert.Equal(t, codes.AlreadyExists, status.Code(err))
	})

	t.Run("update replaces the document", func(t *testing.T) {
		mustPut(t, b, protov4.PutRequest_UPDATE, inst("dev", "a", map[string]interface{}{"v": 2}))
		got, err := b.Get(ctx, &protov4.GetRequest{Kind: protov4.Kind_KIND_INSTALLATION, Namespace: "dev", Name: "a"})
		require.NoError(t, err)
		var fields map[string]interface{}
		require.NoError(t, json.Unmarshal(got.Document.Json, &fields))
		assert.EqualValues(t, 2, fields["v"])
	})

	t.Run("update missing", func(t *testing.T) {
		err := put(t, b, protov4.PutRequest_UPDATE, inst("dev", "missing", nil))
		assert.Equal(t, codes.NotFound, status.Code(err))
	})

	t.Run("upsert creates and replaces", func(t *testing.T) {
		mustPut(t, b, protov4.PutRequest_UPSERT, inst("dev", "b", nil))
		mustPut(t, b, protov4.PutRequest_UPSERT, inst("dev", "b", map[string]interface{}{"v": 3}))
		got, err := b.List(ctx, &protov4.ListRequest{Kind: protov4.Kind_KIND_INSTALLATION, Namespace: "dev"})
		require.NoError(t, err)
		assert.Equal(t, []string{"a", "b"}, names(got.Documents), "the upsert should not create a duplicate")
	})
}

func TestBackend_Get(t *testing.T) {
	ctx := context.Background()
	b := newTestBackend(t)
	mustPut(t, b, protov4.PutRequest_INSERT,
		doc(protov4.Kind_KIND_CREDENTIAL_SET, map[string]interface{}{"namespace": "", "name": "creds", "from": "global"}),
		doc(protov4.Kind_KIND_CREDENTIAL_SET, map[string]interface{}{"namespace": "dev", "name": "creds", "from": "dev"}),
		doc(protov4.Kind_KIND_CREDENTIAL_SET, map[string]interface{}{"namespace": "", "name": "only-global", "from": "global"}),
	)

	from := func(t *testing.T, req *protov4.GetRequest) string {
		got, err := b.Get(ctx, req)
		require.NoError(t, err)
		var fields map[string]interface{}
		require.NoError(t, json.Unmarshal(got.Document.Json, &fields))
		return fields["from"].(string)
	}

	t.Run("fallback prefers the namespace", func(t *testing.T) {
		assert.Equal(t, "dev", from(t, &protov4.GetRequest{Kind: protov4.Kind_KIND_CREDENTIAL_SET, Namespace: "dev", Name: "creds", FallbackToGlobal: true}))
	})

	t.Run("fallback to global", func(t *testing.T) {
		assert.Equal(t, "global", from(t, &protov4.GetRequest{Kind: protov4.Kind_KIND_CREDENTIAL_SET, Namespace: "dev", Name: "only-global", FallbackToGlobal: true}))
		assert.Equal(t, "global", from(t, &protov4.GetRequest{Kind: protov4.Kind_KIND_CREDENTIAL_SET, Namespace: "other", Name: "creds", FallbackToGlobal: true}))
	})

	t.Run("no fallback", func(t *testing.T) {
		_, err := b.Get(ctx, &protov4.GetRequest{Kind: protov4.Kind_KIND_CREDENTIAL_SET, Namespace: "dev", Name: "only-global"})
		assert.Equal(t, codes.NotFound, status.Code(err))
	})
}

func TestBackend_List(t *testing.T) {
	ctx := context.Background()
	b := newTestBackend(t)
	mustPut(t, b, protov4.PutRequest_INSERT,
		inst("prod", "zeta", nil),
		inst("dev", "beta", map[string]interface{}{"labels": map[string]string{"team": "red", "owner": "marie"}}),
		inst("dev", "alpha", map[string]interface{}{"labels": map[string]string{"team": "red"}}),
		inst("dev", "a.c", nil),
		inst("dev", "abc", nil),
	)

	list := func(t *testing.T, req *protov4.ListRequest) []string {
		req.Kind = protov4.Kind_KIND_INSTALLATION
		got, err := b.List(ctx, req)
		require.NoError(t, err)
		return names(got.Documents)
	}

	t.Run("sorted by namespace and name", func(t *testing.T) {
		assert.Equal(t, []string{"a.c", "abc", "alpha", "beta", "zeta"}, list(t, &protov4.ListRequest{Namespace: "*"}))
	})

	t.Run("namespace", func(t *testing.T) {
		assert.Equal(t, []string{"zeta"}, list(t, &protov4.ListRequest{Namespace: "prod"}))
	})

	t.Run("skip and limit", func(t *testing.T) {
		assert.Equal(t, []string{"abc", "alpha"}, list(t, &protov4.ListRequest{Namespace: "dev", Skip: 1, Limit: 2}))
	})

	t.Run("labels must all match", func(t *testing.T) {
		assert.Equal(t, []string{"alpha", "beta"}, list(t, &protov4.ListRequest{Namespace: "dev", Labels: map[string]string{"team": "red"}}))
		assert.Equal(t, []string{"beta"}, list(t, &protov4.ListRequest{Namespace: "dev", Labels: map[string]string{"team": "red", "owner": "marie"}}))
	})

	t.Run("name contains is a literal substring", func(t *testing.T) {
		assert.Equal(t, []string{"alpha"}, list(t, &protov4.ListRequest{Namespace: "dev", NameContains: "lph"}))
		assert.Equal(t, []string{"a.c"}, list(t, &protov4.ListRequest{Namespace: "dev", NameContains: "a.c"}), "a dot is not a wildcard")
		assert.Empty(t, list(t, &protov4.ListRequest{Namespace: "dev", NameContains: "^a"}), "a regular expression is not interpreted")
	})
}

func TestBackend_Remove(t *testing.T) {
	ctx := context.Background()
	b := newTestBackend(t)
	mustPut(t, b, protov4.PutRequest_INSERT, inst("dev", "a", nil))

	_, err := b.Remove(ctx, &protov4.RemoveRequest{Kind: protov4.Kind_KIND_INSTALLATION, Namespace: "dev", Name: "a"})
	require.NoError(t, err)
	_, err = b.Get(ctx, &protov4.GetRequest{Kind: protov4.Kind_KIND_INSTALLATION, Namespace: "dev", Name: "a"})
	assert.Equal(t, codes.NotFound, status.Code(err))

	_, err = b.Remove(ctx, &protov4.RemoveRequest{Kind: protov4.Kind_KIND_INSTALLATION, Namespace: "dev", Name: "a"})
	assert.NoError(t, err, "removing a document that doesn't exist is not an error")
}

func run(id, namespace, installation string) *protov4.Document {
	return doc(protov4.Kind_KIND_RUN, map[string]interface{}{"_id": id, "namespace": namespace, "installation": installation})
}

func result(id, runID, namespace, installation, status string) *protov4.Document {
	return doc(protov4.Kind_KIND_RESULT, map[string]interface{}{
		"_id": id, "runId": runID, "namespace": namespace, "installation": installation, "status": status,
	})
}

func output(name, runID, resultID, namespace, installation, value string) *protov4.Document {
	return doc(protov4.Kind_KIND_OUTPUT, map[string]interface{}{
		"name": name, "runId": runID, "resultId": resultID, "namespace": namespace, "installation": installation, "value": value,
	})
}

func TestBackend_Runs(t *testing.T) {
	ctx := context.Background()
	b := newTestBackend(t)
	ref := &protov4.InstallationRef{Namespace: "dev", Installation: "mybuns"}

	// run1 is complete, run2 is running, run3 hasn't started, other belongs to another installation
	mustPut(t, b, protov4.PutRequest_INSERT,
		run("01", "dev", "mybuns"), run("02", "dev", "mybuns"), run("03", "dev", "mybuns"), run("04", "dev", "other"))
	mustPut(t, b, protov4.PutRequest_INSERT,
		result("r01a", "01", "dev", "mybuns", "running"),
		result("r01b", "01", "dev", "mybuns", "succeeded"),
		result("r02a", "02", "dev", "mybuns", "running"),
		result("r04a", "04", "dev", "other", "running"))

	t.Run("list runs", func(t *testing.T) {
		got, err := b.ListRuns(ctx, &protov4.ListRunsRequest{Installation: ref})
		require.NoError(t, err)
		assert.Equal(t, []string{"01", "02", "03"}, ids(got.Runs))
	})

	t.Run("last run", func(t *testing.T) {
		got, err := b.GetLastRun(ctx, ref)
		require.NoError(t, err)
		assert.Equal(t, "03", got.Document.Id)
	})

	t.Run("last run when there are none", func(t *testing.T) {
		_, err := b.GetLastRun(ctx, &protov4.InstallationRef{Namespace: "dev", Installation: "nope"})
		assert.Equal(t, codes.NotFound, status.Code(err))
	})

	t.Run("active runs include runs without results", func(t *testing.T) {
		got, err := b.GetActiveRuns(ctx, ref)
		require.NoError(t, err)
		assert.ElementsMatch(t, []string{"02", "03"}, ids(got.Documents))
	})

	t.Run("results by run", func(t *testing.T) {
		got, err := b.ListResults(ctx, &protov4.ListResultsRequest{Scope: &protov4.ListResultsRequest_RunId{RunId: "01"}})
		require.NoError(t, err)
		assert.Equal(t, []string{"r01a", "r01b"}, ids(got.Documents))
	})

	t.Run("results by installation", func(t *testing.T) {
		got, err := b.ListResults(ctx, &protov4.ListResultsRequest{Scope: &protov4.ListResultsRequest_Installation{Installation: ref}})
		require.NoError(t, err)
		assert.Equal(t, []string{"r01a", "r01b", "r02a"}, ids(got.Documents))
	})
}

func TestBackend_Outputs(t *testing.T) {
	ctx := context.Background()
	b := newTestBackend(t)
	ref := &protov4.InstallationRef{Namespace: "dev", Installation: "mybuns"}

	_, err := b.InsertOutputs(ctx, &protov4.InsertOutputsRequest{Outputs: []*protov4.Document{
		output("logs", "01", "r01a", "dev", "mybuns", "first logs"),
		output("logs", "01", "r01b", "dev", "mybuns", "last logs"),
		output("port", "01", "r01b", "dev", "mybuns", "8080"),
		output("name", "02", "r02a", "dev", "mybuns", "bob"),
		output("logs", "09", "r09a", "dev", "other", "other logs"),
	}})
	require.NoError(t, err)

	value := func(t *testing.T, resp *protov4.GetResponse) string {
		var fields map[string]interface{}
		require.NoError(t, json.Unmarshal(resp.Document.Json, &fields))
		return fields["value"].(string)
	}

	t.Run("by run is the newest result", func(t *testing.T) {
		got, err := b.GetOutput(ctx, &protov4.GetOutputRequest{Scope: &protov4.GetOutputRequest_RunId{RunId: "01"}, Name: "logs"})
		require.NoError(t, err)
		assert.Equal(t, "last logs", value(t, got))
	})

	t.Run("by installation is the newest result", func(t *testing.T) {
		got, err := b.GetOutput(ctx, &protov4.GetOutputRequest{Scope: &protov4.GetOutputRequest_Installation{Installation: ref}, Name: "logs"})
		require.NoError(t, err)
		assert.Equal(t, "last logs", value(t, got))
	})

	t.Run("missing", func(t *testing.T) {
		_, err := b.GetOutput(ctx, &protov4.GetOutputRequest{Scope: &protov4.GetOutputRequest_RunId{RunId: "01"}, Name: "nope"})
		assert.Equal(t, codes.NotFound, status.Code(err))
	})

	t.Run("by run, sorted by name", func(t *testing.T) {
		got, err := b.ListOutputs(ctx, &protov4.ListOutputsRequest{Scope: &protov4.ListOutputsRequest_RunId{RunId: "01"}})
		require.NoError(t, err)
		assert.Equal(t, []string{"logs", "logs", "port"}, names(got.Documents))
	})

	t.Run("by result", func(t *testing.T) {
		got, err := b.ListOutputs(ctx, &protov4.ListOutputsRequest{Scope: &protov4.ListOutputsRequest_ResultId{ResultId: "r01b"}})
		require.NoError(t, err)
		assert.Equal(t, []string{"logs", "port"}, names(got.Documents))
	})

	t.Run("duplicate output for a result", func(t *testing.T) {
		_, err := b.InsertOutputs(ctx, &protov4.InsertOutputsRequest{Outputs: []*protov4.Document{
			output("port", "01", "r01b", "dev", "mybuns", "9090"),
		}})
		assert.Equal(t, codes.AlreadyExists, status.Code(err))
	})

	t.Run("last output of each name", func(t *testing.T) {
		got, err := b.GetLastOutputs(ctx, ref)
		require.NoError(t, err)
		values := map[string]string{}
		for _, d := range got.Documents {
			var fields map[string]interface{}
			require.NoError(t, json.Unmarshal(d.Json, &fields))
			values[d.Name] = fmt.Sprint(fields["value"])
		}
		assert.Equal(t, map[string]string{"logs": "last logs", "port": "8080", "name": "bob"}, values)
	})
}

func TestBackend_RemoveInstallation(t *testing.T) {
	ctx := context.Background()
	b := newTestBackend(t)
	ref := &protov4.InstallationRef{Namespace: "dev", Installation: "mybuns"}

	mustPut(t, b, protov4.PutRequest_INSERT, inst("dev", "mybuns", nil), inst("dev", "keep", nil))
	mustPut(t, b, protov4.PutRequest_INSERT, run("01", "dev", "mybuns"), run("02", "dev", "keep"))
	mustPut(t, b, protov4.PutRequest_INSERT, result("r01", "01", "dev", "mybuns", "succeeded"), result("r02", "02", "dev", "keep", "succeeded"))
	_, err := b.InsertOutputs(ctx, &protov4.InsertOutputsRequest{Outputs: []*protov4.Document{
		output("logs", "01", "r01", "dev", "mybuns", "x"),
		output("logs", "02", "r02", "dev", "keep", "y"),
	}})
	require.NoError(t, err)

	_, err = b.RemoveInstallation(ctx, &protov4.RemoveInstallationRequest{Installation: ref})
	require.NoError(t, err)

	_, err = b.Get(ctx, &protov4.GetRequest{Kind: protov4.Kind_KIND_INSTALLATION, Namespace: "dev", Name: "mybuns"})
	assert.Equal(t, codes.NotFound, status.Code(err))
	runs, err := b.ListRuns(ctx, &protov4.ListRunsRequest{Installation: ref})
	require.NoError(t, err)
	assert.Empty(t, runs.Runs)
	results, err := b.ListResults(ctx, &protov4.ListResultsRequest{Scope: &protov4.ListResultsRequest_Installation{Installation: ref}})
	require.NoError(t, err)
	assert.Empty(t, results.Documents)
	outputs, err := b.GetLastOutputs(ctx, ref)
	require.NoError(t, err)
	assert.Empty(t, outputs.Documents)

	// the other installation is untouched
	keepRef := &protov4.InstallationRef{Namespace: "dev", Installation: "keep"}
	runs, err = b.ListRuns(ctx, &protov4.ListRunsRequest{Installation: keepRef})
	require.NoError(t, err)
	assert.Equal(t, []string{"02"}, ids(runs.Runs))
	outputs, err = b.GetLastOutputs(ctx, keepRef)
	require.NoError(t, err)
	assert.Len(t, outputs.Documents, 1)
}

func TestBackend_FindInstallationsReferencing(t *testing.T) {
	ctx := context.Background()
	b := newTestBackend(t)
	mustPut(t, b, protov4.PutRequest_INSERT,
		inst("dev", "a", map[string]interface{}{"credentialSets": []string{"creds"}, "parameterSets": []string{"params"}}),
		inst("prod", "b", map[string]interface{}{"credentialSets": []string{"creds", "more"}}),
		inst("dev", "c", map[string]interface{}{
			"status": map[string]interface{}{"references": []map[string]string{{"installation": "dev/a", "dependency": "db"}}},
		}),
	)

	find := func(t *testing.T, req *protov4.FindReferencingRequest) []string {
		got, err := b.FindInstallationsReferencing(ctx, req)
		require.NoError(t, err)
		return names(got.Documents)
	}

	assert.ElementsMatch(t, []string{"a", "b"}, find(t, &protov4.FindReferencingRequest{
		Target: &protov4.FindReferencingRequest_CredentialSet{CredentialSet: "creds"}}))
	assert.Equal(t, []string{"a"}, find(t, &protov4.FindReferencingRequest{
		Target: &protov4.FindReferencingRequest_CredentialSet{CredentialSet: "creds"}, Namespace: "dev"}), "limited to a namespace")
	assert.Equal(t, []string{"a"}, find(t, &protov4.FindReferencingRequest{
		Target: &protov4.FindReferencingRequest_ParameterSet{ParameterSet: "params"}}))
	assert.Equal(t, []string{"c"}, find(t, &protov4.FindReferencingRequest{
		Target: &protov4.FindReferencingRequest_Installation{Installation: "dev/a"}}))
	assert.Empty(t, find(t, &protov4.FindReferencingRequest{
		Target: &protov4.FindReferencingRequest_CredentialSet{CredentialSet: "nope"}}))

	_, err := b.FindInstallationsReferencing(ctx, &protov4.FindReferencingRequest{})
	assert.Equal(t, codes.InvalidArgument, status.Code(err), "a target is required")
}

func TestBackend_FindDependencyCandidates(t *testing.T) {
	ctx := context.Background()
	b := newTestBackend(t)
	bundle := func(repo string) map[string]interface{} {
		return map[string]interface{}{"bundle": map[string]interface{}{"repository": repo}}
	}
	withDigest := func(repo, digest string) map[string]interface{} {
		m := bundle(repo)
		m["status"] = map[string]interface{}{"bundleDigest": digest}
		return m
	}
	uninstalled := bundle("example.com/db")
	uninstalled["uninstalled"] = true

	mustPut(t, b, protov4.PutRequest_INSERT,
		inst("", "global-db", withDigest("example.com/db", "sha256:aaa")),
		inst("dev", "dev-db", withDigest("example.com/db", "sha256:aaa")),
		inst("dev", "dev-db-old", withDigest("example.com/db", "sha256:bbb")),
		inst("prod", "prod-db", withDigest("example.com/db", "sha256:aaa")),
		inst("dev", "other-bundle", withDigest("example.com/other", "sha256:aaa")),
		inst("dev", "removed-db", uninstalled),
	)

	got, err := b.FindDependencyCandidates(ctx, &protov4.FindDependencyCandidatesRequest{
		Namespace: "dev", Repository: "example.com/db", Digest: "sha256:aaa"})
	require.NoError(t, err)
	assert.Equal(t, []string{"dev-db", "global-db"}, names(got.Documents), "the local namespace sorts before the global one")

	got, err = b.FindDependencyCandidates(ctx, &protov4.FindDependencyCandidatesRequest{
		Namespace: "dev", Repository: "example.com/db"})
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"dev-db", "dev-db-old", "global-db"}, names(got.Documents), "without a digest, any version matches")
}

func TestBackend_Schema(t *testing.T) {
	ctx := context.Background()
	b := newTestBackend(t)

	_, err := b.Get(ctx, &protov4.GetRequest{Kind: protov4.Kind_KIND_SCHEMA, Id: "schema"})
	assert.Equal(t, codes.NotFound, status.Code(err))

	mustPut(t, b, protov4.PutRequest_UPSERT, doc(protov4.Kind_KIND_SCHEMA, map[string]interface{}{"_id": "schema", "installations": "1.0.2"}))
	mustPut(t, b, protov4.PutRequest_UPSERT, doc(protov4.Kind_KIND_SCHEMA, map[string]interface{}{"_id": "schema", "installations": "1.0.3"}))

	got, err := b.Get(ctx, &protov4.GetRequest{Kind: protov4.Kind_KIND_SCHEMA, Id: "schema"})
	require.NoError(t, err)
	var fields map[string]interface{}
	require.NoError(t, json.Unmarshal(got.Document.Json, &fields))
	assert.Equal(t, "1.0.3", fields["installations"])
}
