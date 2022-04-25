package filesystem

import (
	"os"
	"path/filepath"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/secrets"
	"get.porter.sh/porter/pkg/secrets/plugins"
	"get.porter.sh/porter/pkg/secrets/plugins/host"
	"github.com/carolynvs/aferox"
	cnabsecrets "github.com/cnabio/cnab-go/secrets"
	"github.com/hashicorp/go-hclog"
	"github.com/pkg/errors"
)

var _ plugins.SecretsPlugin = &Store{}

const (
	SECRET_FOLDER                          = "secrets"
	FileModeSensitiveDirectory os.FileMode = 0700
	FileModeSensitiveWritable  os.FileMode = 0600
)

// Config contains information needed for creatina a new store.
type Config struct {
	secretDir     string
	debugLog      bool
	porterHomeDir string
}

// NewConfig returns a new instance of Config.
func NewConfig(debug bool, porterHomeDir string) Config {
	return Config{
		debugLog:      debug,
		porterHomeDir: porterHomeDir,
	}
}

// Valid checks if the configuration has been properly set.
func (c Config) Valid() bool {
	return c.secretDir != ""
}

// SetSecretDir configures the directory path for storing secrets.
func (c *Config) SetSecretDir() (string, error) {
	var err error
	porterHomeDir := c.porterHomeDir
	if porterHomeDir == "" {
		porterCfg := config.New()
		porterHomeDir, err = porterCfg.GetHomeDir()
		if err != nil {
			return "", errors.Wrap(err, "could not get user home directory")
		}
	}

	c.secretDir = filepath.Join(porterHomeDir, SECRET_FOLDER)
	return c.secretDir, nil
}

// Store implements an file system secrets store for testing and local
// development.
type Store struct {
	config    Config
	logger    hclog.Logger
	hostStore cnabsecrets.Store
	storage   aferox.Aferox
}

// NewStore returns a new instance of the filesystem secret store.
func NewStore(cfg Config, logger hclog.Logger, storage aferox.Aferox) *Store {
	if cfg.debugLog {
		logger.SetLevel(hclog.Debug)
	}

	s := &Store{
		config:    cfg,
		logger:    logger,
		hostStore: host.NewPlugin(),
		storage:   storage,
	}

	return s
}

// Connect implements the Connect method on the secret plugins' interface.
func (s *Store) Connect() error {
	if !s.config.Valid() {
		return errors.New("invalid filesystem configuration")
	}

	if err := s.storage.MkdirAll(s.config.secretDir, FileModeSensitiveDirectory); err != nil && !errors.Is(err, os.ErrExist) {
		return err
	}

	s.logger.Debug("storing secrets in %s", s.config.secretDir)

	return nil
}

// Close implements the Close method on the secret plugins' interface.
func (s *Store) Close() error {
	return nil
}

// Resolve implements the Resolve method on the secret plugins' interface.
func (s *Store) Resolve(keyName string, keyValue string) (string, error) {
	// check if the keyName is secret
	if keyName != secrets.SourceSecret {
		return s.hostStore.Resolve(keyName, keyValue)
	}

	path := filepath.Join(s.config.secretDir, keyValue)
	data, err := s.storage.ReadFile(path)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// Create implements the Create method on the secret plugins' interface.
func (s *Store) Create(keyName string, keyValue string, value string) error {
	// check if the keyName is secret
	if keyName != secrets.SourceSecret {
		return errors.New("invalid key name: " + keyName)
	}

	path := filepath.Join(s.config.secretDir, keyValue)
	f, err := s.storage.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, FileModeSensitiveWritable)
	if err != nil {
		return errors.Wrapf(err, "failed to create key: %s", keyName)
	}
	defer f.Close()

	_, err = f.WriteString(value)
	return err
}
