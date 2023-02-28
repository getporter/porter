package porter

import (
	"testing"

	"get.porter.sh/porter/pkg/encoding"
	"github.com/carolynvs/aferox"
	"github.com/spf13/afero"

	"get.porter.sh/porter/pkg/secrets"
	"get.porter.sh/porter/pkg/storage"
	"get.porter.sh/porter/pkg/test"
	"get.porter.sh/porter/pkg/yaml"
	"github.com/stretchr/testify/require"
)

// Check that AsSpecOnly results in a workflow that doesn't print anything for non-user settable fields
func TestDisplayWorkflow_AsSpecOnly(t *testing.T) {
	w := buildTestWorkflow()
	dw := NewDisplayWorkflow(w).AsSpecOnly()
	result, err := yaml.Marshal(dw)
	require.NoError(t, err, "Marshall failed")
	test.CompareGoldenFile(t, "testdata/workflow/workflow-spec-only.yaml", string(result))
}

func buildTestWorkflow() storage.Workflow {
	return storage.Workflow{
		ID: "abc123",
		WorkflowSpec: storage.WorkflowSpec{
			SchemaType:    storage.SchemaTypeWorkflow,
			SchemaVersion: storage.WorkflowSchemaVersion,
			MaxParallel:   1,
			DebugMode:     false,
			Stages: []storage.Stage{
				{
					Jobs: map[string]*storage.Job{
						"root": {
							Key:    "root",
							Action: "apply",
							Installation: storage.Installation{
								ID: "abc123",
								InstallationSpec: storage.InstallationSpec{
									SchemaVersion: storage.InstallationSchemaVersion,
									Name:          "mybuns",
									Namespace:     "myns",
									Uninstalled:   false,
									Bundle: storage.OCIReferenceParts{
										Repository: "example.com",
										Version:    "1.0.0",
										Digest:     "sha256:992f38e1a81cbdf2903fab2603f6076f3bef54262b4b2aa5b196515bac953688",
										Tag:        "v1.0.0",
									},
									Custom:         map[string]interface{}{"color": "blue"},
									Labels:         map[string]string{"team": "red"},
									CredentialSets: []string{"mycs"},
									Credentials: storage.CredentialSetSpec{
										SchemaVersion: storage.CredentialSetSchemaVersion,
										Namespace:     "",
										Name:          "internal-cs",
										Credentials: []secrets.Strategy{
											{
												Name:   "password",
												Source: secrets.Source{Key: "secret", Value: "mypassword"},
												Value:  "TOPSECRET",
											},
										},
									},
									ParameterSets: []string{"myps"},
									Parameters: storage.ParameterSet{
										ParameterSetSpec: storage.ParameterSetSpec{
											SchemaVersion: storage.ParameterSetSchemaVersion,
											Namespace:     "myns",
											Name:          "myps",
											Labels:        nil,
											Parameters: []secrets.Strategy{
												{
													Name:   "logLevel",
													Source: secrets.Source{Key: "value", Value: "11"},
													Value:  "11",
												},
											},
										},
										Status: storage.ParameterSetStatus{
											Created:  now,
											Modified: now,
										},
									},
								},
								Status: storage.InstallationStatus{
									RunID:           "abc123",
									Action:          "install",
									ResultID:        "abc123",
									ResultStatus:    "failed",
									Created:         now,
									Modified:        now,
									Installed:       &now,
									Uninstalled:     &now,
									BundleReference: "abc123",
									BundleVersion:   "abc123",
									BundleDigest:    "abc123",
								},
							},
							Depends: []string{"other-thing"},
							Status: storage.JobStatus{
								LastRunID:    "abc123",
								LastResultID: "abc123",
								ResultIDs:    []string{"1", "2"},
								Status:       "mystatus",
								Message:      "mymessage",
							},
						},
					},
				},
			},
		},
		Status: storage.WorkflowStatus{},
	}
}

func TestNewWorkflowFromDisplayWorkflow(t *testing.T) {
	fs := aferox.NewAferox(".", afero.NewOsFs())

	var dw DisplayWorkflow
	err := encoding.UnmarshalFile(fs, "testdata/workflow/workflow-spec-only.yaml", &dw)
	require.NoError(t, err, "failed to read testdata")

	w, err := dw.ToWorkflow()
	require.NoError(t, err, "ToWorkflow failed")

	// Prepare the workflow for a golden file comparison, setting unstable values to well-known values for easier comparison
	w.ID = "01GSC42NYVWGXCFAWWNTSHX2N3"
	rootJob := w.Stages[0].Jobs["root"]
	rootJob.Installation.Status.Created = now
	rootJob.Installation.Status.Modified = now
	rootJob.Installation.Parameters.Status.Created = now
	rootJob.Installation.Parameters.Status.Modified = now

	result, err := yaml.Marshal(w)
	require.NoError(t, err, "Marshall failed")
	test.CompareGoldenFile(t, "testdata/workflow/imported-workflow.yaml", string(result))
}
