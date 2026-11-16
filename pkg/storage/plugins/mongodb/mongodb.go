package mongodb

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/pkg/tracing"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
	"go.opentelemetry.io/otel/attribute"
)

var (
	ErrNotConnected = errors.New("cannot execute command against the mongodb plugin because the session is closed (or was never connected)")
)

// Store implements the Porter plugin.StoragePlugin interface for mongodb.
type Store struct {
	*portercontext.Context

	url      string
	database string
	client   *mongo.Client
	timeout  time.Duration
}

// NewStore creates a new storage engine that uses MongoDB.
func NewStore(c *portercontext.Context, cfg PluginConfig) *Store {
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 10 // default to 10 seconds
	}
	return &Store{
		Context: c,
		url:     cfg.URL,
		timeout: time.Duration(timeout) * time.Second,
	}
}

// Connect initializes the plugin for use.
// The plugin itself is responsible for ensuring it was called.
// Close is called automatically when the plugin is used by Porter.
func (s *Store) Connect(ctx context.Context) error {
	if s.client != nil {
		return nil
	}

	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()

	connStr, err := connstring.ParseAndValidate(s.url)
	if err != nil {
		// I'm not tracing additional information like the url since it is sensitive data
		return span.Error(fmt.Errorf("invalid mongodb connection string"))
	}

	if connStr.Database == "" {
		s.database = "porter"
	} else {
		s.database = strings.TrimSuffix(connStr.Database, "/")
	}
	span.SetAttributes(attribute.String("database", s.database))

	cxt, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	client, err := mongo.Connect(cxt, options.Client().ApplyURI(s.url))
	if err != nil {
		return span.Error(err)
	}

	s.client = client
	return nil
}

func (s *Store) Close() error {
	if s.client != nil {
		cxt, cancel := context.WithTimeout(context.Background(), s.timeout)
		defer cancel()

		if err := s.client.Disconnect(cxt); err != nil {
			return err
		}
		s.client = nil
	}
	return nil
}

// Ping the connected session to check if everything is okay.
func (s *Store) Ping(ctx context.Context) error {
	if err := s.Connect(ctx); err != nil {
		return err
	}

	cxt, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	return s.client.Ping(cxt, readpref.Primary())
}

func (s *Store) getCollection(collection string) *mongo.Collection {
	return s.client.Database(s.database).Collection(collection)
}

// RemoveDatabase removes the current database.
func (s *Store) RemoveDatabase(ctx context.Context) error {
	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()

	if err := s.Connect(ctx); err != nil {
		return err
	}

	cxt, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	span.Info("Dropping database", attribute.String("database", s.database))
	return s.client.Database(s.database).Drop(cxt)
}
