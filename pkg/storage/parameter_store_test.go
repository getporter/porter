package storage

import (
	"context"
	"testing"
	"time"

	"get.porter.sh/porter/pkg/secrets"
	"github.com/stretchr/testify/require"
)

func TestParameterStore_CRUD(t *testing.T) {
	paramStore := NewTestParameterProvider(t)
	defer paramStore.Close()

	ctx := context.Background()
	params, err := paramStore.ListParameterSets(ctx, ListOptions{Namespace: "dev"})
	require.NoError(t, err)
	require.Empty(t, params, "Find should return no entries")

	myParamSet := NewParameterSet("dev", "myparams",
		secrets.SourceMap{
			Name: "myparam",
			Source: secrets.Source{
				Strategy: "value",
				Hint:     "myparamvalue",
			},
		})
	myParamSet.Status.Created = time.Date(2020, 1, 1, 12, 0, 0, 0, time.UTC)
	myParamSet.Status.Modified = myParamSet.Status.Created

	err = paramStore.InsertParameterSet(ctx, myParamSet)
	require.NoError(t, err, "Insert should successfully save")

	params, err = paramStore.ListParameterSets(ctx, ListOptions{Namespace: "dev"})
	require.NoError(t, err)
	require.Len(t, params, 1, "expected 1 parameter set")
	require.Equal(t, myParamSet.Name, params[0].Name, "expected to retrieve myparam")

	params, err = paramStore.ListParameterSets(ctx, ListOptions{})
	require.NoError(t, err)
	require.Len(t, params, 0, "expected no global parameter sets")

	params, err = paramStore.ListParameterSets(ctx, ListOptions{Namespace: "*"})
	require.NoError(t, err)
	require.Len(t, params, 1, "expected 1 parameter set defined in all namespaces")

	pset, err := paramStore.GetParameterSet(ctx, myParamSet.Namespace, myParamSet.Name)
	require.NoError(t, err)
	require.Equal(t, myParamSet, pset, "Get should return the saved parameter set")

	myParamSet2 := NewParameterSet("dev", "myparams2",
		secrets.SourceMap{
			Name: "myparam2",
			Source: secrets.Source{
				Strategy: "value2",
				Hint:     "myparamvalue2",
			},
		})
	myParamSet2.Status.Created = time.Date(2020, 1, 1, 12, 0, 0, 0, time.UTC)
	myParamSet2.Status.Modified = myParamSet2.Status.Created

	err = paramStore.InsertParameterSet(ctx, myParamSet2)
	require.NoError(t, err, "Insert should successfully save")

	params, err = paramStore.ListParameterSets(ctx, ListOptions{Namespace: "dev", Skip: 1})
	require.NoError(t, err)
	require.Len(t, params, 1, "expected 1 parameter set")
	require.Equal(t, myParamSet2.Name, params[0].Name, "expected to retrieve myparam2")

	params, err = paramStore.ListParameterSets(ctx, ListOptions{Namespace: "dev", Limit: 1})
	require.NoError(t, err)
	require.Len(t, params, 1, "expected 1 parameter set")
	require.Equal(t, myParamSet.Name, params[0].Name, "expected to retrieve myparam")

	err = paramStore.RemoveParameterSet(ctx, myParamSet.Namespace, myParamSet.Name)
	require.NoError(t, err, "Remove should successfully delete the parameter set")

	err = paramStore.RemoveParameterSet(ctx, myParamSet2.Namespace, myParamSet2.Name)
	require.NoError(t, err, "Remove should successfully delete the parameter set")

	params, err = paramStore.ListParameterSets(ctx, ListOptions{Namespace: "dev"})
	require.NoError(t, err)
	require.Empty(t, params, "List should return no entries")

	_, err = paramStore.GetParameterSet(ctx, "", myParamSet.Name)
	require.ErrorIs(t, err, ErrNotFound{})

	_, err = paramStore.GetParameterSet(ctx, "", myParamSet2.Name)
	require.ErrorIs(t, err, ErrNotFound{})
}

func TestParameterStorage_ResolveAll(t *testing.T) {
	// The inmemory secret store currently only supports secret sources
	// So all of these have this same source
	testParameterSet := NewParameterSet("", "myparamset",
		secrets.SourceMap{
			Name: "param1",
			Source: secrets.Source{
				Strategy: "secret",
				Hint:     "param1",
			},
		},
		secrets.SourceMap{
			Name: "param2",
			Source: secrets.Source{
				Strategy: "secret",
				Hint:     "param2",
			},
		})

	t.Run("resolve params success", func(t *testing.T) {
		paramStore := NewTestParameterProvider(t)
		defer paramStore.Close()

		paramStore.AddSecret("param1", "param1_value")
		paramStore.AddSecret("param2", "param2_value")

		expected := secrets.Set{
			"param1": "param1_value",
			"param2": "param2_value",
		}

		resolved, err := paramStore.ResolveAll(context.Background(), testParameterSet)
		require.NoError(t, err)
		require.Equal(t, expected, resolved)
	})

	t.Run("resolve params failure", func(t *testing.T) {
		paramStore := NewTestParameterProvider(t)
		defer paramStore.Close()

		// Purposefully only adding one secret
		paramStore.AddSecret("param1", "param1_value")

		expected := secrets.Set{
			"param1": "param1_value",
			"param2": "",
		}

		resolved, err := paramStore.ResolveAll(context.Background(), testParameterSet)
		require.EqualError(t, err, "1 error occurred:\n\t* unable to resolve parameter myparamset.param2 from secret param2: secret not found\n\n")
		require.Equal(t, expected, resolved)
	})
}

func TestParameterStorage_Validate(t *testing.T) {
	t.Run("valid sources", func(t *testing.T) {
		s := ParameterStore{}

		testParameterSet := NewParameterSet("", "myparams",
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
			},
			secrets.SourceMap{
				Source: secrets.Source{
					Strategy: "path",
					Hint:     "/some/path",
				},
			},
			secrets.SourceMap{
				Source: secrets.Source{
					Strategy: "command",
					Hint:     "some command",
				},
			},
			secrets.SourceMap{
				Source: secrets.Source{
					Strategy: "secret",
					Hint:     "secret",
				},
			})

		err := s.Validate(context.Background(), testParameterSet)
		require.NoError(t, err, "Validate did not return errors")
	})

	t.Run("invalid sources", func(t *testing.T) {
		s := ParameterStore{}
		testParameterSet := NewParameterSet("", "myparams",
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
			})

		err := s.Validate(context.Background(), testParameterSet)
		require.Error(t, err, "Validate returned errors")
	})
}
