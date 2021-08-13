package porter

import (
	"testing"

	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-go/bundle/definition"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func buildExampleBundle() bundle.Bundle {
	bunV, _ := bundle.GetDefaultSchemaVersion()
	bun := bundle.Bundle{
		SchemaVersion:    bunV,
		InvocationImages: []bundle.InvocationImage{{BaseImage: bundle.BaseImage{Image: "example.com/foo:v1.0.0"}}},
		Actions: map[string]bundle.Action{
			"blah": {
				Stateless: true,
			},
			"other": {
				Stateless: false,
				Modifies:  true,
			},
		},
		Definitions: map[string]*definition.Schema{
			"porter-debug-parameter": {
				Comment: "porter-internal",
				ID: "https://porter.sh/generated-bundle/#porter-debug",
				Default: false,
				Description: "Print debug information from Porter when executing the bundle",
				Type: "boolean",
			},
		},
		Parameters: map[string]bundle.Parameter{
			"porter-debug": {
					Definition: "porter-debug-parameter",
					Description: "Print debug information from Porter when executing the bundle",
					Destination: &bundle.Location{
						EnvironmentVariable: "PORTER_DEUBG",
					},
			},
		},
	}
	return bun
}

func TestInvokeOptions_Validate_ActionRequired(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Teardown()

	opts := NewInvokeOptions()

	err := opts.Validate(nil, p.Porter)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "--action is required")
}