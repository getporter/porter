package porter

import (
	"sort"
	"testing"

	"get.porter.sh/porter/pkg/parameters"
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
	err := opts.Validate(nil, p.Porter)
	require.NoError(t, err, "Validate failed")

	err = p.GenerateParameters(opts)
	require.NoError(t, err, "no error should have existed")
	creds, err := p.Parameters.GetParameterSet(opts.Namespace, "kool-params")
	require.NoError(t, err, "expected parameter to have been generated")
	assert.Equal(t, map[string]string{"env": "dev"}, creds.Labels)
}

func TestPorter_ListParameters(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Teardown()

	p.TestParameters.InsertParameterSet(parameters.NewParameterSet("", "shared-mysql"))
	p.TestParameters.InsertParameterSet(parameters.NewParameterSet("dev", "carolyn-wordpress"))
	p.TestParameters.InsertParameterSet(parameters.NewParameterSet("dev", "vaughn-wordpress"))
	p.TestParameters.InsertParameterSet(parameters.NewParameterSet("test", "staging-wordpress"))
	p.TestParameters.InsertParameterSet(parameters.NewParameterSet("test", "iat-wordpress"))
	p.TestParameters.InsertParameterSet(parameters.NewParameterSet("test", "shared-mysql"))

	t.Run("all-namespaces", func(t *testing.T) {
		opts := ListOptions{AllNamespaces: true}
		results, err := p.ListParameters(opts)
		require.NoError(t, err)
		assert.Len(t, results, 6)
	})

	t.Run("local namespace", func(t *testing.T) {
		opts := ListOptions{Namespace: "dev"}
		results, err := p.ListParameters(opts)
		require.NoError(t, err)
		assert.Len(t, results, 2)

		opts = ListOptions{Namespace: "test"}
		results, err = p.ListParameters(opts)
		require.NoError(t, err)
		assert.Len(t, results, 3)
	})

	t.Run("global namespace", func(t *testing.T) {
		opts := ListOptions{Namespace: ""}
		results, err := p.ListParameters(opts)
		require.NoError(t, err)
		assert.Len(t, results, 1)
	})
}
