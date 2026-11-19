package generator

import (
	"fmt"
	"os"
	"testing"

	"get.porter.sh/porter/pkg/cnab"
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

func TestBuildSurveySelectDefaultSource(t *testing.T) {
	testcases := []struct {
		name          string
		source        cnab.CredentialDefaultSource
		wantDefault   string
		wantSuggested string
	}{
		{name: "no suggestion", wantDefault: questionEnvVar},
		{
			name:          "environment variable",
			source:        cnab.CredentialDefaultSource{EnvironmentVariable: "GITHUB_TOKEN"},
			wantDefault:   questionEnvVar,
			wantSuggested: "Suggested source: environment variable GITHUB_TOKEN\n",
		},
		{
			name:          "file path",
			source:        cnab.CredentialDefaultSource{Path: "~/.kube/config"},
			wantDefault:   questionPath,
			wantSuggested: "Suggested source: file path ~/.kube/config\n",
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			survey := buildSurveySelect("name", surveyCredentials, withDescription("details"), withDefaultSource(tc.source))
			assert.Equal(t, tc.wantDefault, survey.Default)
			assert.Contains(t, survey.Options, survey.Default)
			wantMessage := fmt.Sprintf(surveryFormatString, surveyPrefix, surveyCredentials, "name", "details\n"+tc.wantSuggested)
			assert.Equal(t, wantMessage, survey.Message)
		})
	}
}

func TestBuildSurveyInputDefaultSource(t *testing.T) {
	suggestPath := withDefaultSource(cnab.CredentialDefaultSource{Path: "~/.kube/config"})

	t.Run("suggested source selected", func(t *testing.T) {
		input := buildSurveyInput("message", questionPath, suggestPath)
		assert.Equal(t, "message", input.Message)
		assert.Equal(t, "~/.kube/config", input.Default)
	})

	t.Run("other source selected", func(t *testing.T) {
		input := buildSurveyInput("message", questionEnvVar, suggestPath)
		assert.Empty(t, input.Default, "the suggested value must not be offered for a different source")
	})

	t.Run("no suggestion", func(t *testing.T) {
		input := buildSurveyInput("message", questionEnvVar)
		assert.Empty(t, input.Default)
	})
}
