package generator

import (
	"fmt"
	"os"
	"strings"

	"get.porter.sh/porter/pkg/secrets"
	"github.com/cnabio/cnab-go/secrets/host"
	survey "gopkg.in/AlecAivazis/survey.v1"
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

	questionSecret      = "secret"
	questionValue       = "specific value"
	questionEnvVar      = "environment variable"
	questionPath        = "file path"
	questionCommand     = "shell command"
	questionSkip        = "skip"
	surveryFormatString = "%s %s %q\n%s"
	surveyPrefix        = "How would you like to set"
)

type surveyOptions struct {
	required    bool
	description string
}

type surveyOption func(*surveyOptions)

func withDescription(description string) surveyOption {
	return func(s *surveyOptions) {
		s.description = formatDescriptionForSurvey(description)
	}
}

func withRequired(required bool) surveyOption {
	return func(s *surveyOptions) {
		s.required = required
	}
}

type generator func(name string, surveyType SurveyType, opts ...surveyOption) (secrets.SourceMap, error)

func genEmptySet(name string, surveyType SurveyType, opts ...surveyOption) (secrets.SourceMap, error) {
	return secrets.SourceMap{
		Name:   name,
		Source: secrets.Source{Hint: "TODO"},
	}, nil
}

func formatDescriptionForSurvey(description string) string {
	if description != "" {
		description = description + "\n"
	}
	return description
}

func buildSurveySelect(name string, surveyType SurveyType, opts ...surveyOption) *survey.Select {
	surveyOptions := &surveyOptions{}
	for _, opt := range opts {
		opt(surveyOptions)
	}

	selectOptions := []string{questionSecret, questionValue, questionEnvVar, questionPath, questionCommand}
	if !surveyOptions.required {
		selectOptions = append(selectOptions, questionSkip)
	}

	// extra space-suffix to align question and answer. Unfortunately misaligns help text
	return &survey.Select{
		Message: fmt.Sprintf(surveryFormatString, surveyPrefix, surveyType, name, surveyOptions.description),
		Options: selectOptions,
		Default: "environment variable",
	}

}

func genSurvey(name string, surveyType SurveyType, opts ...surveyOption) (secrets.SourceMap, error) {
	if surveyType != surveyCredentials && surveyType != surveyParameters {
		return secrets.SourceMap{}, fmt.Errorf("unsupported survey type: %s", surveyType)
	}

	// extra space-suffix to align question and answer. Unfortunately misaligns help text
	sourceTypePrompt := buildSurveySelect(name, surveyType, opts...)

	// extra space-suffix to align question and answer. Unfortunately misaligns help text
	sourceValuePromptTemplate := "Enter the %s that will be used to set %s %q\n "

	c := secrets.SourceMap{Name: name}

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
	case questionSkip:
		promptMsg = fmt.Sprintf(sourceValuePromptTemplate, "skip", surveyType, name)
	}

	if source == questionSkip {
		return secrets.SourceMap{}, nil
	}

	sourceValuePrompt := &survey.Input{
		Message: promptMsg,
	}

	value := ""
	if err := survey.AskOne(sourceValuePrompt, &value, nil); err != nil {
		return c, err
	}
	value, err := checkUserHomeDir(value)
	if err != nil {
		return c, err
	}
	switch source {
	case questionSecret:
		c.Source.Strategy = secrets.SourceSecret
		c.Source.Hint = value
	case questionValue:
		c.Source.Strategy = host.SourceValue
		c.Source.Hint = value
	case questionEnvVar:
		c.Source.Strategy = host.SourceEnv
		c.Source.Hint = value
	case questionPath:
		c.Source.Strategy = host.SourcePath
		c.Source.Hint = value
	case questionCommand:
		c.Source.Strategy = host.SourceCommand
		c.Source.Hint = value
	}
	return c, nil
}

func checkUserHomeDir(value string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	if strings.HasPrefix(value, "~/") {
		return strings.Replace(value, "~/", home+"/", 1), nil
	}
	return value, nil
}
