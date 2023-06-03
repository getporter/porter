package generator

import (
	"fmt"

	"get.porter.sh/porter/pkg/secrets"
	"github.com/cnabio/cnab-go/secrets/host"
	"gopkg.in/AlecAivazis/survey.v1"
)

// GenerateOptions are the options to generate a parameter or credential set
type GenerateOptions struct {
	// Name of the parameter or credential set.
	Name      string
	Namespace string
	Labels    map[string]string

	// Should we survey?
	Silent bool
}

// SurveyType indicates whether the survey is for a parameter or credential
type SurveyType string

const (
	surveyCredentials SurveyType = "credential"
	surveyParameters  SurveyType = "parameter"

	questionSecret  = "secret"
	questionValue   = "specific value"
	questionEnvVar  = "environment variable"
	questionPath    = "file path"
	questionCommand = "shell command"
)

type generator func(name string, surveyType SurveyType) (secrets.Source, error)

func genEmptySet(name string, surveyType SurveyType) (secrets.Source, error) {
	return secrets.Source{Hint: "TODO"}, nil
}

func genSurvey(name string, surveyType SurveyType) (secrets.Source, error) {
	if surveyType != surveyCredentials && surveyType != surveyParameters {
		return secrets.Source{}, fmt.Errorf("unsupported survey type: %s", surveyType)
	}

	// extra space-suffix to align question and answer. Unfortunately misaligns help text
	sourceTypePrompt := &survey.Select{
		Message: fmt.Sprintf("How would you like to set %s %q\n ", surveyType, name),
		Options: []string{questionSecret, questionValue, questionEnvVar, questionPath, questionCommand},
		Default: "environment variable",
	}

	// extra space-suffix to align question and answer. Unfortunately misaligns help text
	sourceValuePromptTemplate := "Enter the %s that will be used to set %s %q\n "

	source := ""
	if err := survey.AskOne(sourceTypePrompt, &source, nil); err != nil {
		return secrets.Source{}, err
	}

	promptMsg := ""
	switch source {
	case questionSecret:
		promptMsg = fmt.Sprintf(sourceValuePromptTemplate, "secret", surveyType, name)
	case questionValue:
		promptMsg = fmt.Sprintf(sourceValuePromptTemplate, "value", surveyType, name)
	case questionEnvVar:
		promptMsg = fmt.Sprintf(sourceValuePromptTemplate, "environment variable", surveyType, name)
	case questionPath:
		promptMsg = fmt.Sprintf(sourceValuePromptTemplate, "path", surveyType, name)
	case questionCommand:
		promptMsg = fmt.Sprintf(sourceValuePromptTemplate, "command", surveyType, name)
	}

	sourceValuePrompt := &survey.Input{
		Message: promptMsg,
	}

	value := ""
	if err := survey.AskOne(sourceValuePrompt, &value, nil); err != nil {
		return secrets.Source{}, err
	}

	result := secrets.Source{Hint: value}
	switch source {
	case questionSecret:
		result.Strategy = secrets.SourceSecret
	case questionValue:
		result.Strategy = host.SourceValue
	case questionEnvVar:
		result.Strategy = host.SourceEnv
	case questionPath:
		result.Strategy = host.SourcePath
	case questionCommand:
		result.Strategy = host.SourceCommand
	default:
		return secrets.Source{}, fmt.Errorf("unrecogized secret source strategy %q", source)
	}
	return result, nil
}
