package generator

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-go/credentials"
	"github.com/cnabio/cnab-go/valuesource"
)

// GenerateCredentialsOptions are the options to generate a Credential Set
type GenerateCredentialsOptions struct {
	GenerateOptions

	// Credentials from the bundle
	Credentials map[string]bundle.Credential
}

// GenerateCredentials will generate a credential set based on the given options
func GenerateCredentials(opts GenerateCredentialsOptions) (credentials.CredentialSet, error) {
	if opts.Name == "" {
		return credentials.CredentialSet{}, errors.New("credentialset name is required")
	}
	generator := genSurvey
	if opts.Silent {
		generator = genEmptySet
	}
	credSet, err := genCredentialSet(opts.Name, opts.Credentials, generator)
	if err != nil {
		return credentials.CredentialSet{}, err
	}
	return credSet, nil
}

func genCredentialSet(name string, creds map[string]bundle.Credential, fn generator) (credentials.CredentialSet, error) {
	cs := credentials.NewCredentialSet(name)
	cs.Credentials = []valuesource.Strategy{}

	if strings.ContainsAny(name, "./\\") {
		return cs, fmt.Errorf("credentialset name '%s' cannot contain the following characters: './\\'", name)
	}

	var credentialNames []string
	for name := range creds {
		credentialNames = append(credentialNames, name)
	}

	sort.Strings(credentialNames)

	for _, name := range credentialNames {
		c, err := fn(name, surveyCredentials, nil)
		if err != nil {
			return cs, err
		}
		cs.Credentials = append(cs.Credentials, c)
	}

	return cs, nil
}
