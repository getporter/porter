package credentials

import (
	"testing"

	"github.com/cnabio/cnab-go/credentials"
	"github.com/stretchr/testify/require"
)

func TestCredentialStorage_Validate_GoodSources(t *testing.T) {
	s := CredentialStorage{}
	testCreds := credentials.CredentialSet{
		Credentials: []credentials.CredentialStrategy{
			{
				Source: credentials.Source{
					Key:   "env",
					Value: "SOME_ENV",
				},
			},
			{
				Source: credentials.Source{
					Key:   "value",
					Value: "somevalue",
				},
			},
		},
	}

	err := s.Validate(testCreds)
	require.NoError(t, err, "Validate did not return errors")
}

func TestCredentialStorage_Validate_BadSources(t *testing.T) {
	s := CredentialStorage{}
	testCreds := credentials.CredentialSet{
		Credentials: []credentials.CredentialStrategy{
			{
				Source: credentials.Source{
					Key:   "wrongthing",
					Value: "SOME_ENV",
				},
			},
			{
				Source: credentials.Source{
					Key:   "anotherwrongthing",
					Value: "somevalue",
				},
			},
		},
	}

	err := s.Validate(testCreds)
	require.Error(t, err, "Validate returned errors")
}
