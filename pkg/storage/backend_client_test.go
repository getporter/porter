package storage

import (
	"context"
	"errors"
	"testing"

	"get.porter.sh/porter/pkg/storage/plugins"
	"get.porter.sh/porter/pkg/storage/plugins/protov4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// fakeBackend records what it receives and returns canned responses.
type fakeBackend struct {
	plugins.StorageBackend

	put     *protov4.PutRequest
	getResp *protov4.GetResponse
	listRes *protov4.ListResponse
	err     error
}

func (f *fakeBackend) Put(ctx context.Context, req *protov4.PutRequest) (*protov4.PutResponse, error) {
	f.put = req
	return &protov4.PutResponse{}, f.err
}

func (f *fakeBackend) Get(ctx context.Context, req *protov4.GetRequest) (*protov4.GetResponse, error) {
	return f.getResp, f.err
}

func (f *fakeBackend) List(ctx context.Context, req *protov4.ListRequest) (*protov4.ListResponse, error) {
	return f.listRes, f.err
}

func TestToDocument(t *testing.T) {
	t.Run("installation", func(t *testing.T) {
		doc, err := toDocument(NewInstallation("dev", "mybuns"))
		require.NoError(t, err)
		assert.Equal(t, protov4.Kind_KIND_INSTALLATION, doc.Kind)
		assert.Equal(t, "dev", doc.Namespace)
		assert.Equal(t, "mybuns", doc.Name)
		assert.Contains(t, string(doc.Json), `"name":"mybuns"`)
	})

	t.Run("run", func(t *testing.T) {
		run := NewRun("dev", "mybuns")
		doc, err := toDocument(run)
		require.NoError(t, err)
		assert.Equal(t, protov4.Kind_KIND_RUN, doc.Kind)
		assert.Equal(t, run.ID, doc.Id)
		assert.Equal(t, "dev", doc.Namespace)
		assert.Equal(t, "mybuns", doc.Installation)
		assert.Contains(t, string(doc.Json), `"_id":"`+run.ID+`"`, "the calculated _id must be in the json")
	})

	t.Run("result", func(t *testing.T) {
		run := NewRun("dev", "mybuns")
		result := run.NewResult("running")
		doc, err := toDocument(result)
		require.NoError(t, err)
		assert.Equal(t, protov4.Kind_KIND_RESULT, doc.Kind)
		assert.Equal(t, result.ID, doc.Id)
		assert.Equal(t, run.ID, doc.RunId)
		assert.Equal(t, "mybuns", doc.Installation)
	})

	t.Run("output", func(t *testing.T) {
		doc, err := toDocument(Output{Name: "logs", Namespace: "dev", Installation: "mybuns", RunID: "r1", ResultID: "res1"})
		require.NoError(t, err)
		assert.Equal(t, protov4.Kind_KIND_OUTPUT, doc.Kind)
		assert.Equal(t, "logs", doc.Name)
		assert.Equal(t, "r1", doc.RunId)
		assert.Equal(t, "res1", doc.ResultId)
	})

	t.Run("credential set", func(t *testing.T) {
		doc, err := toDocument(NewCredentialSet("dev", "creds"))
		require.NoError(t, err)
		assert.Equal(t, protov4.Kind_KIND_CREDENTIAL_SET, doc.Kind)
		assert.Equal(t, "dev", doc.Namespace)
		assert.Equal(t, "creds", doc.Name)
	})

	t.Run("parameter set", func(t *testing.T) {
		doc, err := toDocument(NewParameterSet("dev", "params"))
		require.NoError(t, err)
		assert.Equal(t, protov4.Kind_KIND_PARAMETER_SET, doc.Kind)
		assert.Equal(t, "params", doc.Name)
	})

	t.Run("workflow", func(t *testing.T) {
		w := NewWorkflow("dev")
		doc, err := toDocument(w)
		require.NoError(t, err)
		assert.Equal(t, protov4.Kind_KIND_WORKFLOW, doc.Kind)
		assert.Equal(t, w.ID, doc.Id)
		assert.Equal(t, "dev", doc.Namespace)
	})

	t.Run("schema", func(t *testing.T) {
		doc, err := toDocument(NewSchema())
		require.NoError(t, err)
		assert.Equal(t, protov4.Kind_KIND_SCHEMA, doc.Kind)
		assert.Equal(t, "schema", doc.Id)
	})

	t.Run("unsupported", func(t *testing.T) {
		_, err := toDocument(struct{}{})
		require.Error(t, err)
	})
}

func TestBackendClient_put(t *testing.T) {
	ctx := context.Background()
	f := &fakeBackend{}
	c := NewBackendClient(f)

	inst := NewInstallation("dev", "mybuns")
	require.NoError(t, c.put(ctx, protov4.PutRequest_UPSERT, inst))
	require.NotNil(t, f.put)
	assert.Equal(t, protov4.PutRequest_UPSERT, f.put.Mode)
	require.Len(t, f.put.Documents, 1)
	assert.Equal(t, "mybuns", f.put.Documents[0].Name)

	t.Run("update of a missing document", func(t *testing.T) {
		f.err = status.Error(codes.NotFound, "installations not found")
		err := c.put(ctx, protov4.PutRequest_UPDATE, inst)
		assert.ErrorIs(t, err, ErrNotFound{})
	})
}

func TestGetDocument(t *testing.T) {
	ctx := context.Background()
	f := &fakeBackend{}
	c := NewBackendClient(f)

	want := NewInstallation("dev", "mybuns")
	doc, err := toDocument(want)
	require.NoError(t, err)
	f.getResp = &protov4.GetResponse{Document: doc}

	got, err := getDocument[Installation](ctx, c, &protov4.GetRequest{Kind: protov4.Kind_KIND_INSTALLATION, Namespace: "dev", Name: "mybuns"})
	require.NoError(t, err)
	assert.Equal(t, want.Name, got.Name)
	assert.Equal(t, want.Namespace, got.Namespace)

	t.Run("not found", func(t *testing.T) {
		f.getResp, f.err = nil, status.Error(codes.NotFound, "installations not found")
		_, err := getDocument[Installation](ctx, c, &protov4.GetRequest{Kind: protov4.Kind_KIND_INSTALLATION, Name: "missing"})
		require.ErrorIs(t, err, ErrNotFound{})

		var notFound ErrNotFound
		require.True(t, errors.As(err, &notFound))
		assert.Equal(t, CollectionInstallations, notFound.Collection)
		assert.Equal(t, "missing", notFound.Item)
	})

	t.Run("other errors pass through", func(t *testing.T) {
		boom := status.Error(codes.Internal, "boom")
		f.err = boom
		_, err := getDocument[Installation](ctx, c, &protov4.GetRequest{Kind: protov4.Kind_KIND_INSTALLATION})
		assert.NotErrorIs(t, err, ErrNotFound{})
		assert.Equal(t, codes.Internal, status.Code(err))
	})
}

func TestListDocuments(t *testing.T) {
	ctx := context.Background()
	f := &fakeBackend{}
	c := NewBackendClient(f)

	a, err := toDocument(NewInstallation("dev", "a"))
	require.NoError(t, err)
	b, err := toDocument(NewInstallation("dev", "b"))
	require.NoError(t, err)
	f.listRes = &protov4.ListResponse{Documents: []*protov4.Document{a, b}}

	got, err := listDocuments[Installation](ctx, c, &protov4.ListRequest{Kind: protov4.Kind_KIND_INSTALLATION, Namespace: "dev"})
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, "a", got[0].Name)
	assert.Equal(t, "b", got[1].Name)

	f.listRes = &protov4.ListResponse{}
	got, err = listDocuments[Installation](ctx, c, &protov4.ListRequest{Kind: protov4.Kind_KIND_INSTALLATION})
	require.NoError(t, err)
	assert.Empty(t, got)
}
