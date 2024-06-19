package cnabprovider

import (
	"context"
	"testing"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/secrets"
	"get.porter.sh/porter/pkg/storage"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-go/secrets/host"
	"github.com/cnabio/cnab-go/valuesource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRuntime_loadCredentials(t *testing.T) {
	t.Parallel()

	r := NewTestRuntime(t)
	defer r.Close()

	r.TestCredentials.AddSecret("password", "mypassword")
	r.TestCredentials.AddSecret("db-password", "topsecret")

	b := cnab.NewBundle(bundle.Bundle{
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
	})

	run := storage.Run{
		Action: cnab.ActionInstall,
		Credentials: storage.NewInternalCredentialSet(secrets.SourceMap{
			Name: "password",
			Source: secrets.Source{
				Strategy: secrets.SourceSecret,
				Hint:     "password",
			},
		}),
	}
	err := r.loadCredentials(context.Background(), b, &run)
	require.NoError(t, err, "loadCredentials failed")
	require.Equal(t, "sha256:2d6d3c91ef272afeef2bb29b2fb4b1670c756c623195e71916b8ee138fba60cb",
		run.CredentialsDigest, "expected loadCredentials to set the digest of resolved credentials")
	require.NotEmpty(t, run.Credentials.Credentials[0].ResolvedValue, "expected loadCredentials to set the resolved value of the credentials on the Run")

	gotValues := run.Credentials.ToCNAB()
	wantValues := valuesource.Set{
		"password": "mypassword",
	}
	assert.Equal(t, wantValues, gotValues, "resolved unexpected credential values")
}

func TestRuntime_loadCredentials_WithApplyTo(t *testing.T) {
	getBundle := func(required bool) cnab.ExtendedBundle {
		return cnab.ExtendedBundle{Bundle: bundle.Bundle{
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

		run := &storage.Run{Action: "status"}
		b := getBundle(true)
		err := r.loadCredentials(context.Background(), b, run)
		require.NoError(t, err, "loadCredentials failed")

		gotValues := run.Credentials.ToCNAB()
		assert.Empty(t, gotValues)
	})

	t.Run("optional credential missing", func(t *testing.T) {
		t.Parallel()
		r := NewTestRuntime(t)
		defer r.Close()

		run := &storage.Run{Action: "install"}
		b := getBundle(false)
		err := r.loadCredentials(context.Background(), b, run)
		require.NoError(t, err, "loadCredentials failed")

		gotValues := run.Credentials.ToCNAB()
		assert.Empty(t, gotValues)
	})

	t.Run("required credential missing", func(t *testing.T) {
		t.Parallel()
		r := NewTestRuntime(t)
		defer r.Close()

		run := &storage.Run{Action: "install"}
		b := getBundle(true)
		err := r.loadCredentials(context.Background(), b, run)
		require.Error(t, err, "expected the credential to be required")
	})

	t.Run("credential resolved", func(t *testing.T) {
		t.Parallel()
		r := NewTestRuntime(t)
		defer r.Close()

		r.TestCredentials.AddSecret("password", "mypassword")

		b := getBundle(true)
		run := &storage.Run{
			Action:         cnab.ActionInstall,
			CredentialSets: []string{"mycreds"},
			Credentials: storage.NewInternalCredentialSet(secrets.SourceMap{
				Name: "password",
				Source: secrets.Source{
					Strategy: secrets.SourceSecret,
					Hint:     "password",
				},
			}),
		}
		err := r.loadCredentials(context.Background(), b, run)
		require.NoError(t, err, "loadCredentials failed")
		require.Equal(t, "sha256:2d6d3c91ef272afeef2bb29b2fb4b1670c756c623195e71916b8ee138fba60cb",
			run.CredentialsDigest, "expected loadCredentials to set the digest of resolved credentials")
		require.NotEmpty(t, run.Credentials.Credentials[0].ResolvedValue, "expected loadCredentials to set the resolved value of the credentials on the Run")

		gotValues := run.Credentials.ToCNAB()
		assert.Equal(t, valuesource.Set{"password": "mypassword"}, gotValues, "incorrect resolved credentials")
	})

}

func TestRuntime_nonSecretValue_loadCredentials(t *testing.T) {
	t.Parallel()

	r := NewTestRuntime(t)
	defer r.Close()

	b := cnab.NewBundle(bundle.Bundle{
		Credentials: map[string]bundle.Credential{
			"password": {
				Location: bundle.Location{
					EnvironmentVariable: "PASSWORD",
				},
			},
		},
	})

	run := storage.Run{
		Action: cnab.ActionInstall,
		Credentials: storage.NewInternalCredentialSet(secrets.SourceMap{
			Name: "password",
			Source: secrets.Source{
				Strategy: host.SourceValue,
				Hint:     "mypassword",
			},
		}),
	}
	err := r.loadCredentials(context.Background(), b, &run)
	require.NoError(t, err, "loadCredentials failed")
	require.Equal(t, "sha256:9b6063069a6d911421cf53b30b91836b70957c30eddc70a760eff4765b8cede5",
		run.CredentialsDigest, "expected loadCredentials to set the digest of resolved credentials")
	require.NotEmpty(t, run.Credentials.Credentials[0].ResolvedValue, "expected loadCredentials to set the resolved value of the credentials on the Run")

	gotValues := run.Credentials.ToCNAB()
	wantValues := valuesource.Set{
		"password": "mypassword",
	}
	assert.Equal(t, wantValues, gotValues, "resolved unexpected credential values")
}
