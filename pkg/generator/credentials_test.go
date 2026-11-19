package generator

import (
	"fmt"
	"testing"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/secrets"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/stretchr/testify/assert"
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

func TestGenCredentialSet_DefaultSources(t *testing.T) {
	creds := map[string]bundle.Credential{
		"kubeconfig": {},
		"password":   {},
		"token":      {},
	}
	defaultSources := cnab.CredentialDefaultSources{
		"kubeconfig": {Path: "~/.kube/config"},
		"token":      {EnvironmentVariable: "GITHUB_TOKEN"},
	}

	got := map[string]surveyOptions{}
	spy := func(name string, surveyType SurveyType, opts ...surveyOption) (secrets.SourceMap, error) {
		var options surveyOptions
		for _, opt := range opts {
			opt(&options)
		}
		got[name] = options
		return secrets.SourceMap{Name: name}, nil
	}

	_, err := genCredentialSet("", "mycreds", creds, defaultSources, spy)
	require.NoError(t, err)

	assert.Equal(t, questionPath, got["kubeconfig"].defaultQuestion)
	assert.Equal(t, "~/.kube/config", got["kubeconfig"].defaultValue)
	assert.Equal(t, questionEnvVar, got["token"].defaultQuestion)
	assert.Equal(t, "GITHUB_TOKEN", got["token"].defaultValue)
	assert.Empty(t, got["password"].defaultQuestion, "password does not have a suggested source")
	assert.Empty(t, got["password"].defaultValue, "password does not have a suggested source")
}

func TestGenerateCredentials_SilentIgnoresDefaultSources(t *testing.T) {
	// Without a prompt the user can't confirm the suggestion, so the bundle must not be able to pick the source
	cs, err := GenerateCredentials(GenerateCredentialsOptions{
		GenerateOptions: GenerateOptions{Name: "mycreds", Silent: true},
		Credentials:     map[string]bundle.Credential{"kubeconfig": {}},
		DefaultSources:  cnab.CredentialDefaultSources{"kubeconfig": {Path: "~/.kube/config"}},
	})
	require.NoError(t, err)
	require.Len(t, cs.Credentials, 1)
	assert.Equal(t, secrets.Source{Hint: "TODO"}, cs.Credentials[0].Source)
}
