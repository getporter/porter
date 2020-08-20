package parameters

import (
	"testing"

	"github.com/cnabio/cnab-go/utils/crud"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/secrets"
	inmemorysecrets "get.porter.sh/porter/pkg/secrets/in-memory"
	"github.com/cnabio/cnab-go/valuesource"
	"github.com/stretchr/testify/require"
)

func TestParameterStorage_ResolveAll(t *testing.T) {
	// The inmemory secret store currently only supports secret sources
	// So all of these have this same source
	testParameterSet := NewParameterSet("myparamset",
		valuesource.Strategy{
			Name: "param1",
			Source: valuesource.Source{
				Key:   "secret",
				Value: "param1",
			},
		},
		valuesource.Strategy{
			Name: "param2",
			Source: valuesource.Source{
				Key:   "secret",
				Value: "param2",
			},
		})

	t.Run("resolve params success", func(t *testing.T) {
		tc := config.NewTestConfig(t)
		backingSecrets := inmemorysecrets.NewStore()
		backingParams := crud.NewMockStore()
		paramStore := NewParameterStore(backingParams)
		secretStore := secrets.NewSecretStore(backingSecrets)

		parameterStorage := ParameterStorage{
			Config:          tc.Config,
			ParametersStore: paramStore,
			SecretsStore:    secretStore,
		}

		backingSecrets.AddSecret("param1", "param1_value")
		backingSecrets.AddSecret("param2", "param2_value")

		expected := valuesource.Set{
			"param1": "param1_value",
			"param2": "param2_value",
		}

		resolved, err := parameterStorage.ResolveAll(testParameterSet)
		require.NoError(t, err)
		require.Equal(t, expected, resolved)
	})

	t.Run("resolve params failure", func(t *testing.T) {
		tc := config.NewTestConfig(t)
		backingSecrets := inmemorysecrets.NewStore()
		backingParams := crud.NewMockStore()
		paramStore := NewParameterStore(backingParams)
		secretStore := secrets.NewSecretStore(backingSecrets)

		parameterStorage := ParameterStorage{
			Config:          tc.Config,
			ParametersStore: paramStore,
			SecretsStore:    secretStore,
		}

		// Purposefully only adding one secret
		backingSecrets.AddSecret("param1", "param1_value")

		expected := valuesource.Set{
			"param1": "param1_value",
			"param2": "",
		}

		resolved, err := parameterStorage.ResolveAll(testParameterSet)
		require.EqualError(t, err, "1 error occurred:\n\t* unable to resolve parameter myparamset.param2 from secret param2: secret not found\n\n")
		require.Equal(t, expected, resolved)
	})
}

func TestParameterStorage_Validate(t *testing.T) {
	t.Run("valid sources", func(t *testing.T) {
		s := ParameterStorage{}

		testParameterSet := NewParameterSet("myparams",
			valuesource.Strategy{
				Source: valuesource.Source{
					Key:   "env",
					Value: "SOME_ENV",
				},
			},
			valuesource.Strategy{
				Source: valuesource.Source{
					Key:   "value",
					Value: "somevalue",
				},
			},
			valuesource.Strategy{
				Source: valuesource.Source{
					Key:   "path",
					Value: "/some/path",
				},
			},
			valuesource.Strategy{
				Source: valuesource.Source{
					Key:   "command",
					Value: "some command",
				},
			},
			valuesource.Strategy{
				Source: valuesource.Source{
					Key:   "secret",
					Value: "secret",
				},
			})

		err := s.Validate(testParameterSet)
		require.NoError(t, err, "Validate did not return errors")
	})

	t.Run("invalid sources", func(t *testing.T) {
		s := ParameterStorage{}
		testParameterSet := NewParameterSet("myparams",
			valuesource.Strategy{
				Source: valuesource.Source{
					Key:   "wrongthing",
					Value: "SOME_ENV",
				},
			},
			valuesource.Strategy{
				Source: valuesource.Source{
					Key:   "anotherwrongthing",
					Value: "somevalue",
				},
			})

		err := s.Validate(testParameterSet)
		require.Error(t, err, "Validate returned errors")
	})
}
