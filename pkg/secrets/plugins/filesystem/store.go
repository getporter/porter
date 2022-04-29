package filesystem

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/secrets"
	"get.porter.sh/porter/pkg/secrets/plugins"
	"get.porter.sh/porter/pkg/secrets/plugins/host"
	"get.porter.sh/porter/pkg/tracing"
	"github.com/pkg/errors"
)

var _ plugins.SecretsProtocol = &Store{}

const (
	SECRET_FOLDER                          = "secrets"
	FileModeSensitiveDirectory os.FileMode = 0700
	FileModeSensitiveWritable  os.FileMode = 0600
)

// Store implements an file system secrets store for testing and local
// development.
type Store struct {
	config    *config.Config
	secretDir string
	hostStore plugins.SecretsProtocol
}

// NewStore returns a new instance of the filesystem secret store.
func NewStore(c *config.Config) *Store {
	s := &Store{
		config:    c,
		hostStore: host.NewStore(),
	}

	return s
}

// Connect initializes the plugin for use.
// The plugin itself is responsible for ensuring it was called.
// Close is called automatically when the plugin is used by Porter.
func (s *Store) Connect(ctx context.Context) error {
	if s.secretDir != "" {
		return nil
	}

	ctx, log := tracing.StartSpan(ctx)
	defer log.EndSpan()

	if _, err := s.SetSecretDir(); err != nil {
		return log.Error(err)
	}

	if err := s.config.FileSystem.MkdirAll(s.secretDir, FileModeSensitiveDirectory); err != nil && !errors.Is(err, os.ErrExist) {
		return log.Error(err)
	}

	log.Debugf("storing secrets in %s", s.secretDir)
	return nil
}

// SetSecretDir configures the directory path for storing secrets.
func (s *Store) SetSecretDir() (string, error) {
	porterHomeDir, err := s.config.GetHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not get user home directory: %w", err)
	}

	s.secretDir = filepath.Join(porterHomeDir, SECRET_FOLDER)
	return s.secretDir, nil
}

// Close implements the Close method on the secret plugins' interface.
func (s *Store) Close() error {
	return nil
}

// Resolve implements the Resolve method on the secret plugins' interface.
func (s *Store) Resolve(ctx context.Context, keyName string, keyValue string) (string, error) {
	ctx, log := tracing.StartSpan(ctx)
	defer log.EndSpan()

	if err := s.Connect(ctx); err != nil {
		return "", err
	}

	// check if the keyName is secret
	if keyName != secrets.SourceSecret {
		value, err := s.hostStore.Resolve(ctx, keyName, keyValue)
		return value, log.Error(err)
	}

	path := filepath.Join(s.secretDir, keyValue)
	data, err := s.config.FileSystem.ReadFile(path)
	if err != nil {
		return "", log.Error(fmt.Errorf("error reading secret from filesystem: %w", err))
	}

	return string(data), nil
}

// Create implements the Create method on the secret plugins' interface.
func (s *Store) Create(ctx context.Context, keyName string, keyValue string, value string) error {
	ctx, log := tracing.StartSpan(ctx)
	defer log.EndSpan()

	if err := s.Connect(ctx); err != nil {
		return err
	}

	// check if the keyName is secret
	if keyName != secrets.SourceSecret {
		return log.Error(errors.New("invalid key name: " + keyName))
	}

	path := filepath.Join(s.secretDir, keyValue)
	f, err := s.config.FileSystem.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, FileModeSensitiveWritable)
	if err != nil {
		return log.Error(fmt.Errorf("failed to create key: %s: %w", keyName, err))
	}
	defer f.Close()

	_, err = f.WriteString(value)
	if err != nil {
		return log.Error(fmt.Errorf("error writing secret to filesystem: %w", err))
	}
	return nil
}
