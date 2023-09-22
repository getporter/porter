package storage

import (
	"context"
	"testing"

	"get.porter.sh/porter/pkg/secrets"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCredentialStorage_CRUD(t *testing.T) {
	cs := NewCredentialSet("dev", "sekrets", secrets.SourceMap{
		Name: "password", Source: secrets.Source{
			Strategy: "secret",
			Hint:     "dbPassword"}})

	cp := NewTestCredentialProvider(t)
	defer cp.Close()

	require.NoError(t, cp.InsertCredentialSet(context.Background(), cs))

	creds, err := cp.ListCredentialSets(context.Background(), ListOptions{Namespace: "dev"})
	require.NoError(t, err)
	require.Len(t, creds, 1, "expected 1 credential set")
	require.Equal(t, cs.Name, creds[0].Name, "expected to retrieve sekrets credentials")
	require.Equal(t, cs.Namespace, creds[0].Namespace, "expected to retrieve sekrets credentials")

	creds, err = cp.ListCredentialSets(context.Background(), ListOptions{})
	require.NoError(t, err)
	require.Len(t, creds, 0, "expected no global credential sets")

	creds, err = cp.ListCredentialSets(context.Background(), ListOptions{Namespace: "*"})
	require.NoError(t, err)
	require.Len(t, creds, 1, "expected 1 credential set defined in all namespaces")

	cs.Credentials = append(cs.Credentials, secrets.SourceMap{
		Name: "token", Source: secrets.Source{
			Strategy: "secret",
			Hint:     "github-token",
		},
	})
	require.NoError(t, cp.UpdateCredentialSet(context.Background(), cs))
	cs, err = cp.GetCredentialSet(context.Background(), cs.Namespace, cs.Name)
	require.NoError(t, err)
	assert.Len(t, cs.Credentials, 2)

	cs2 := NewCredentialSet("dev", "sekrets-2", secrets.SourceMap{
		Name: "password-2", Source: secrets.Source{
			Strategy: "secret-2",
			Hint:     "dbPassword-2"}})
	require.NoError(t, cp.InsertCredentialSet(context.Background(), cs2))

	creds, err = cp.ListCredentialSets(context.Background(), ListOptions{Namespace: "dev", Skip: 1})
	require.NoError(t, err)
	require.Len(t, creds, 1, "expected 1 credential set")
	require.Equal(t, cs2.Name, creds[0].Name, "expected to retrieve sekrets-2 credentials")
	require.Equal(t, cs2.Namespace, creds[0].Namespace, "expected to retrieve sekrets-2 credentials")

	creds, err = cp.ListCredentialSets(context.Background(), ListOptions{Namespace: "dev", Limit: 1})
	require.NoError(t, err)
	require.Len(t, creds, 1, "expected 1 credential set")
	require.Equal(t, cs.Name, creds[0].Name, "expected to retrieve sekrets credentials")
	require.Equal(t, cs.Namespace, creds[0].Namespace, "expected to retrieve sekrets credentials")

	require.NoError(t, cp.RemoveCredentialSet(context.Background(), cs.Namespace, cs.Name))
	require.NoError(t, cp.RemoveCredentialSet(context.Background(), cs2.Namespace, cs2.Name))
	_, err = cp.GetCredentialSet(context.Background(), cs.Namespace, cs.Name)
	require.ErrorIs(t, err, ErrNotFound{})
	_, err = cp.GetCredentialSet(context.Background(), cs2.Namespace, cs2.Name)
	require.ErrorIs(t, err, ErrNotFound{})
}

func TestCredentialStorage_Validate_GoodSources(t *testing.T) {
	s := CredentialStore{}
	testCreds := NewCredentialSet("dev", "mycreds",
		secrets.SourceMap{
			Source: secrets.Source{
				Strategy: "env",
				Hint:     "SOME_ENV",
			},
		},
		secrets.SourceMap{
			Source: secrets.Source{
				Strategy: "value",
				Hint:     "somevalue",
			},
		})

	err := s.Validate(context.Background(), testCreds)
	require.NoError(t, err, "Validate did not return errors")
}

func TestCredentialStorage_Validate_BadSources(t *testing.T) {
	s := CredentialStore{}
	testCreds := NewCredentialSet("dev", "mycreds",
		secrets.SourceMap{
			Source: secrets.Source{
				Strategy: "wrongthing",
				Hint:     "SOME_ENV",
			},
		},
		secrets.SourceMap{
			Source: secrets.Source{
				Strategy: "anotherwrongthing",
				Hint:     "somevalue",
			},
		},
	)

	err := s.Validate(context.Background(), testCreds)
	require.Error(t, err, "Validate returned errors")
}
