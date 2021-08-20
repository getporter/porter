package cnabprovider

import (
	"path/filepath"
	"strings"

	"get.porter.sh/porter/pkg/credentials"
	"get.porter.sh/porter/pkg/encoding"
	"get.porter.sh/porter/pkg/secrets"
	"get.porter.sh/porter/pkg/storage"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/globalsign/mgo/bson"
	"github.com/pkg/errors"
)

func (r *Runtime) loadCredentials(b bundle.Bundle, args ActionArguments) (secrets.Set, error) {
	if len(args.CredentialIdentifiers) == 0 {
		return nil, credentials.Validate(nil, b.Credentials, args.Action)
	}

	// The strategy here is "last one wins". We loop through each credential file and
	// calculate its credentials. Then we insert them into the creds map in the order
	// in which they were supplied on the CLI.
	resolvedCredentials := secrets.Set{}
	for _, name := range args.CredentialIdentifiers {
		var cset credentials.CredentialSet
		var err error
		if r.isPathy(name) {
			cset, err = r.loadCredentialFromFile(name)
		} else {
			// Try to get the creds in the local namespace first, fallback to the global creds
			query := storage.FindOptions{
				Sort: []string{"-namespace"},
				Filter: bson.M{
					"name": name,
					"$or": []bson.M{
						{"namespace": ""},
						{"namespace": args.Installation.Namespace},
					},
				},
			}
			store := r.credentials.GetDataStore()
			err = store.FindOne(credentials.CollectionCredentials, query, &cset)
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
	return resolvedCredentials, credentials.Validate(resolvedCredentials, b.Credentials, args.Action)
}

// isPathy checks to see if a name looks like a path.
func (r *Runtime) isPathy(name string) bool {
	return strings.Contains(name, string(filepath.Separator))
}

func (r *Runtime) loadCredentialFromFile(path string) (credentials.CredentialSet, error) {
	var cs credentials.CredentialSet
	err := encoding.UnmarshalFile(r.FileSystem, path, &cs)
	return cs, errors.Wrapf(err, "error loading credential set in %s", path)
}
