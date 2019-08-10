package cnabprovider

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/deislabs/cnab-go/bundle"
	"github.com/deislabs/cnab-go/claim"
	"github.com/deislabs/cnab-go/driver/docker"

	"github.com/deislabs/porter/pkg/config"
)

func TestNewDriver_Docker(t *testing.T) {
	c := config.NewTestConfig(t)
	d := NewDuffle(c.Config)

	driver, err := d.newDriver("docker", "myclaim")
	require.NoError(t, err)

	if _, ok := driver.(*docker.Driver); ok {
		// TODO: check dockerConfigurationOptions to verify expected bind mount setup,
		// once we're able to (add ability to dockerdriver pkg)
	} else {
		t.Fatal("expected driver to be of type *dockerdriver.Driver")
	}
}

func TestWriteClaimOutputs(t *testing.T) {
	c := config.NewTestConfig(t)
	d := NewDuffle(c.Config)

	homeDir, err := c.GetHomeDir()
	require.NoError(t, err)

	c.TestContext.AddTestDirectory("../../porter/testdata/outputs", filepath.Join(homeDir, "outputs"))

	claim, err := claim.New("test-bundle")
	require.NoError(t, err)

	// Expect error when claim has no associated bundle
	err = d.WriteClaimOutputs(claim, "install")
	require.EqualError(t, err, "claim has no bundle")

	claim.Bundle = &bundle.Bundle{}

	// Expect no error if Bundle.Outputs is empty
	err = d.WriteClaimOutputs(claim, "install")
	require.NoError(t, err)

	claim.Bundle.Outputs = map[string]bundle.Output{
		"foo": bundle.Output{},
	}

	// Expect no error; by default, outputs apply to all actions
	err = d.WriteClaimOutputs(claim, "install")
	require.NoError(t, err)

	claim.Bundle.Outputs["foo"] = bundle.Output{
		ApplyTo: []string{
			"status",
		},
	}

	// Expect no error if output does not apply to action
	err = d.WriteClaimOutputs(claim, "install")
	require.NoError(t, err)
}
