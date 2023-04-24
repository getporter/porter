package generator

import (
	"testing"

	"get.porter.sh/porter/pkg/secrets"
	"github.com/stretchr/testify/require"
)

func Test_genEmptySet(t *testing.T) {
	expected := secrets.SourceMap{
		Name:   "emptyset",
		Source: secrets.Source{Hint: "TODO"},
	}

	got, err := genEmptySet("emptyset", surveyParameters)
	require.NoError(t, err)
	require.Equal(t, expected, got)
}

func Test_genSurvey_unsupported(t *testing.T) {
	got, err := genSurvey("myturtleset", SurveyType("turtles"))
	require.EqualError(t, err, "unsupported survey type: turtles")
	require.Equal(t, secrets.SourceMap{}, got)
}
