package kubernetes

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMixin_PrintSchema(t *testing.T) {
	m := NewTestMixin(t)

	err := m.PrintSchema()
	require.NoError(t, err)

	gotSchema := m.TestContext.GetOutput()

	wantSchema, err := ioutil.ReadFile("testdata/schema.json")
	require.NoError(t, err)

	assert.Equal(t, string(wantSchema), gotSchema)
}

func TestMixin_ValidatePayload(t *testing.T) {
	testcases := []struct {
		name  string
		step  string
		pass  bool
		error string
	}{
		{"install", "testdata/install-input.yaml", true, ""},
		{"upgrade", "testdata/upgrade-input.yaml", true, ""},
		{"uninstall", "testdata/uninstall-input.yaml", true, ""},
		{"install-default-manifest-ok", "testdata/install-input-no-manifests.yaml", true, ""},
		{"install-bad-wait-flag", "testdata/install-input-bad-wait-flag.yaml", false, "install.0.kubernetes.wait: Invalid type. Expected: boolean, given: string"},
		{"install-bad-outputs", "testdata/install-input-bad-outputs.yaml", false, "install.0.kubernetes.outputs.0: resourceType is required\n\t* install.0.kubernetes.outputs.0: resourceName is required\n\t* install.0.kubernetes.outputs.0: jsonPath is required"},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			m := NewTestMixin(t)
			b, err := ioutil.ReadFile(tc.step)
			require.NoError(t, err)

			err = m.ValidatePayload(b)
			if tc.pass {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.error)
			}
		})
	}
}
