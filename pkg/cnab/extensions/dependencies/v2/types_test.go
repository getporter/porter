package v2

import (
	"testing"

	"get.porter.sh/porter/tests"
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
