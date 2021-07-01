package generator

import (
	"testing"

	"github.com/cnabio/cnab-go/valuesource"
	"github.com/stretchr/testify/require"
)

func Test_genEmptySet(t *testing.T) {
	expected := valuesource.Strategy{
		Name:   "emptyset",
		Source: valuesource.Source{Value: "TODO"},
	}

	got, err := genEmptySet("emptyset", surveyParameters, nil)
	require.NoError(t, err)
	require.Equal(t, expected, got)
}

func Test_genSurvey_unsupported(t *testing.T) {
	got, err := genSurvey("myturtleset", SurveyType("turtles"), nil)
	require.EqualError(t, err, "unsupported survey type: turtles")
	require.Equal(t, valuesource.Strategy{}, got)
}
