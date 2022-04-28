package cnabprovider

import (
	"context"
	"testing"

	"get.porter.sh/porter/pkg/claims"
	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/credentials"
	"get.porter.sh/porter/pkg/secrets"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRuntime_loadCredentials(t *testing.T) {
	t.Parallel()

	r := NewTestRuntime(t)
	defer r.Close()

	r.TestCredentials.AddSecret("password", "mypassword")
	r.TestCredentials.AddSecret("db-password", "topsecret")

	r.TestConfig.TestContext.AddTestFile("testdata/db-creds.json", "/db-creds.json")

	cs1 := credentials.NewCredentialSet("", "mycreds", secrets.Strategy{
		Name: "password",
		Source: secrets.Source{
			Key:   secrets.SourceSecret,
			Value: "password",
		},
	})

	err := r.credentials.InsertCredentialSet(context.Background(), cs1)
	require.NoError(t, err, "Save credential set failed")

	b := cnab.ExtendedBundle{bundle.Bundle{
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
	}}

	args := ActionArguments{Installation: claims.Installation{CredentialSets: []string{"mycreds"}}, Action: "install"}
	gotValues, err := r.loadCredentials(context.Background(), b, args)
	require.NoError(t, err, "loadCredentials failed")

	wantValues := secrets.Set{
		"password": "mypassword",
	}
	assert.Equal(t, wantValues, gotValues, "resolved unexpected credential values")

	args = ActionArguments{Installation: claims.Installation{CredentialSets: []string{"/db-creds.json"}}, Action: "install"}
	_, err = r.loadCredentials(context.Background(), b, args)
	require.Error(t, err, "loadCredentials should not load from a file")
}

func TestRuntime_loadCredentials_WithApplyTo(t *testing.T) {
	getBundle := func(required bool) cnab.ExtendedBundle {
		return cnab.ExtendedBundle{bundle.Bundle{
			Credentials: map[string]bundle.Credential{
				"password": {
					Location: bundle.Location{
						EnvironmentVariable: "PASSWORD",
					},
					Required: required,
					ApplyTo:  []string{"install", "upgrade", "uninstall"},
				},
			},
		},
		}
	}

	t.Run("missing required credential does not apply", func(t *testing.T) {
		t.Parallel()
		r := NewTestRuntime(t)
		defer r.Close()

		args := ActionArguments{Action: "status"}
		b := getBundle(true)
		gotValues, err := r.loadCredentials(context.Background(), b, args)
		require.NoError(t, err, "loadCredentials failed")

		var wantValues secrets.Set
		assert.Equal(t, wantValues, gotValues)
	})

	t.Run("optional credential missing", func(t *testing.T) {
		t.Parallel()
		r := NewTestRuntime(t)
		defer r.Close()

		args := ActionArguments{Action: "install"}
		b := getBundle(false)
		gotValues, err := r.loadCredentials(context.Background(), b, args)
		require.NoError(t, err, "loadCredentials failed")

		var wantValues secrets.Set
		assert.Equal(t, wantValues, gotValues)
	})

	t.Run("required credential missing", func(t *testing.T) {
		t.Parallel()
		r := NewTestRuntime(t)
		defer r.Close()

		args := ActionArguments{Action: "install"}
		b := getBundle(true)
		_, err := r.loadCredentials(context.Background(), b, args)
		require.Error(t, err, "expected the credential to be required")
	})

	t.Run("credential resolved", func(t *testing.T) {
		t.Parallel()
		r := NewTestRuntime(t)
		defer r.Close()

		r.TestCredentials.AddSecret("password", "mypassword")

		cs1 := credentials.NewCredentialSet("", "mycreds", secrets.Strategy{
			Name: "password",
			Source: secrets.Source{
				Key:   secrets.SourceSecret,
				Value: "password",
			},
		})

		err := r.credentials.InsertCredentialSet(context.Background(), cs1)
		require.NoError(t, err, "Save credential set failed")

		b := getBundle(true)
		args := ActionArguments{Installation: claims.Installation{CredentialSets: []string{"mycreds"}}, Action: "install"}
		gotValues, err := r.loadCredentials(context.Background(), b, args)
		require.NoError(t, err, "loadCredentials failed")
		assert.Equal(t, secrets.Set{"password": "mypassword"}, gotValues)
	})

}
