package porter

import (
	"testing"

	v2 "get.porter.sh/porter/pkg/cnab/extensions/dependencies/v2"
	"get.porter.sh/porter/pkg/secrets"
	"get.porter.sh/porter/pkg/storage"
	"github.com/cnabio/cnab-go/secrets/host"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWireJobParameters(t *testing.T) {
	t.Parallel()

	tg := newTestGraph()
	depA := tg.addNode("depA")
	depB := tg.addNode("depB")
	tg.addRequires(tg.g.Root, depA, "a")
	tg.addRequires(tg.g.Root, depB, "b")
	tg.g.addEdge(Edge{
		From: depA, To: depB, Kind: EdgeKindWiring, ToAlias: "b",
		Detail: &WiringDetail{Field: "parameters", FieldName: "connstr", SourceOutput: "connstr"},
	})

	order, err := tg.g.TopologicalOrder()
	require.NoError(t, err)
	jobIDs := buildJobIDs(order)

	dep := v2.Dependency{
		Parameters: map[string]string{
			"env":     "prod",
			"appName": "${bundle.parameters.name}",
			"connstr": "${bundle.dependencies.b.outputs.connstr}",
		},
	}
	rootParameters := secrets.Set{"name": "myapp"}

	job := &storage.Job{}
	err = wireJobParameters(job, dep, tg.g, depA, jobIDs, rootParameters, nil)
	require.NoError(t, err)

	params := job.Installation.Parameters.Parameters
	require.Len(t, params, 3)

	byName := make(map[string]secrets.SourceMap, len(params))
	for _, p := range params {
		byName[p.Name] = p
	}

	assert.Equal(t, host.SourceValue, byName["env"].Source.Strategy)
	assert.Equal(t, "prod", byName["env"].Source.Hint)

	assert.Equal(t, host.SourceValue, byName["appName"].Source.Strategy)
	assert.Equal(t, "myapp", byName["appName"].Source.Hint)

	assert.Equal(t, wiringStrategy, byName["connstr"].Source.Strategy)
	assert.Equal(t, "workflow.jobs."+jobIDs[depB]+".outputs.connstr", byName["connstr"].Source.Hint)
}

func TestWireJobCredentials(t *testing.T) {
	t.Parallel()

	tg := newTestGraph()
	dep := tg.addNode("dep")
	tg.addRequires(tg.g.Root, dep, "a")

	order, err := tg.g.TopologicalOrder()
	require.NoError(t, err)
	jobIDs := buildJobIDs(order)

	d := v2.Dependency{
		Credentials: map[string]string{
			"apiKey": "${bundle.credentials.masterKey}",
		},
	}
	// Same name as a root parameter: must resolve from credentials.
	rootParameters := secrets.Set{"masterKey": "wrong"}
	rootCredentials := secrets.Set{"masterKey": "s3cr3t"}

	job := &storage.Job{}
	err = wireJobCredentials(job, d, tg.g, dep, jobIDs, rootParameters, rootCredentials)
	require.NoError(t, err)

	require.Len(t, job.Credentials, 1)
	assert.Equal(t, "apiKey", job.Credentials[0].Name)
	assert.Equal(t, host.SourceValue, job.Credentials[0].Source.Strategy)
	assert.Equal(t, "s3cr3t", job.Credentials[0].Source.Hint)
}

func TestWireJobParameters_MissingRootValueErrors(t *testing.T) {
	t.Parallel()

	tg := newTestGraph()
	dep := tg.addNode("dep")
	tg.addRequires(tg.g.Root, dep, "a")

	order, err := tg.g.TopologicalOrder()
	require.NoError(t, err)
	jobIDs := buildJobIDs(order)

	d := v2.Dependency{Parameters: map[string]string{"env": "${bundle.parameters.missing}"}}

	job := &storage.Job{}
	err = wireJobParameters(job, d, tg.g, dep, jobIDs, secrets.Set{}, secrets.Set{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing")
}

func TestPropagateNamedSets(t *testing.T) {
	t.Parallel()

	parent := storage.NewInstallation("dev", "myapp")
	parent.CredentialSets = []string{"prod-creds"}
	parent.ParameterSets = []string{"prod-params"}

	dep := storage.NewInstallation("dev", "db")
	propagateNamedSets(&dep, parent)

	assert.Equal(t, parent.CredentialSets, dep.CredentialSets)
	assert.Equal(t, parent.ParameterSets, dep.ParameterSets)

	// Mutating the dependency's sets must not touch the parent's.
	dep.CredentialSets[0] = "changed"
	assert.Equal(t, "prod-creds", parent.CredentialSets[0])
}
