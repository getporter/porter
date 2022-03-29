package parameters

import (
	"testing"
	"time"

	"get.porter.sh/porter/pkg/secrets"
	"get.porter.sh/porter/pkg/storage"
	"github.com/stretchr/testify/require"
)

func TestParameterStore_CRUD(t *testing.T) {
	paramStore := NewTestParameterProvider(t)
	defer paramStore.Teardown()

	params, err := paramStore.ListParameterSets("dev", "", nil)
	require.NoError(t, err)
	require.Empty(t, params, "Find should return no entries")

	myParamSet := NewParameterSet("dev", "myparams",
		secrets.Strategy{
			Name: "myparam",
			Source: secrets.Source{
				Key:   "value",
				Value: "myparamvalue",
			},
		})
	myParamSet.Status.Created = time.Date(2020, 1, 1, 12, 0, 0, 0, time.UTC)
	myParamSet.Status.Modified = myParamSet.Status.Created

	err = paramStore.InsertParameterSet(myParamSet)
	require.NoError(t, err, "Insert should successfully save")

	params, err = paramStore.ListParameterSets("dev", "", nil)
	require.NoError(t, err)
	require.Len(t, params, 1, "expected 1 parameter set")
	require.Equal(t, myParamSet.Name, params[0].Name, "expected to retrieve myparams")

	params, err = paramStore.ListParameterSets("", "", nil)
	require.NoError(t, err)
	require.Len(t, params, 0, "expected no global parameter sets")

	params, err = paramStore.ListParameterSets("*", "", nil)
	require.NoError(t, err)
	require.Len(t, params, 1, "expected 1 parameter set defined in all namespaces")

	pset, err := paramStore.GetParameterSet(myParamSet.Namespace, myParamSet.Name)
	require.NoError(t, err)
	require.Equal(t, myParamSet, pset, "Get should return the saved parameter set")

	err = paramStore.RemoveParameterSet(myParamSet.Namespace, myParamSet.Name)
	require.NoError(t, err, "Remove should successfully delete the parameter set")

	params, err = paramStore.ListParameterSets("dev", "", nil)
	require.NoError(t, err)
	require.Empty(t, params, "List should return no entries")

	pset, err = paramStore.GetParameterSet("", myParamSet.Name)
	require.ErrorIs(t, err, storage.ErrNotFound{})

}

func TestParameterStorage_ResolveAll(t *testing.T) {
	// The inmemory secret store currently only supports secret sources
	// So all of these have this same source
	testParameterSet := NewParameterSet("", "myparamset",
		secrets.Strategy{
			Name: "param1",
			Source: secrets.Source{
				Key:   "secret",
				Value: "param1",
			},
		},
		secrets.Strategy{
			Name: "param2",
			Source: secrets.Source{
				Key:   "secret",
				Value: "param2",
			},
		})

	t.Run("resolve params success", func(t *testing.T) {
		paramStore := NewTestParameterProvider(t)
		defer paramStore.Teardown()

		paramStore.AddSecret("param1", "param1_value")
		paramStore.AddSecret("param2", "param2_value")

		expected := []secrets.Strategy{
			secrets.Strategy{
				Name: "param1",
				Source: secrets.Source{
					Key:   "secret",
					Value: "param1",
				},
				Value: "param1_value",
			},
			secrets.Strategy{
				Name: "param2",
				Source: secrets.Source{
					Key:   "secret",
					Value: "param2",
				},
				Value: "param2_value",
			},
		}

		resolved, err := paramStore.ResolveAll(testParameterSet)
		require.NoError(t, err)
		for _, res := range resolved {
			require.Contains(t, expected, res)
		}
	})

	t.Run("resolve params failure", func(t *testing.T) {
		paramStore := NewTestParameterProvider(t)
		defer paramStore.Teardown()

		// Purposefully only adding one secret
		paramStore.AddSecret("param1", "param1_value")

		expected := []secrets.Strategy{
			secrets.Strategy{
				Name: "param1",
				Source: secrets.Source{
					Key:   "secret",
					Value: "param1",
				},
				Value: "param1_value",
			},
			secrets.Strategy{
				Name: "param2",
				Source: secrets.Source{
					Key:   "secret",
					Value: "param2",
				},
			},
		}

		resolved, err := paramStore.ResolveAll(testParameterSet)
		require.EqualError(t, err, "1 error occurred:\n\t* unable to resolve parameter myparamset.param2 from secret param2: secret not found\n\n")
		for _, res := range resolved {
			require.Contains(t, expected, res)
		}
	})
}

func TestParameterStorage_Validate(t *testing.T) {
	t.Run("valid sources", func(t *testing.T) {
		s := ParameterStore{}

		testParameterSet := NewParameterSet("", "myparams",
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
			},
			secrets.Strategy{
				Source: secrets.Source{
					Key:   "path",
					Value: "/some/path",
				},
			},
			secrets.Strategy{
				Source: secrets.Source{
					Key:   "command",
					Value: "some command",
				},
			},
			secrets.Strategy{
				Source: secrets.Source{
					Key:   "secret",
					Value: "secret",
				},
			})

		err := s.Validate(testParameterSet)
		require.NoError(t, err, "Validate did not return errors")
	})

	t.Run("invalid sources", func(t *testing.T) {
		s := ParameterStore{}
		testParameterSet := NewParameterSet("", "myparams",
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
			})

		err := s.Validate(testParameterSet)
		require.Error(t, err, "Validate returned errors")
	})
}
