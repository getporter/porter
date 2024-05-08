package generator

import (
	"os"
	"testing"

	"get.porter.sh/porter/pkg/secrets"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_genEmptySet(t *testing.T) {
	expected := secrets.SourceMap{
		Name:   "emptyset",
		Source: secrets.Source{Hint: "TODO"},
	}

	got, err := genEmptySet("emptyset", surveyParameters, false)
	require.NoError(t, err)
	require.Equal(t, expected, got)
}

func Test_genSurvey_unsupported(t *testing.T) {
	got, err := genSurvey("myturtleset", SurveyType("turtles"), false)
	require.EqualError(t, err, "unsupported survey type: turtles")
	require.Equal(t, secrets.SourceMap{}, got)
}

func TestCheckUserHomeDir(t *testing.T) {
	home, err := os.UserHomeDir()
	assert.NoError(t, err)
	tests := map[string]struct {
		val           string
		expectedValue string
	}{
		"home directory":     {val: "~/.kube/config", expectedValue: home + "/.kube/config"},
		"non-home directory": {val: "tmp", expectedValue: "tmp"},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			newVal, err := checkUserHomeDir(test.val)
			assert.NoError(t, err)
			assert.Equal(t, test.expectedValue, newVal)
		})
	}
}
