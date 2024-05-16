package generator

import (
	"fmt"
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

	got, err := genEmptySet("emptyset", surveyParameters)
	require.NoError(t, err)
	require.Equal(t, expected, got)
}

func Test_genSurvey_unsupported(t *testing.T) {
	got, err := genSurvey("myturtleset", SurveyType("turtles"))
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

func TestBuildSurveySelectRequiredTrue(t *testing.T) {
	survey := buildSurveySelect("name", surveyCredentials, withRequired(true))
	assert.NotContains(t, survey.Options, questionSkip)
}

func TestBuildSurveySelectRequiredFalse(t *testing.T) {
	survey := buildSurveySelect("name", surveyCredentials, withRequired(false))
	assert.Contains(t, survey.Options, questionSkip)
}

func TestBuildSurveySelectEmptyDescription(t *testing.T) {
	name := "name_value"
	description := ""
	survey := buildSurveySelect(name, surveyCredentials, withDescription(description))
	assert.Equal(t, survey.Message, fmt.Sprintf(surveryFormatString, surveyPrefix, surveyCredentials, name, formatDescriptionForSurvey(description)))
}

func TestBuildSurveySelectValidDescription(t *testing.T) {
	name := "name_value"
	description := "here are details on how to fill out the survey"
	survey := buildSurveySelect(name, surveyCredentials, withDescription(description))
	assert.Equal(t, survey.Message, fmt.Sprintf(surveryFormatString, surveyPrefix, surveyCredentials, name, formatDescriptionForSurvey(description)))
}
