package generator

import (
	"fmt"
	"testing"

	"github.com/cnabio/cnab-go/bundle"
	"github.com/stretchr/testify/require"
)

func TestBadCredentialsName(t *testing.T) {
	name := "this.hasadot"
	opts := GenerateCredentialsOptions{
		GenerateOptions: GenerateOptions{
			Name:   name,
			Silent: true,
		},
	}

	cs, err := GenerateCredentials(opts)
	require.Error(t, err, "bad name should have resulted in an error")
	require.Empty(t, cs, "credential set should have been empty")
	require.EqualError(t, err, fmt.Sprintf("credentialset name '%s' cannot contain the following characters: './\\'", name))
}

func TestGoodCredentialsName(t *testing.T) {
	name := "this-name-is-valid"
	opts := GenerateCredentialsOptions{
		GenerateOptions: GenerateOptions{
			Name:   name,
			Silent: true,
		},
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

	cs, err := GenerateCredentials(opts)
	require.NoError(t, err, "name should NOT have resulted in an error")
	require.Equal(t, 3, len(cs.Credentials), "should have had 3 entries")
}

func TestNoCredentials(t *testing.T) {
	name := "this-name-is-valid"
	opts := GenerateCredentialsOptions{
		GenerateOptions: GenerateOptions{
			Name:   name,
			Silent: true,
		},
	}
	cs, err := GenerateCredentials(opts)
	require.NoError(t, err, "no credentials should have generated an empty credential set")
	require.NotNil(t, cs, "empty credentials should still return an empty credential set")
}

func TestEmptyCredentials(t *testing.T) {
	name := "this-name-is-valid"
	opts := GenerateCredentialsOptions{
		GenerateOptions: GenerateOptions{
			Name:   name,
			Silent: true,
		},
		Credentials: map[string]bundle.Credential{},
	}
	cs, err := GenerateCredentials(opts)
	require.NoError(t, err, "empty credentials should have generated an empty credential set")
	require.NotEmpty(t, cs, "empty credentials should still return an empty credential set")
}

func TestNoCredentialsName(t *testing.T) {
	opts := GenerateCredentialsOptions{
		GenerateOptions: GenerateOptions{
			Silent: true,
		},
	}
	cs, err := GenerateCredentials(opts)
	require.Error(t, err, "expected an error because name is required")
	require.Empty(t, cs, "credential set should have been empty")
}
