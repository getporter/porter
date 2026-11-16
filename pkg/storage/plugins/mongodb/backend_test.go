package mongodb

import (
	"encoding/json"
	"testing"

	"get.porter.sh/porter/pkg/storage/plugins/protov4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestCollectionFor(t *testing.T) {
	tests := map[protov4.Kind]string{
		protov4.Kind_KIND_INSTALLATION:   "installations",
		protov4.Kind_KIND_RUN:            "runs",
		protov4.Kind_KIND_RESULT:         "results",
		protov4.Kind_KIND_OUTPUT:         "outputs",
		protov4.Kind_KIND_CREDENTIAL_SET: "credentials",
		protov4.Kind_KIND_PARAMETER_SET:  "parameters",
		protov4.Kind_KIND_WORKFLOW:       "workflows",
		protov4.Kind_KIND_SCHEMA:         "config",
	}
	for kind, want := range tests {
		got, err := collectionFor(kind)
		require.NoError(t, err, kind.String())
		assert.Equal(t, want, got, kind.String())
	}

	_, err := collectionFor(protov4.Kind_KIND_UNSPECIFIED)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

// The filters must match the document filters that v3 used, so that v4 can
// replace documents written by v3.
func TestKeyFilter(t *testing.T) {
	d := &protov4.Document{Id: "id1", Namespace: "ns", Name: "n", ResultId: "res1"}

	tests := []struct {
		kind protov4.Kind
		want bson.M
	}{
		{protov4.Kind_KIND_INSTALLATION, bson.M{"namespace": "ns", "name": "n"}},
		{protov4.Kind_KIND_CREDENTIAL_SET, bson.M{"namespace": "ns", "name": "n"}},
		{protov4.Kind_KIND_PARAMETER_SET, bson.M{"namespace": "ns", "name": "n"}},
		{protov4.Kind_KIND_RUN, bson.M{"_id": "id1"}},
		{protov4.Kind_KIND_RESULT, bson.M{"_id": "id1"}},
		{protov4.Kind_KIND_WORKFLOW, bson.M{"namespace": "ns", "_id": "id1"}},
		{protov4.Kind_KIND_OUTPUT, bson.M{"resultId": "res1", "name": "n"}},
		{protov4.Kind_KIND_SCHEMA, bson.M{"_id": "id1"}},
	}
	for _, tt := range tests {
		d.Kind = tt.kind
		assert.Equal(t, tt.want, keyFilter(d), tt.kind.String())
	}
}

func TestGetFilter(t *testing.T) {
	assert.Equal(t, bson.M{"_id": "id1"},
		getFilter(&protov4.GetRequest{Kind: protov4.Kind_KIND_RUN, Id: "id1"}))
	assert.Equal(t, bson.M{"namespace": "ns", "name": "n"},
		getFilter(&protov4.GetRequest{Kind: protov4.Kind_KIND_INSTALLATION, Namespace: "ns", Name: "n"}))
	assert.Equal(t, bson.M{"namespace": "ns", "_id": "w1"},
		getFilter(&protov4.GetRequest{Kind: protov4.Kind_KIND_WORKFLOW, Namespace: "ns", Id: "w1"}))
}

func TestRawToDocument(t *testing.T) {
	raw, err := bson.Marshal(bson.M{
		"_id":          "run1",
		"namespace":    "ns",
		"installation": "mybuns",
		"runId":        "run1",
		"nested":       bson.M{"a": "b"},
	})
	require.NoError(t, err)

	doc, err := rawToDocument(protov4.Kind_KIND_RESULT, raw)
	require.NoError(t, err)

	assert.Equal(t, protov4.Kind_KIND_RESULT, doc.Kind)
	assert.Equal(t, "run1", doc.Id)
	assert.Equal(t, "ns", doc.Namespace)
	assert.Equal(t, "mybuns", doc.Installation)
	assert.Equal(t, "run1", doc.RunId)
	assert.Empty(t, doc.Name)

	var got map[string]interface{}
	require.NoError(t, json.Unmarshal(doc.Json, &got))
	assert.Equal(t, map[string]interface{}{"a": "b"}, got["nested"])
}

func TestJsonToDocument(t *testing.T) {
	doc, err := jsonToDocument([]byte(`{"_id":"abc","name":"foo"}`))
	require.NoError(t, err)
	assert.Equal(t, "abc", doc["_id"])

	_, err = jsonToDocument([]byte(`not json`))
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}
