package credentialsgenerator

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"get.porter.sh/porter/pkg/secrets"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-go/credentials"
	"github.com/cnabio/cnab-go/secrets/host"
	survey "gopkg.in/AlecAivazis/survey.v1"
)

type GenerateOptions struct {

	// Name of the credential.
	Name string

	//Credentials from the bundle
	Credentials map[string]bundle.Credential

	//Should we survey?
	Silent bool
}

type credentialAnswers struct {
	Source string `survey:"source"`
	Value  string `survey:"value"`
}

const (
	questionSecret  = "secret"
	questionValue   = "specific value"
	questionEnvVar  = "environment variable"
	questionPath    = "file path"
	questionCommand = "shell command"
)

type credentialGenerator func(name string) (credentials.CredentialStrategy, error)

// GenerateCredentials will generate a credential set based on the given options
func GenerateCredentials(opts GenerateOptions) (*credentials.CredentialSet, error) {
	if opts.Name == "" {
		return nil, errors.New("credentialset name is required")
	}
	generator := genCredentialSurvey
	if opts.Silent {
		generator = genEmptyCredentials
	}
	credSet, err := genCredentialSet(opts.Name, opts.Credentials, generator)
	if err != nil {
		return nil, err
	}
	return &credSet, nil
}

func genCredentialSet(name string, creds map[string]bundle.Credential, fn credentialGenerator) (credentials.CredentialSet, error) {
	cs := credentials.CredentialSet{
		Name: name,
	}
	cs.Credentials = []credentials.CredentialStrategy{}

	if strings.ContainsAny(name, "./\\") {
		return cs, fmt.Errorf("credentialset name '%s' cannot contain the following characters: './\\'", name)
	}

	var credentialNames []string
	for name := range creds {
		credentialNames = append(credentialNames, name)
	}

	sort.Strings(credentialNames)

	for _, name := range credentialNames {
		c, err := fn(name)
		if err != nil {
			return cs, err
		}
		cs.Credentials = append(cs.Credentials, c)
	}

	return cs, nil
}

func genEmptyCredentials(name string) (credentials.CredentialStrategy, error) {
	return credentials.CredentialStrategy{
		Name:   name,
		Source: credentials.Source{Value: "TODO"},
	}, nil
}

func genCredentialSurvey(name string) (credentials.CredentialStrategy, error) {

	// extra space-suffix to align question and answer
	sourceTypePrompt := &survey.Select{
		Message: fmt.Sprintf("How would you like to set credential %q\n ", name),
		Options: []string{questionSecret, questionValue, questionEnvVar, questionPath, questionCommand},
		Default: "environment variable",
	}

	// extra space-suffix to align question and answer
	sourceValuePromptTemplate := "Enter the %s that will be used to set %q\n "

	c := credentials.CredentialStrategy{Name: name}

	source := ""
	if err := survey.AskOne(sourceTypePrompt, &source, nil); err != nil {
		return c, err
	}

	promptMsg := ""
	switch source {
	case questionSecret:
		promptMsg = fmt.Sprintf(sourceValuePromptTemplate, "secret", name)
	case questionValue:
		promptMsg = fmt.Sprintf(sourceValuePromptTemplate, "value", name)
	case questionEnvVar:
		promptMsg = fmt.Sprintf(sourceValuePromptTemplate, "environment variable", name)
	case questionPath:
		promptMsg = fmt.Sprintf(sourceValuePromptTemplate, "path", name)
	case questionCommand:
		promptMsg = fmt.Sprintf(sourceValuePromptTemplate, "command", name)
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
