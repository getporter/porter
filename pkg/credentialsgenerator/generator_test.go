package credentialsgenerator

import (
	"fmt"
	"os"
	"testing"

	"github.com/cnabio/cnab-go/bundle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/AlecAivazis/survey.v1"
)

func TestBadName(t *testing.T) {
	name := "this.hasadot"
	opts := GenerateOptions{
		Name:   name,
		Silent: true,
	}

	cs, err := GenerateCredentials(opts, survey.WithStdio(os.Stdin, os.Stdout, os.Stderr))
	require.Error(t, err, "bad name should have resulted in an error")
	require.Nil(t, cs, "credential set should have been empty")
	require.EqualError(t, err, fmt.Sprintf("credentialset name '%s' cannot contain the following characters: './\\'", name))
}

func TestGoodName(t *testing.T) {
	name := "this-name-is-valid"
	opts := GenerateOptions{
		Name:   name,
		Silent: true,
		Credentials: map[string]bundle.Credential{
			"one": {
				Location: bundle.Location{
					EnvironmentVariable: "BLAH",
				},
			},
			"two": {
				Location: bundle.Location{
					EnvironmentVariable: "BLAH",
				},
			},
			"three": {
				Location: bundle.Location{
					Path: "/something",
				},
			},
		},
	}

	cs, err := GenerateCredentials(opts, survey.WithStdio(os.Stdin, os.Stdout, os.Stderr))
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
	cs, err := GenerateCredentials(opts, survey.WithStdio(os.Stdin, os.Stdout, os.Stderr))
	require.NoError(t, err, "no credentials should have generated an empty credential set")
	require.NotNil(t, cs, "empty credentials should still return an empty credential set")
}

func TestEmptyCredentials(t *testing.T) {
	name := "this-name-is-valid"
	opts := GenerateOptions{
		Name:        name,
		Silent:      true,
		Credentials: map[string]bundle.Credential{},
	}
	cs, err := GenerateCredentials(opts, survey.WithStdio(os.Stdin, os.Stdout, os.Stderr))
	require.NoError(t, err, "no credentials should have generated an empty credential set")
	require.NotNil(t, cs, "empty credentials should still return an empty credential set")
}

func TestNoName(t *testing.T) {
	opts := GenerateOptions{
		Silent: true,
	}
	_, err := GenerateCredentials(opts, survey.WithStdio(os.Stdin, os.Stdout, os.Stderr))
	require.Error(t, err, "expected an error because name is required")
}
