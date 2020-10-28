package cnabprovider

import (
	"encoding/json"
	"path/filepath"
	"strings"

	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-go/credentials"
	"github.com/cnabio/cnab-go/valuesource"
	"github.com/pkg/errors"
)

func (r *Runtime) loadCredentials(b bundle.Bundle, creds []string) (valuesource.Set, error) {
	if len(creds) == 0 {
		return nil, credentials.Validate(nil, b.Credentials)
	}

	// The strategy here is "last one wins". We loop through each credential file and
	// calculate its credentials. Then we insert them into the creds map in the order
	// in which they were supplied on the CLI.
	resolvedCredentials := valuesource.Set{}
	for _, name := range creds {
		var cset credentials.CredentialSet
		var err error
		if r.isPathy(name) {
			cset, err = r.loadCredentialFromFile(name)
		} else {
			cset, err = r.credentials.Read(name)
		}
		if err != nil {
			return nil, err
		}

		rc, err := r.credentials.ResolveAll(cset)
		if err != nil {
			return nil, err
		}

		for k, v := range rc {
			resolvedCredentials[k] = v
		}
	}
	return resolvedCredentials, credentials.Validate(resolvedCredentials, b.Credentials)
}

// isPathy checks to see if a name looks like a path.
func (r *Runtime) isPathy(name string) bool {
	// TODO: export back outta Compton

	return strings.Contains(name, string(filepath.Separator))
}

func (r *Runtime) loadCredentialFromFile(path string) (credentials.CredentialSet, error) {
	data, err := r.FileSystem.ReadFile(path)
	if err != nil {
		return credentials.CredentialSet{}, errors.Wrapf(err, "could not read file %s", path)
	}

	var cs credentials.CredentialSet
	err = json.Unmarshal(data, &cs)
	return cs, errors.Wrapf(err, "error loading credential set in %s", path)
}
