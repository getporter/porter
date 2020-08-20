package parameters

import (
	"testing"
	"time"

	"github.com/cnabio/cnab-go/utils/crud"
	"github.com/cnabio/cnab-go/valuesource"
	"github.com/stretchr/testify/require"
)

func TestNewParameterStore(t *testing.T) {
	backingParams := crud.NewBackingStore(crud.NewMockStore())
	paramStore := NewParameterStore(backingParams)

	params, err := paramStore.List()
	require.NoError(t, err)
	require.Empty(t, params, "List should return no entries")

	myParamSet := NewParameterSet("myparams",
		valuesource.Strategy{
			Name: "myparam",
			Source: valuesource.Source{
				Key:   "value",
				Value: "myparamvalue",
			},
		})
	myParamSet.Created = time.Date(2020, 1, 1, 12, 0, 0, 0, time.UTC)
	myParamSet.Modified = myParamSet.Created

	err = paramStore.Save(myParamSet)
	require.NoError(t, err, "Save should successfully save")

	params, err = paramStore.List()
	require.NoError(t, err)
	require.Equal(t, []string{"myparams"}, params, "List should return the saved entry")

	pset, err := paramStore.Read("myparams")
	require.NoError(t, err)
	require.Equal(t, myParamSet, pset, "Read should return the saved parameter set")

	psets, err := paramStore.ReadAll()
	require.NoError(t, err)
	require.Equal(t, []ParameterSet{myParamSet}, psets, "ReadAll should return all parameter sets")

	err = paramStore.Delete("myparams")
	require.NoError(t, err, "Delete should successfully delete the parameter set")

	params, err = paramStore.List()
	require.NoError(t, err)
	require.Empty(t, params, "List should return no entries")

	pset, err = paramStore.Read("myparams")
	require.EqualError(t, err, "Parameter set does not exist", "Read should return an error")
}
