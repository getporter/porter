package porter

import (
	"path/filepath"
	"testing"

	"github.com/deislabs/porter/pkg/config"

	"github.com/deislabs/cnab-go/claim"
	"github.com/stretchr/testify/require"
)

func TestPorter_fetchBundleOutputs_Error(t *testing.T) {
	p := NewTestPorter(t)
	p.TestConfig.SetupPorterHome()

	homeDir, err := p.TestConfig.GetHomeDir()
	require.NoError(t, err)

	p.TestConfig.TestContext.AddTestDirectory("testdata/outputs", filepath.Join(homeDir, "outputs"))

	_, err = p.fetchBundleOutputs("bad-outputs-bundle")
	require.EqualError(t, err,
		"unable to read output 'bad-output' for claim 'bad-outputs-bundle': unable to unmarshal output \"bad-output\" for claim \"bad-outputs-bundle\"")
}

func TestPorter_fetchBundleOutputs(t *testing.T) {
	p := NewTestPorter(t)
	p.TestConfig.SetupPorterHome()

	homeDir, err := p.TestConfig.GetHomeDir()
	require.NoError(t, err)

	p.TestConfig.TestContext.AddTestDirectory("testdata/outputs", filepath.Join(homeDir, "outputs"))

	got, err := p.fetchBundleOutputs("test-bundle")
	require.NoError(t, err)

	want := &config.Outputs{
		{
			Name:      "foo",
			Type:      "string",
			Value:     "foo-value",
			Sensitive: true,
		},
		{
			Name:      "bar",
			Type:      "string",
			Value:     "bar-value",
			Sensitive: false,
		},
	}

	require.Equal(t, want, got)
}

func TestPorter_printOutputsTable(t *testing.T) {
	p := NewTestPorter(t)
	p.TestConfig.SetupPorterHome()
	p.CNAB = NewTestCNABProvider()

	outputs := &config.Outputs{
		{
			Name:      "foo",
			Type:      "string",
			Value:     "foo-value",
			Sensitive: true,
		},
		{
			Name:      "bar",
			Type:      "string",
			Value:     "bar-value",
			Sensitive: false,
		},
	}

	// Create test claim
	claim := claim.Claim{
		Name: "test-bundle",
	}
	if testy, ok := p.CNAB.(*TestCNABProvider); ok {
		testy.CreateClaim(&claim)
	} else {
		t.Fatal("expected p.CNAB to be of type *TestCNABProvider")
	}

	want := `-----------------------------------------------------
  Name  Type    Value (Path if sensitive)              
-----------------------------------------------------
  foo   string  /root/.porter/outputs/test-bundle/foo  
  bar   string  bar-value                              
`

	err := p.printOutputsTable(outputs, "test-bundle")
	require.NoError(t, err)

	got := p.TestConfig.TestContext.GetOutput()
	require.Equal(t, want, got)
}
