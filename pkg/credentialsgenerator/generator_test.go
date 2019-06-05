package credentialsgenerator

import (
	"fmt"
	"testing"

	"github.com/deislabs/duffle/pkg/bundle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBadName(t *testing.T) {
	name := "this.hasadot"
	opts := GenerateOptions{
		Name:   name,
		Silent: true,
	}

	cs, err := GenerateCredentials(opts)
	require.Error(t, err, "bad name should have resulted in an error")
	require.Nil(t, cs, "credential set should have been empty")
	require.EqualError(t, err, fmt.Sprintf("credentialset name '%s' cannot contain the following characters: './\\'", name))
}

func TestGoodName(t *testing.T) {
	name := "this-name-is-valid"
	opts := GenerateOptions{
		Name:   name,
		Silent: true,
		Credentials: map[string]bundle.Location{
			"one": bundle.Location{
				EnvironmentVariable: "BLAH",
			},
			"two": bundle.Location{
				EnvironmentVariable: "BLAH",
			},
			"three": bundle.Location{
				Path: "/something",
			},
		},
	}

	cs, err := GenerateCredentials(opts)
	require.NoError(t, err, "name should NOT have resulted in an error")
	require.NotNil(t, cs, "credential set should have been empty")
	assert.Equal(t, 3, len(cs.Credentials), "should have had a single record")
}

func TestNoCredentials(t *testing.T) {
	name := "this-name-is-valid"
	opts := GenerateOptions{
		Name:   name,
		Silent: true,
	}
	cs, err := GenerateCredentials(opts)
	require.NoError(t, err, "no credentials should have generated an empty credential set")
	require.NotNil(t, cs, "empty credentials should still return an empty credential set")
}

func TestEmptyCredentials(t *testing.T) {
	name := "this-name-is-valid"
	opts := GenerateOptions{
		Name:        name,
		Silent:      true,
		Credentials: map[string]bundle.Location{},
	}
	cs, err := GenerateCredentials(opts)
	require.NoError(t, err, "no credentials should have generated an empty credential set")
	require.NotNil(t, cs, "empty credentials should still return an empty credential set")
}

func TestNoName(t *testing.T) {
	opts := GenerateOptions{
		Silent: true,
	}
	_, err := GenerateCredentials(opts)
	require.Error(t, err, "expected an error because name is required")
}
