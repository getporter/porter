package porter

import (
	"os"
	"testing"

	"github.com/deislabs/porter/pkg/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveMapParam(t *testing.T) {
	p := NewTestPorter(t)
	p.Config.Manifest = &config.Manifest{
		Parameters: []config.ParameterDefinition{
			config.ParameterDefinition{
				Name: "person",
			},
		},
	}

	os.Setenv("PERSON", "Ralpha")
	s := &config.Step{
		Description: "a test step",
		Data: map[string]interface{}{
			"Parameters": map[string]interface{}{
				"Thing": map[string]interface{}{
					"source": "bundle.parameters.person",
				},
			},
		},
	}

	err := p.resolveSourcedValues(s)
	require.NoError(t, err)
	pms, ok := s.Data["Parameters"].(map[string]interface{})
	assert.True(t, ok)
	val, ok := pms["Thing"].(string)
	assert.True(t, ok)
	assert.Equal(t, "Ralpha", val)
}

func TestResolveMapParamUnknown(t *testing.T) {

	p := NewTestPorter(t)
	p.Config.Manifest = &config.Manifest{
		Parameters: []config.ParameterDefinition{},
	}

	s := &config.Step{
		Description: "a test step",
		Data: map[string]interface{}{
			"Parameters": map[string]interface{}{
				"Thing": map[string]interface{}{
					"source": "bundle.parameters.person",
				},
			},
		},
	}
	err := p.resolveSourcedValues(s)
	require.Error(t, err)
	assert.Equal(t, "unable to source value: unable to find parameter", err.Error())
}

func TestResolveArrayUnknown(t *testing.T) {
	p := NewTestPorter(t)
	p.Config.Manifest = &config.Manifest{
		Parameters: []config.ParameterDefinition{
			config.ParameterDefinition{
				Name: "name",
			},
		},
	}

	s := &config.Step{
		Description: "a test step",
		Data: map[string]interface{}{
			"Arguments": []string{
				"source: bundle.parameters.person",
			},
		},
	}

	err := p.resolveSourcedValues(s)
	require.Error(t, err)
	assert.Equal(t, "unable to source value: unable to find parameter", err.Error())
}

func TestResolveArray(t *testing.T) {
	p := NewTestPorter(t)
	p.Config.Manifest = &config.Manifest{
		Parameters: []config.ParameterDefinition{
			config.ParameterDefinition{
				Name: "person",
			},
		},
	}

	os.Setenv("PERSON", "Ralpha")
	s := &config.Step{
		Description: "a test step",
		Data: map[string]interface{}{
			"Arguments": []string{
				"source: bundle.parameters.person",
			},
		},
	}

	err := p.resolveSourcedValues(s)
	require.NoError(t, err)
	args, ok := s.Data["Arguments"].([]string)
	assert.True(t, ok)
	assert.Equal(t, "Ralpha", args[0])
}
