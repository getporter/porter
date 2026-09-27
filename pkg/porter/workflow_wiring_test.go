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
	err = wireJobParameters(job, dep, tg.g, depA, jobIDs, jobIDs[tg.g.Root])
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
	err = wireJobCredentials(job, d, tg.g, dep, jobIDs, jobIDs[tg.g.Root])
	require.NoError(t, err)

	require.Len(t, job.Credentials, 1)
	assert.Equal(t, "apiKey", job.Credentials[0].Name)
	// The secret itself must never be written into the (persisted) workflow.
	assert.Equal(t, wiringStrategy, job.Credentials[0].Source.Strategy)
	assert.Equal(t, "workflow.jobs."+jobIDs[tg.g.Root]+".credentials.masterKey", job.Credentials[0].Source.Hint)
	assert.Empty(t, job.Credentials[0].ResolvedValue)
}

func TestWireJobParameters_CompositeTemplate(t *testing.T) {
	t.Parallel()

	tg := newTestGraph()
	depA := tg.addNode("depA")
	depB := tg.addNode("depB")
	tg.addRequires(tg.g.Root, depA, "a")
	tg.addRequires(tg.g.Root, depB, "b")
	// The graph builder creates a wiring edge for the sibling reference
	// inside the composite; it must not produce a second entry.
	tg.g.addEdge(Edge{
		From: depA, To: depB, Kind: EdgeKindWiring, ToAlias: "b",
		Detail: &WiringDetail{Field: "parameters", FieldName: "url", SourceOutput: "host"},
	})
	order, err := tg.g.TopologicalOrder()
	require.NoError(t, err)
	jobIDs := buildJobIDs(order)

	dep := v2.Dependency{Parameters: map[string]string{
		"url": "https://${bundle.dependencies.b.outputs.host}:${bundle.parameters.port}/app ${ bundle.credentials.token }",
	}}

	job := &storage.Job{}
	require.NoError(t, wireJobParameters(job, dep, tg.g, depA, jobIDs, jobIDs[tg.g.Root]))

	params := job.Installation.Parameters.Parameters
	require.Len(t, params, 1)
	assert.Equal(t, "url", params[0].Name)
	assert.Equal(t, wiringTemplateStrategy, params[0].Source.Strategy)
	root, sibling := jobIDs[tg.g.Root], jobIDs[depB]
	// Literal text and whitespace preserved; references rewritten to job
	// form; no secret value anywhere in the stored hint.
	assert.Equal(t,
		"https://${workflow.jobs."+sibling+".outputs.host}:${workflow.jobs."+root+".parameters.port}/app ${workflow.jobs."+root+".credentials.token}",
		params[0].Source.Hint)
	assert.Empty(t, params[0].ResolvedValue)
}

func TestWireJobParameters_CompositeTemplateErrors(t *testing.T) {
	t.Parallel()

	tests := map[string]string{
		"root output (never resolvable)":       "x-${bundle.outputs.x}",
		"unknown sibling":                      "x-${bundle.dependencies.nope.outputs.host}",
		"a dependency's parameter, not output": "x-${bundle.dependencies.b.parameters.p}",
		"whole ref, unknown sibling":           "${bundle.dependencies.nope.outputs.host}",
		"whole ref, dependency parameter":      "${bundle.dependencies.b.parameters.port}",
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
			err = wireJobParameters(job, v2.Dependency{Parameters: map[string]string{"env": value}}, tg.g, dep, jobIDs, jobIDs[tg.g.Root])
			require.Error(t, err)
			assert.Empty(t, job.Installation.Parameters.Parameters)
		})
	}
}

