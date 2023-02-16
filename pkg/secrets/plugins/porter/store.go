package inmemory

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	"get.porter.sh/porter/pkg/secrets/plugins"
	"get.porter.sh/porter/pkg/storage"
	"github.com/cnabio/cnab-go/secrets/host"
)

var _ plugins.SecretsProtocol = &Store{}

// Store implements an in-process plugin for retrieving values from Porter's database
// This never runs in an external process, or even as an internal plugin, because it requires loading all of Porter's config.
type Store struct {
	Installations storage.InstallationStore
}

func NewStore() *Store {
	return &Store{}
}

var workflowSecretRegexp = regexp.MustCompile(`workflow\.([^.]+)?\.jobs\.([^.]+)?\.outputs\.(.+)`)

func (s *Store) Resolve(ctx context.Context, keyName string, keyValue string) (string, error) {
	if keyName == "porter" {
		// We support retrieving certain data from Porter's database:
		// workflow.WORKFLOWID.jobs.JOBKEY.outputs.OUTPUT
		// The WORKFLOWID is set to the current executing workflow by the workflow engine before running the job
		// It is not stored in the database and is always set dynamically when the job is run.

		matches := workflowSecretRegexp.FindStringSubmatch(keyValue)
		if len(matches) != 4 {
			return "", fmt.Errorf("invalid porter secret mapping value: expected the format workflow.WORKFLOWID.jobs.JOBKEY.outputs.OUTPUT but got %s", keyValue)
		}

		// Lookup the result and run associated with the job run in that workflow
		panic("porter secret store is not implemented")
	}

	// Fallback to the host secret plugin
	hostStore := host.SecretStore{}
	return hostStore.Resolve(keyName, keyValue)
}

func (s *Store) Create(ctx context.Context, keyName string, keyValue string, value string) error {
	return errors.New("The porter secrets plugin does not support the create function, because it is for internal use within Porter only. Chek your porter configuration file and make sure that you are using a supported secrets plugin and not the porter secrets plugin directly.")
}
