package v2

import (
	"testing"

	"get.porter.sh/porter/tests"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/stretchr/testify/require"
)

func TestDependencySource(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name               string
		bundleWiring       string
		wantSource         DependencySource
		wantWorkflowWiring string
		wantErr            string
	}{
		{
			name:               "not wiring",
			bundleWiring:       "${bundle.bloop}",
			wantSource:         DependencySource{Value: "${bundle.bloop}"},
			wantWorkflowWiring: "${bundle.bloop}",
		},
		{
			name:         "value",
			bundleWiring: "great data",
			wantSource: DependencySource{
				Value: "great data",
			},
			wantWorkflowWiring: "great data",
		},
		{
			name:         "parameter",
			bundleWiring: "bundle.parameters.color",
			wantSource: DependencySource{
				Parameter: "color",
			},
			wantWorkflowWiring: "workflow.jobs.1.parameters.color",
		},
		{
			name:         "credential",
			bundleWiring: "bundle.credentials.kubeconfig",
			wantSource: DependencySource{
				Credential: "kubeconfig",
			},
			wantWorkflowWiring: "workflow.jobs.1.credentials.kubeconfig",
		},
		{
			name:         "invalid: output",
			bundleWiring: "bundle.outputs.port",
			wantErr:      "cannot pass the root bundle output to a dependency",
		},
		{
			name:         "dependency parameter",
			bundleWiring: "bundle.dependencies.mysql.parameters.name",
			wantSource: DependencySource{
				Dependency: "mysql",
				Parameter:  "name",
			},
			wantWorkflowWiring: "workflow.jobs.1.parameters.name",
		},
		{
			name:         "dependency credential",
			bundleWiring: "bundle.dependencies.mysql.credentials.password",
			wantSource: DependencySource{
				Dependency: "mysql",
				Credential: "password",
			},
			wantWorkflowWiring: "workflow.jobs.1.credentials.password",
		},
		{
			name:         "dependency output",
			bundleWiring: "bundle.dependencies.mysql.outputs.connstr",
			wantSource: DependencySource{
				Dependency: "mysql",
				Output:     "connstr",
			},
			wantWorkflowWiring: "workflow.jobs.1.outputs.connstr",
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gotSource, err := ParseDependencySource(tc.bundleWiring)
			if tc.wantErr == "" {
				require.Equal(t, tc.wantSource, gotSource, "incorrect DependencySource was parsed")

				// Check that we can convert it back to a bundle wiring string
				gotBundleWiring := gotSource.AsBundleWiring()
				require.Equal(t, tc.bundleWiring, gotBundleWiring, "incorrect bundle wiring was returned")

				// Check that we can convert to a workflow wiring form
				gotWorkflowWiring := gotSource.AsWorkflowWiring("1")
				require.Equal(t, tc.wantWorkflowWiring, gotWorkflowWiring, "incorrect workflow wiring was returned")
			} else {
				tests.RequireErrorContains(t, err, tc.wantErr)
			}
		})
	}
}

func TestParseAllDependencySources(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name         string
		bundleWiring string
		want         []DependencySource
		wantInvalid  []string
	}{
		{
			name:         "hardcoded value",
			bundleWiring: "myenvdb",
			want:         []DependencySource{},
		},
		{
			name:         "own bundle parameter, not a dependency reference",
			bundleWiring: "${bundle.parameters.logLevel}",
			want:         []DependencySource{{Parameter: "logLevel"}},
		},
		{
			name:         "single dependency output reference",
			bundleWiring: "${bundle.dependencies.infra.outputs.mysql-connstr}",
			want:         []DependencySource{{Dependency: "infra", Output: "mysql-connstr"}},
		},
		{
			name:         "composite value with a cross-dependency ref and a same-dependency shorthand ref",
			bundleWiring: "https://${bundle.dependencies.infra.outputs.ip}:${outputs.port}/myapp",
			want:         []DependencySource{{Dependency: "infra", Output: "ip"}},
		},
		{
			name:         "root bundle output reference is reported as invalid, not silently dropped",
			bundleWiring: "${bundle.outputs.port}",
			want:         []DependencySource{},
			wantInvalid:  []string{"${bundle.outputs.port}"},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, gotInvalid := ParseAllDependencySources(tc.bundleWiring)
			require.Equal(t, tc.want, got, "incorrect DependencySources were parsed")
			require.Equal(t, tc.wantInvalid, gotInvalid, "incorrect invalid matches were reported")
		})
	}
}

func TestDependencyInterfaceDocument_IsEmpty(t *testing.T) {
	t.Parallel()

	require.True(t, DependencyInterfaceDocument{}.IsEmpty(), "a zero-value document should be empty")

	require.False(t, DependencyInterfaceDocument{
		Outputs: map[string]bundle.Output{"connstr": {}},
	}.IsEmpty(), "a document with an output should not be empty")

	require.False(t, DependencyInterfaceDocument{
		Parameters: map[string]bundle.Parameter{"logLevel": {}},
	}.IsEmpty(), "a document with a parameter should not be empty")

	require.False(t, DependencyInterfaceDocument{
		Credentials: map[string]bundle.Credential{"token": {}},
	}.IsEmpty(), "a document with a credential should not be empty")
}

func TestDependencyInterfaceDocument_Names(t *testing.T) {
	t.Parallel()

	doc := DependencyInterfaceDocument{
		Outputs: map[string]bundle.Output{
			"connstr": {}, "port": {},
		},
		Parameters: map[string]bundle.Parameter{
			"logLevel": {},
		},
		Credentials: map[string]bundle.Credential{
			"token": {}, "apiKey": {},
		},
	}

	outputs, parameters, credentials := doc.Names()
	require.Equal(t, []string{"connstr", "port"}, outputs)
	require.Equal(t, []string{"logLevel"}, parameters)
	require.Equal(t, []string{"apiKey", "token"}, credentials)

	emptyOutputs, emptyParameters, emptyCredentials := DependencyInterfaceDocument{}.Names()
	require.Empty(t, emptyOutputs)
	require.Empty(t, emptyParameters)
	require.Empty(t, emptyCredentials)
}
