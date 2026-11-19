package cnab

import (
	"testing"

	"github.com/cnabio/cnab-go/bundle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCredentialDefaultSource_Validate(t *testing.T) {
	testcases := []struct {
		name    string
		source  CredentialDefaultSource
		wantErr string
	}{
		{name: "env", source: CredentialDefaultSource{EnvironmentVariable: "GITHUB_TOKEN"}},
		{name: "path", source: CredentialDefaultSource{Path: "~/.kube/config"}},
		{name: "empty", wantErr: "exactly one of env or path must be set"},
		{name: "both", source: CredentialDefaultSource{EnvironmentVariable: "GITHUB_TOKEN", Path: "~/.kube/config"}, wantErr: "exactly one of env or path must be set"},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.source.Validate()
			if tc.wantErr != "" {
				require.EqualError(t, err, tc.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestExtendedBundle_ReadCredentialDefaultSources(t *testing.T) {
	t.Run("not defined", func(t *testing.T) {
		sources, err := NewBundle(bundle.Bundle{}).ReadCredentialDefaultSources()
		require.NoError(t, err)
		assert.Nil(t, sources)
	})

	t.Run("loaded from bundle.json", func(t *testing.T) {
		// Unmarshaling a bundle.json results in untyped data in the custom section
		b := NewBundle(bundle.Bundle{Custom: map[string]interface{}{
			CredentialDefaultSourcesExtensionKey: map[string]interface{}{
				"kubeconfig": map[string]interface{}{"path": "~/.kube/config"},
				"token":      map[string]interface{}{"env": "GITHUB_TOKEN"},
			},
		}})

		sources, err := b.ReadCredentialDefaultSources()
		require.NoError(t, err)
		assert.Equal(t, CredentialDefaultSources{
			"kubeconfig": {Path: "~/.kube/config"},
			"token":      {EnvironmentVariable: "GITHUB_TOKEN"},
		}, sources)
	})

	t.Run("wrong type", func(t *testing.T) {
		b := NewBundle(bundle.Bundle{Custom: map[string]interface{}{
			CredentialDefaultSourcesExtensionKey: "kubeconfig",
		}})

		_, err := b.ReadCredentialDefaultSources()
		require.ErrorContains(t, err, "could not unmarshal the \"org.getporter.credential-default-sources\" extension")
	})

	t.Run("unsupported source", func(t *testing.T) {
		// A bundle must not be able to get a command suggested to the user
		b := NewBundle(bundle.Bundle{Custom: map[string]interface{}{
			CredentialDefaultSourcesExtensionKey: map[string]interface{}{
				"token": map[string]interface{}{"command": "cat ~/.ssh/id_rsa"},
			},
		}})

		_, err := b.ReadCredentialDefaultSources()
		require.ErrorContains(t, err, `unsupported credential default source "command", only env or path can be set`)
	})

	t.Run("unsupported source with env", func(t *testing.T) {
		b := NewBundle(bundle.Bundle{Custom: map[string]interface{}{
			CredentialDefaultSourcesExtensionKey: map[string]interface{}{
				"token": map[string]interface{}{"env": "GITHUB_TOKEN", "command": "cat ~/.ssh/id_rsa"},
			},
		}})

		_, err := b.ReadCredentialDefaultSources()
		require.ErrorContains(t, err, `unsupported credential default source "command", only env or path can be set`)
	})

	t.Run("no source", func(t *testing.T) {
		b := NewBundle(bundle.Bundle{Custom: map[string]interface{}{
			CredentialDefaultSourcesExtensionKey: map[string]interface{}{
				"token": map[string]interface{}{},
			},
		}})

		_, err := b.ReadCredentialDefaultSources()
		require.ErrorContains(t, err, `credential "token": exactly one of env or path must be set`)
	})
}