func TestWireJobParameters_UsesParentJobNotRoot(t *testing.T) {
	t.Parallel()

	// root -> svc -> db: db's ${bundle.parameters.x} and
	// ${bundle.credentials.y} are svc's (the bundle that declared db), not
	// the root's.
	tg := newTestGraph()
	svc := tg.addNode("svc")
	db := tg.addNode("db")
	tg.addRequires(tg.g.Root, svc, "svc")
	tg.addRequires(svc, db, "db")
	order, err := tg.g.TopologicalOrder()
	require.NoError(t, err)
	jobIDs := buildJobIDs(order)

	dep := v2.Dependency{
		Parameters:  map[string]string{"p": "${bundle.parameters.x}", "t": "a-${bundle.parameters.x}"},
		Credentials: map[string]string{"c": "${bundle.credentials.y}"},
	}

	job := &storage.Job{}
	require.NoError(t, wireJobParameters(job, dep, tg.g, db, jobIDs, jobIDs[svc]))
	require.NoError(t, wireJobCredentials(job, dep, tg.g, db, jobIDs, jobIDs[svc]))

	byName := map[string]string{}
	for _, p := range job.Installation.Parameters.Parameters {
		byName[p.Name] = p.Source.Hint
	}
	assert.Equal(t, "workflow.jobs."+jobIDs[svc]+".parameters.x", byName["p"])
	assert.Equal(t, "a-${workflow.jobs."+jobIDs[svc]+".parameters.x}", byName["t"])
	require.Len(t, job.Credentials, 1)
	assert.Equal(t, "workflow.jobs."+jobIDs[svc]+".credentials.y", job.Credentials[0].Source.Hint)
	assert.NotContains(t, byName["p"], jobIDs[tg.g.Root])
}

func TestWireJobParameters_MergesByName(t *testing.T) {
	t.Parallel()

	tg := newTestGraph()
	dep := tg.addNode("dep")
	tg.addRequires(tg.g.Root, dep, "a")
	order, err := tg.g.TopologicalOrder()
	require.NoError(t, err)
	jobIDs := buildJobIDs(order)

	// A reused installation already carrying overrides, one for a name the
	// mapping also wires.
	original := secrets.StrategyList{storage.ValueStrategy("env", "stale"), storage.ValueStrategy("keep", "me")}
	job := &storage.Job{}
	job.Installation.Parameters.Parameters = original

	d := v2.Dependency{Parameters: map[string]string{"env": "prod", "new": "x"}}
	require.NoError(t, wireJobParameters(job, d, tg.g, dep, jobIDs, jobIDs[tg.g.Root]))

	got := job.Installation.Parameters.Parameters
	require.Len(t, got, 3, "no duplicate names")
	assert.Equal(t, []string{"env", "keep", "new"}, []string{got[0].Name, got[1].Name, got[2].Name})
	assert.Equal(t, "prod", got[0].Source.Hint, "wired value replaces the existing override")
	assert.Equal(t, "me", got[1].Source.Hint)

	// The list the job started with (shared with a reused installation)
	// must be untouched.
	assert.Equal(t, "stale", original[0].Source.Hint)
}

func TestWireJobParameters_LiteralThatLooksLikeAReference(t *testing.T) {
	t.Parallel()

	tg := newTestGraph()
	dep := tg.addNode("dep")
	tg.addRequires(tg.g.Root, dep, "a")
	order, err := tg.g.TopologicalOrder()
	require.NoError(t, err)
	jobIDs := buildJobIDs(order)

	// Not a parameters/credentials/outputs reference, so it's just text.
	d := v2.Dependency{Parameters: map[string]string{"url": "https://example.com/bundle.foo.bar"}}
	job := &storage.Job{}
	require.NoError(t, wireJobParameters(job, d, tg.g, dep, jobIDs, jobIDs[tg.g.Root]))

	require.Len(t, job.Installation.Parameters.Parameters, 1)
	assert.Equal(t, host.SourceValue, job.Installation.Parameters.Parameters[0].Source.Strategy)
	assert.Equal(t, "https://example.com/bundle.foo.bar", job.Installation.Parameters.Parameters[0].Source.Hint)
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
