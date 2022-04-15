package mongodb

import (
	"context"
	"fmt"
	"strings"
	"time"

	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/pkg/storage/plugins"
	"get.porter.sh/porter/pkg/tracing"
	"github.com/davecgh/go-spew/spew"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
	"go.opentelemetry.io/otel/attribute"
)

var (
	_               plugins.StorageProtocol = &Store{}
	ErrNotConnected                         = errors.New("cannot execute command against the mongodb plugin because the session is closed (or was never connected)")
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

func (s *Store) Close(ctx context.Context) error {
	if s.client != nil {
		cxt, cancel := context.WithTimeout(ctx, s.timeout)
		defer cancel()

		s.client.Disconnect(cxt)
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

func (s *Store) Aggregate(ctx context.Context, opts plugins.AggregateOptions) ([]bson.Raw, error) {
	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()
	if err := s.Connect(ctx); err != nil {
		return nil, err
	}

	// TODO(carolynvs): wrap each call with session.refresh  on error and a single retry
	c := s.getCollection(opts.Collection)

	cxt, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	cur, err := c.Aggregate(cxt, opts.Pipeline)
	if err != nil {
		return nil, err
	}

	var results []bson.Raw
	for cur.Next(cxt) {
		results = append(results, cur.Current)
	}
	return results, err
}

// EnsureIndexes makes sure that the specified indexes exist and are
// defined appropriately.
func (s *Store) EnsureIndex(ctx context.Context, opts plugins.EnsureIndexOptions) error {
	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()
	if err := s.Connect(ctx); err != nil {
		return err
	}

	indices := make(map[string][]mongo.IndexModel, len(opts.Indices))
	for _, index := range opts.Indices {
		model := mongo.IndexModel{
			Keys:    index.Keys,
			Options: options.Index(),
		}
		model.Options.SetUnique(index.Unique)
		model.Options.SetBackground(true)

		c, ok := indices[index.Collection]
		if !ok {
			c = make([]mongo.IndexModel, 0, 1)
		}
		c = append(c, model)
		indices[index.Collection] = c
	}

	cxt, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	for collectionName, models := range indices {
		c := s.getCollection(collectionName)
		if _, err := c.Indexes().CreateMany(cxt, models); err != nil {
			return span.Error(fmt.Errorf("invalid index specified: %v: %w", spew.Sdump(models), err))
		}
	}

	return nil
}

func (s *Store) Count(ctx context.Context, opts plugins.CountOptions) (int64, error) {
	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()
	if err := s.Connect(ctx); err != nil {
		return 0, err
	}

	c := s.getCollection(opts.Collection)

	cxt, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	count, err := c.CountDocuments(cxt, opts.Filter)
	return count, span.Error(err)
}

func (s *Store) Find(ctx context.Context, opts plugins.FindOptions) ([]bson.Raw, error) {
	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()
	if err := s.Connect(ctx); err != nil {
		return nil, err
	}

	c := s.getCollection(opts.Collection)
	findOpts, err := s.buildFindOptions(opts)
	if err != nil {
		return nil, span.Error(err)
	}

	cxt, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	cur, err := c.Find(cxt, opts.Filter, findOpts)
	if err != nil {
		return nil, span.Error(err)
	}

	var results []bson.Raw
	for cur.Next(cxt) {
		results = append(results, cur.Current)
	}
	return results, span.Error(err)
}

func (s *Store) buildFindOptions(opts plugins.FindOptions) (*options.FindOptions, error) {
	query := options.Find()

	if opts.Select != nil {
		query.SetProjection(opts.Select)
	}

	if opts.Limit > 0 {
		query.SetLimit(opts.Limit)
	}

	if opts.Skip > 0 {
		query.SetSkip(opts.Skip)
	}

	if opts.Sort != nil {
		query.SetSort(opts.Sort)
	}

	return query, nil
}

func (s *Store) Insert(ctx context.Context, opts plugins.InsertOptions) error {
	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()

	if err := s.Connect(ctx); err != nil {
		return err
	}

	c := s.getCollection(opts.Collection)

	cxt, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	docs := make([]interface{}, len(opts.Documents))
	for i, doc := range opts.Documents {
		docs[i] = doc
	}
	_, err := c.InsertMany(cxt, docs)
	return span.Error(err)
}

func (s *Store) Patch(ctx context.Context, opts plugins.PatchOptions) error {
	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()

	if err := s.Connect(ctx); err != nil {
		return err
	}

	c := s.getCollection(opts.Collection)

	cxt, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	_, err := c.UpdateOne(cxt, opts.QueryDocument, opts.Transformation)
	return span.Error(err)
}

func (s *Store) Remove(ctx context.Context, opts plugins.RemoveOptions) error {
	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()

	if err := s.Connect(ctx); err != nil {
		return err
	}

	c := s.getCollection(opts.Collection)

	cxt, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	if opts.All {
		_, err := c.DeleteMany(ctx, opts.Filter)
		return span.Error(err)
	}

	_, err := c.DeleteOne(cxt, opts.Filter)
	return span.Error(err)
}

func (s *Store) Update(ctx context.Context, opts plugins.UpdateOptions) error {
	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()

	if err := s.Connect(ctx); err != nil {
		return err
	}

	c := s.getCollection(opts.Collection)

	cxt, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	_, err := c.ReplaceOne(cxt, opts.Filter, opts.Document, &options.ReplaceOptions{Upsert: &opts.Upsert})
	return span.Error(err)
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
