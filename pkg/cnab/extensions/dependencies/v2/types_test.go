package v2

import (
	"testing"

	"get.porter.sh/porter/tests"
	"github.com/stretchr/testify/require"
)

func TestDependencySource(t *testing.T) {
	t.Parallel()

	jobKey := "1"
	testcases := []struct {
		name               string
		bundleWiring       string
		wantSource         DependencySource
		wantWorkflowWiring string
		wantErr            string
	}{
		{ // Check that we can still pass hard-coded values in a workflow
			name:         "value",
			bundleWiring: "11",
			wantSource: DependencySource{
				Value: "11",
			},
			wantWorkflowWiring: "11",
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
				gotWorkflowWiringValue := gotSource.AsWorkflowWiring(jobKey)
				require.Equal(t, tc.wantWorkflowWiring, gotWorkflowWiringValue, "incorrect workflow wiring string value was returned")
			} else {
				tests.RequireErrorContains(t, err, tc.wantErr)
			}
		})
	}
}

func TestParseWorkflowWiring(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name               string
		wiringStr          string
		wantWorkflowWiring WorkflowWiring
		wantErr            string
	}{
		{ // Check that we can still pass hard-coded values in a workflow
			name:      "value not supported",
			wiringStr: "11",
			wantErr:   "invalid workflow wiring",
		},
		{
			name:      "parameter",
			wiringStr: "workflow.abc123.jobs.myjerb.parameters.logLevel",
			wantWorkflowWiring: WorkflowWiring{
				WorkflowID: "abc123",
				JobKey:     "myjerb",
				Parameter:  "logLevel",
			},
		},
		{
			name:      "credential",
			wiringStr: "workflow.myworkflow.jobs.root.credentials.kubeconfig",
			wantWorkflowWiring: WorkflowWiring{
				WorkflowID: "myworkflow",
				JobKey:     "root",
				Credential: "kubeconfig",
			},
		},
		{
			name:      "output",
			wiringStr: "workflow.abc123.jobs.mydb.outputs.connstr",
			wantWorkflowWiring: WorkflowWiring{
				WorkflowID: "abc123",
				JobKey:     "mydb",
				Output:     "connstr",
			},
		},
		{
			name:      "dependencies not allowed",
			wiringStr: "workflow.abc123.jobs.root.dependencies.mydb.outputs.connstr",
			wantErr:   "invalid workflow wiring",
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gotWiring, err := ParseWorkflowWiring(tc.wiringStr)
			if tc.wantErr == "" {
				require.Equal(t, tc.wantWorkflowWiring, gotWiring, "incorrect WorkflowWiring was parsed")
			} else {
				tests.RequireErrorContains(t, err, tc.wantErr)
			}
		})
	}
}
