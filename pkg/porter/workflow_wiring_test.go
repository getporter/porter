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

	job := &storage.Job{}
	err = wireJobParameters(job, dep, tg.g, depA, jobIDs)
	require.NoError(t, err)

	params := job.Installation.Parameters.Parameters
	require.Len(t, params, 3)

	byName := make(map[string]secrets.SourceMap, len(params))
	for _, p := range params {
		byName[p.Name] = p
	}

	assert.Equal(t, host.SourceValue, byName["env"].Source.Strategy)
	assert.Equal(t, "prod", byName["env"].Source.Hint)

	// Root parameter is referenced via the root job, never resolved here.
	assert.Equal(t, wiringStrategy, byName["appName"].Source.Strategy)
	assert.Equal(t, "workflow.jobs."+jobIDs[tg.g.Root]+".parameters.name", byName["appName"].Source.Hint)

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

	job := &storage.Job{}
	err = wireJobCredentials(job, d, tg.g, dep, jobIDs)
	require.NoError(t, err)

	require.Len(t, job.Credentials, 1)
	assert.Equal(t, "apiKey", job.Credentials[0].Name)
	// The secret itself must never be written into the (persisted) workflow.
	assert.Equal(t, wiringStrategy, job.Credentials[0].Source.Strategy)
	assert.Equal(t, "workflow.jobs."+jobIDs[tg.g.Root]+".credentials.masterKey", job.Credentials[0].Source.Hint)
	assert.Empty(t, job.Credentials[0].ResolvedValue)
}

func TestWireJobParameters_CompositeTemplateErrors(t *testing.T) {
	t.Parallel()

	tests := map[string]string{
		"reference with literal text":    "https://${bundle.dependencies.db.outputs.host}",
		"two references":                 "${bundle.dependencies.db.outputs.host}:${bundle.parameters.port}",
		"root reference with text":       "prefix-${bundle.parameters.name}",
		"root output (never resolvable)": "${bundle.outputs.x}",
	}
	for name, value := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tg := newTestGraph()
			dep := tg.addNode("dep")
			tg.addRequires(tg.g.Root, dep, "a")
			order, err := tg.g.TopologicalOrder()
			require.NoError(t, err)
			jobIDs := buildJobIDs(order)

			job := &storage.Job{}
			err = wireJobParameters(job, v2.Dependency{Parameters: map[string]string{"env": value}}, tg.g, dep, jobIDs)
			require.Error(t, err)
			assert.Empty(t, job.Installation.Parameters.Parameters)
		})
	}
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
