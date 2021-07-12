package generator

import (
	"fmt"

	"get.porter.sh/porter/pkg/secrets"
	"github.com/cnabio/cnab-go/secrets/host"
	"github.com/cnabio/cnab-go/valuesource"
	survey "gopkg.in/AlecAivazis/survey.v1"
)

// GenerateOptions are the options to generate a parameter or credential set
type GenerateOptions struct {
	// Name of the parameter or credential set.
	Name string

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

type generator func(name string, surveyType SurveyType, defaultVal interface{}) (valuesource.Strategy, error)

func genEmptySet(name string, surveyType SurveyType, defaultVal interface{}) (valuesource.Strategy, error) {
	return valuesource.Strategy{
		Name:   name,
		Source: valuesource.Source{Value: "TODO"},
	}, nil
}

func genSurvey(name string, surveyType SurveyType, defaultVal interface{}) (valuesource.Strategy, error) {
	if surveyType != surveyCredentials && surveyType != surveyParameters {
		return valuesource.Strategy{}, fmt.Errorf("unsupported survey type: %s", surveyType)
	}

	options := []string{questionSecret, questionValue, questionEnvVar, questionPath, questionCommand}
	questionDefault := fmt.Sprintf("use default value (%s)", defaultVal)

	if defaultVal != nil {
		options = append(options, questionDefault)
	}

	// extra space-suffix to align question and answer. Unfortunately misaligns help text
	sourceTypePrompt := &survey.Select{
		Message: fmt.Sprintf("How would you like to set %s %q\n ", surveyType, name),
		Options: options,
		Default: "environment variable",
	}

	// extra space-suffix to align question and answer. Unfortunately misaligns help text
	sourceValuePromptTemplate := "Enter the %s that will be used to set %s %q\n "

	c := valuesource.Strategy{Name: name}

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

	value := ""
	if source != questionDefault {
		sourceValuePrompt := &survey.Input{
			Message: promptMsg,
		}

		if err := survey.AskOne(sourceValuePrompt, &value, nil); err != nil {
			return c, err
		}
	} else {
		return valuesource.Strategy{}, nil
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
