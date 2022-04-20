package filesystem

import (
	"io"
	"os"
	"path/filepath"

	"get.porter.sh/porter/pkg/secrets"
	"get.porter.sh/porter/pkg/secrets/plugins"
	"get.porter.sh/porter/pkg/secrets/plugins/host"
	cnabsecrets "github.com/cnabio/cnab-go/secrets"
	"github.com/hashicorp/go-hclog"
	"github.com/pkg/errors"
)

var _ plugins.SecretsPlugin = &Store{}

const SECRET_FOLDER = "secrets"

type Config struct {
	// default to `$PORTER_HOME/secrets/`
	PathPrefix string `json:"path_prefix"`
	secretDir  string
}

// Store implements an file system secrets store for testing and local
// development.
type Store struct {
	config    Config
	logger    hclog.Logger
	hostStore cnabsecrets.Store
}

// NewStore returns a new instance of the filesystem secret store.
func NewStore(cfg Config, logger hclog.Logger) *Store {
	s := &Store{
		config:    cfg,
		logger:    logger,
		hostStore: host.NewPlugin(),
	}

	return s
}

// Connect implements the Connect method on the secret plugins' interface.
func (s *Store) Connect() error {
	if s.config.PathPrefix == "" {
		userHome, err := os.UserHomeDir()
		if err != nil {
			return errors.Wrap(err, "could not get user home directory")
		}
		s.config.PathPrefix = filepath.Join(userHome, ".porter")
	}

	secretDir := filepath.Join(s.config.PathPrefix, SECRET_FOLDER)

	if err := os.Mkdir(secretDir, os.ModePerm); err != nil && !errors.Is(err, os.ErrExist) {
		return err
	}
	s.config.secretDir = secretDir

	return nil
}

// Close implements the Close method on the secret plugins' interface.
func (s *Store) Close() error {
	return nil
}

func (s *Store) Resolve(keyName string, keyValue string) (string, error) {
	// check if the keyName is secret
	if keyName != secrets.SourceSecret {
		return s.hostStore.Resolve(keyName, keyValue)
	}

	f, err := s.OpenSecretFile(keyValue, os.O_RDONLY)
	if err != nil {
		return "", errors.Wrapf(err, "failed to resolve key: %s", keyName)
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (s *Store) Create(keyName string, keyValue string, value string) error {
	// check if the keyName is secret
	if keyName != secrets.SourceSecret {
		return errors.New("invalid key name: " + keyName)
	}
	f, err := s.OpenSecretFile(keyValue, os.O_WRONLY|os.O_CREATE|os.O_TRUNC)
	if err != nil {
		return errors.Wrapf(err, "failed to create key: %s", keyName)
	}
	defer f.Close()

	_, err = f.WriteString(value)
	return err
}

func (s *Store) OpenSecretFile(name string, flag int) (*os.File, error) {
	fileInfo, err := os.OpenFile(filepath.Join(s.config.secretDir, name), flag, 0600)
	if err != nil {
		return nil, err
	}

	return fileInfo, nil

}
