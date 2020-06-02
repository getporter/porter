package credentials

import (
	"testing"

	"github.com/cnabio/cnab-go/credentials"
	"github.com/cnabio/cnab-go/valuesource"
	"github.com/stretchr/testify/require"
)

func TestCredentialStorage_Validate_GoodSources(t *testing.T) {
	s := CredentialStorage{}
	testCreds := credentials.CredentialSet{
		Credentials: []valuesource.Strategy{
			{
				Source: valuesource.Source{
					Key:   "env",
					Value: "SOME_ENV",
				},
			},
			{
				Source: valuesource.Source{
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
		Credentials: []valuesource.Strategy{
			{
				Source: valuesource.Source{
					Key:   "wrongthing",
					Value: "SOME_ENV",
				},
			},
			{
				Source: valuesource.Source{
					Key:   "anotherwrongthing",
					Value: "somevalue",
				},
			},
		},
	}

	err := s.Validate(testCreds)
	require.Error(t, err, "Validate returned errors")
}
