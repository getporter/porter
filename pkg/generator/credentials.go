package generator

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"get.porter.sh/porter/pkg/secrets"
	"get.porter.sh/porter/pkg/storage"
	"github.com/cnabio/cnab-go/bundle"
)

// GenerateCredentialsOptions are the options to generate a Credential Set
type GenerateCredentialsOptions struct {
	GenerateOptions

	// Credentials from the bundle
	Credentials map[string]bundle.Credential
}

// GenerateCredentials will generate a credential set based on the given options
func GenerateCredentials(opts GenerateCredentialsOptions) (storage.CredentialSet, error) {
	if opts.Name == "" {
		return storage.CredentialSet{}, errors.New("credentialset name is required")
	}
	generator := genSurvey
	if opts.Silent {
		generator = genEmptySet
	}
	credSet, err := genCredentialSet(opts.Namespace, opts.Name, opts.Credentials, generator)
	if err != nil {
		return storage.CredentialSet{}, err
	}

	credSet.Labels = opts.Labels
	return credSet, nil
}

func genCredentialSet(namespace string, name string, creds map[string]bundle.Credential, fn generator) (storage.CredentialSet, error) {
	cs := storage.NewCredentialSet(namespace, name)
	cs.Credentials = []secrets.SourceMap{}

	if strings.ContainsAny(name, "./\\") {
		return cs, fmt.Errorf("credentialset name '%s' cannot contain the following characters: './\\'", name)
	}

	var credentialNames []string
	for name := range creds {
		credentialNames = append(credentialNames, name)
	}

	sort.Strings(credentialNames)

	for _, name := range credentialNames {
		c, err := fn(name, surveyCredentials)
		if err != nil {
			return cs, err
		}
		cs.Credentials = append(cs.Credentials, c)
	}

	return cs, nil
}
