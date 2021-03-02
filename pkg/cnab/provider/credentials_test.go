package cnabprovider

import (
	"testing"

	"get.porter.sh/porter/pkg/secrets"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-go/credentials"
	"github.com/cnabio/cnab-go/valuesource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRuntime_loadCredentials(t *testing.T) {
	t.Parallel()

	r := NewTestRuntime(t)

	r.TestCredentials.TestSecrets.AddSecret("password", "mypassword")
	r.TestCredentials.TestSecrets.AddSecret("db-password", "topsecret")

	r.TestConfig.TestContext.AddTestFile("testdata/db-creds.json", "/db-creds.json")

	cs1 := credentials.NewCredentialSet("mycreds",
		valuesource.Strategy{
			Name: "password",
			Source: valuesource.Source{
				Key:   secrets.SourceSecret,
				Value: "password",
			},
		})

	err := r.credentials.Save(cs1)
	require.NoError(t, err, "Save credential set failed")

	b := bundle.Bundle{
		Credentials: map[string]bundle.Credential{
			"password": {
				Location: bundle.Location{
					EnvironmentVariable: "PASSWORD",
				},
			},
			"db-password": {
				Location: bundle.Location{
					EnvironmentVariable: "DB_PASSWORD",
				},
			},
		},
	}

	args := ActionArguments{CredentialIdentifiers: []string{"mycreds", "/db-creds.json"}, Action: "install"}
	gotValues, err := r.loadCredentials(b, args)
	require.NoError(t, err, "loadCredentials failed")

	wantValues := valuesource.Set{
		"password":    "mypassword",
		"db-password": "topsecret",
	}
	assert.Equal(t, wantValues, gotValues, "resolved unexpected credential values")
}

func TestRuntime_loadCredentials_WithApplyTo(t *testing.T) {
	getBundle := func(required bool) bundle.Bundle {
		return bundle.Bundle{
			Credentials: map[string]bundle.Credential{
				"password": {
					Location: bundle.Location{
						EnvironmentVariable: "PASSWORD",
					},
					Required: required,
					ApplyTo:  []string{"install", "upgrade", "uninstall"},
				},
			},
		}
	}

	t.Run("missing required credential does not apply", func(t *testing.T) {
		t.Parallel()
		r := NewTestRuntime(t)

		args := ActionArguments{Action: "status"}
		b := getBundle(true)
		gotValues, err := r.loadCredentials(b, args)
		require.NoError(t, err, "loadCredentials failed")

		var wantValues valuesource.Set
		assert.Equal(t, wantValues, gotValues)
	})

	t.Run("optional credential missing", func(t *testing.T) {
		t.Parallel()
		r := NewTestRuntime(t)

		args := ActionArguments{Action: "install"}
		b := getBundle(false)
		gotValues, err := r.loadCredentials(b, args)
		require.NoError(t, err, "loadCredentials failed")

		var wantValues valuesource.Set
		assert.Equal(t, wantValues, gotValues)
	})

	t.Run("required credential missing", func(t *testing.T) {
		t.Parallel()
		r := NewTestRuntime(t)

		args := ActionArguments{Action: "install"}
		b := getBundle(true)
		_, err := r.loadCredentials(b, args)
		require.Error(t, err, "expected the credential to be required")
	})

	t.Run("credential resolved", func(t *testing.T) {
		t.Parallel()
		r := NewTestRuntime(t)
		r.TestCredentials.TestSecrets.AddSecret("password", "mypassword")

		cs1 := credentials.NewCredentialSet("mycreds",
			valuesource.Strategy{
				Name: "password",
				Source: valuesource.Source{
					Key:   secrets.SourceSecret,
					Value: "password",
				},
			})

		err := r.credentials.Save(cs1)
		require.NoError(t, err, "Save credential set failed")

		b := getBundle(true)
		args := ActionArguments{CredentialIdentifiers: []string{"mycreds"}, Action: "install"}
		gotValues, err := r.loadCredentials(b, args)
		require.NoError(t, err, "loadCredentials failed")
		assert.Equal(t, valuesource.Set{"password": "mypassword"}, gotValues)
	})

}
