package cnabprovider

import (
	"path/filepath"
	"strings"

	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-go/credentials"
)

const (
	// CredentialsDirectory represents the name of the directory where credentials are stored
	CredentialsDirectory = "credentials"
)

func (d *Runtime) loadCredentials(b *bundle.Bundle, files []string) (map[string]string, error) {
	// TODO: export back outta Compton

	creds := map[string]string{}
	if len(files) == 0 {
		return creds, credentials.Validate(creds, b.Credentials)
	}

	// The strategy here is "last one wins". We loop through each credential file and
	// calculate its credentials. Then we insert them into the creds map in the order
	// in which they were supplied on the CLI.
	for _, file := range files {
		if !d.isPathy(file) {
			credsPath, err := d.Config.GetCredentialsDir()
			if err != nil {
				return nil, err
			}
			file = filepath.Join(credsPath, file+".yaml")
		}
		cset, err := credentials.Load(file)
		if err != nil {
			return creds, err
		}
		res, err := cset.Resolve()
		if err != nil {
			return res, err
		}

		for k, v := range res {
			creds[k] = v
		}
	}
	return creds, credentials.Validate(creds, b.Credentials)
}

// isPathy checks to see if a name looks like a path.
func (d *Runtime) isPathy(name string) bool {
	// TODO: export back outta Compton

	return strings.Contains(name, string(filepath.Separator))
}
