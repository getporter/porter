package runtime

import (
	"testing"

	"get.porter.sh/porter/pkg/manifest"
	"get.porter.sh/porter/pkg/yaml"
	"github.com/cnabio/cnab-go/claim"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestActionInput_MarshalYAML(t *testing.T) {
	s := &manifest.Step{
		Data: map[string]interface{}{
			"exec": map[string]interface{}{
				"command": "echo hi",
			},
		},
	}

	input := &ActionInput{
		action: claim.ActionInstall,
		Steps:  []*manifest.Step{s},
	}

	b, err := yaml.Marshal(input)
	require.NoError(t, err)
	wantYaml := `install:
  - exec:
      command: echo hi
`
	assert.Equal(t, wantYaml, string(b))
}
