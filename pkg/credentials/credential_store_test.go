package credentials

import (
	"testing"

	"get.porter.sh/porter/pkg/secrets"
	"get.porter.sh/porter/pkg/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCredentialStorage_CRUD(t *testing.T) {
	cs := NewCredentialSet("dev", "sekrets", secrets.Strategy{
		Name: "password", Source: secrets.Source{
			Key:   "secret",
			Value: "dbPassword"}})

	cp := NewTestCredentialProvider(t)
	defer cp.Teardown()

	require.NoError(t, cp.InsertCredentialSet(cs))

	creds, err := cp.ListCredentialSets("dev")
	require.NoError(t, err)
	require.Len(t, creds, 1, "expected 1 credential set")
	require.Equal(t, cs.Name, creds[0].Name, "expected to retrieve secreks credentials")
	require.Equal(t, cs.Namespace, creds[0].Namespace, "expected to retrieve secreks credentials")

	creds, err = cp.ListCredentialSets("")
	require.NoError(t, err)
	require.Len(t, creds, 0, "expected no global credential sets")

	creds, err = cp.ListCredentialSets("*")
	require.NoError(t, err)
	require.Len(t, creds, 1, "expected 1 credential set defined in all namespaces")

	cs.Credentials = append(cs.Credentials, secrets.Strategy{
		Name: "token", Source: secrets.Source{
			Key:   "secret",
			Value: "github-token",
		},
	})
	require.NoError(t, cp.UpdateCredentialSet(cs))
	cs, err = cp.GetCredentialSet(cs.Namespace, cs.Name)
	require.NoError(t, err)
	assert.Len(t, cs.Credentials, 2)

	require.NoError(t, cp.RemoveCredentialSet(cs.Namespace, cs.Name))
	_, err = cp.GetCredentialSet(cs.Namespace, cs.Name)
	require.ErrorIs(t, err, storage.ErrNotFound{})
}

func TestCredentialStorage_Validate_GoodSources(t *testing.T) {
	s := CredentialStore{}
	testCreds := NewCredentialSet("dev", "mycreds",
		secrets.Strategy{
			Source: secrets.Source{
				Key:   "env",
				Value: "SOME_ENV",
			},
		},
		secrets.Strategy{
			Source: secrets.Source{
				Key:   "value",
				Value: "somevalue",
			},
		})

	err := s.Validate(testCreds)
	require.NoError(t, err, "Validate did not return errors")
}

func TestCredentialStorage_Validate_BadSources(t *testing.T) {
	s := CredentialStore{}
	testCreds := NewCredentialSet("dev", "mycreds",
		secrets.Strategy{
			Source: secrets.Source{
				Key:   "wrongthing",
				Value: "SOME_ENV",
			},
		},
		secrets.Strategy{
			Source: secrets.Source{
				Key:   "anotherwrongthing",
				Value: "somevalue",
			},
		},
	)

	err := s.Validate(testCreds)
	require.Error(t, err, "Validate returned errors")
}
