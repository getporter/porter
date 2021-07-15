package generator

import (
	"fmt"

	"get.porter.sh/porter/pkg/secrets"
	"github.com/cnabio/cnab-go/secrets/host"
	survey "gopkg.in/AlecAivazis/survey.v1"
)

// GenerateOptions are the options to generate a parameter or credential set
type GenerateOptions struct {
	// Name of the parameter or credential set.
	Name      string
	Namespace string

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

type generator func(name string, surveyType SurveyType) (secrets.Strategy, error)

func genEmptySet(name string, surveyType SurveyType) (secrets.Strategy, error) {
	return secrets.Strategy{
		Name:   name,
		Source: secrets.Source{Value: "TODO"},
	}, nil
}

func genSurvey(name string, surveyType SurveyType) (secrets.Strategy, error) {
	if surveyType != surveyCredentials && surveyType != surveyParameters {
		return secrets.Strategy{}, fmt.Errorf("unsupported survey type: %s", surveyType)
	}

	// extra space-suffix to align question and answer. Unfortunately misaligns help text
	sourceTypePrompt := &survey.Select{
		Message: fmt.Sprintf("How would you like to set %s %q\n ", surveyType, name),
		Options: []string{questionSecret, questionValue, questionEnvVar, questionPath, questionCommand},
		Default: "environment variable",
	}

	// extra space-suffix to align question and answer. Unfortunately misaligns help text
	sourceValuePromptTemplate := "Enter the %s that will be used to set %s %q\n "

	c := secrets.Strategy{Name: name}

	source := ""
	if err := survey.AskOne(sourceTypePrompt, &source, nil); err != nil {
		return c, err
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
		return c, err
	}

	switch source {
	case questionSecret:
		c.Source.Key = secrets.SourceSecret
		c.Source.Value = value
	case questionValue:
		c.Source.Key = host.SourceValue
		c.Source.Value = value
	case questionEnvVar:
		c.Source.Key = host.SourceEnv
		c.Source.Value = value
	case questionPath:
		c.Source.Key = host.SourcePath
		c.Source.Value = value
	case questionCommand:
		c.Source.Key = host.SourceCommand
		c.Source.Value = value
	}
	return c, nil
}
