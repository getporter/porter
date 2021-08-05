package porter

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDisplayValuesSort(t *testing.T) {
	v := DisplayValues{
		{Name: "b"},
		{Name: "c"},
		{Name: "a"},
	}

	sort.Sort(v)

	assert.Equal(t, "a", v[0].Name)
	assert.Equal(t, "b", v[1].Name)
	assert.Equal(t, "c", v[2].Name)
}

func TestGenerateParameterSet(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Teardown()

	p.TestConfig.TestContext.AddTestFile("testdata/bundle.json", "/bundle.json")

	opts := ParameterOptions{
		Silent: true,
	}
	opts.Namespace = "dev"
	opts.Name = "kool-params"
	opts.Labels = []string{"env=dev"}
	opts.CNABFile = "/bundle.json"
	err := opts.Validate(nil, p.Context)
	require.NoError(t, err, "Validate failed")

	err = p.GenerateParameters(opts)
	require.NoError(t, err, "no error should have existed")
	creds, err := p.Parameters.GetParameterSet(opts.Namespace, "kool-params")
	require.NoError(t, err, "expected parameter to have been generated")
	assert.Equal(t, map[string]string{"env": "dev"}, creds.Labels)
}
