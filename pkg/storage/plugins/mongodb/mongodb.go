package mongodb

import (
	"context"
	"fmt"
	"strings"
	"time"

	"get.porter.sh/porter/pkg/plugins"
	"get.porter.sh/porter/pkg/portercontext"
	storageplugins "get.porter.sh/porter/pkg/storage/plugins"
	"get.porter.sh/porter/pkg/tracing"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
	"go.opentelemetry.io/otel/attribute"
	"golang.org/x/sync/errgroup"
)

var (
	_               plugins.Plugin                 = &Store{}
	_               storageplugins.StorageProtocol = &Store{}
	ErrNotConnected                                = errors.New("cannot execute command against the mongodb plugin because the session is closed (or was never connected)")
)

// Store implements the Porter plugin.StoragePlugin interface for mongodb.
type Store struct {
	*portercontext.Context

	cmdCtx   context.Context
	tracer   tracing.Tracer
	url      string
	database string
	client   *mongo.Client
	timeout  time.Duration
}

// NewStore creates a new storage engine that uses MongoDB.
func NewStore(ctx context.Context, c *portercontext.Context, cfg PluginConfig) *Store {
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 10 // default to 10 seconds
	}

	return &Store{
		Context: c,
		tracer:  c.NewTracer(ctx, "porter.storage.mongodb"),
		url:     cfg.URL,
		timeout: time.Duration(timeout) * time.Second,
	}
}

func (s *Store) Connect(ctx context.Context) error {
	if s.client != nil {
		return nil
	}

	ctx, span := tracing.StartSpanForComponent(ctx, s.tracer)
	defer span.EndSpan()

	connStr, err := connstring.ParseAndValidate(s.url)
	if err != nil {
		return span.Error(errors.Wrapf(err, "invalid mongodb connection string %s", s.url))
	}

	if connStr.Database == "" {
		s.database = "porter"
	} else {
		s.database = strings.TrimSuffix(connStr.Database, "/")
	}

	span.Infof("Connecting to mongo database %s at %s", s.database, connStr.Hosts)

	// Trace commands executed
	cmdMonitor := &event.CommandMonitor{
		Started: func(ctx context.Context, evt *event.CommandStartedEvent) {
			log := tracing.LoggerFromContext(ctx)
			log.Warn(evt.CommandName, attribute.String("database", evt.DatabaseName), attribute.String("command", evt.Command.String()))
		},
	}

	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(s.url).SetMonitor(cmdMonitor))
	if err != nil {
		return span.Errorf("error connecting to mongo: %w", err)
	}

	s.client = client
	return nil
}

func (s *Store) Close(ctx context.Context) error {
	ctx, span := tracing.StartSpanForComponent(ctx, s.tracer)
	span.Debug("Closing plugin")
	span.EndSpan()
	s.tracer.Close(ctx)

	if s.client != nil {
		ctx, cancel := context.WithTimeout(ctx, s.timeout)
		defer cancel()

		s.client.Disconnect(ctx)
		s.client = nil
	}

	return nil
}

// Ping the connected session to check if everything is okay.
func (s *Store) Ping() error {
	ctx, span := tracing.StartSpanForComponent(ctx, s.tracer)
	defer span.EndSpan()
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	err := s.client.Ping(ctx, readpref.Primary())
	return span.Error(err)
}

func (s *Store) Aggregate(opts storageplugins.AggregateOptions) ([]bson.Raw, error) {
	ctx, span := s.log.StartSpan(attribute.String("collection", opts.Collection))
	defer span.EndSpan()
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	// TODO(carolynvs): wrap each call with session.refresh  on error and a single retry
	c := s.getCollection(opts.Collection)
	cur, err := c.Aggregate(ctx, opts.Pipeline)
	if err != nil {
		return nil, err
	}

	var results []bson.Raw
	for cur.Next(ctx) {
		results = append(results, cur.Current)
	}
	return results, err
}

// EnsureIndexes makes sure that the specified indexes exist and are
// defined appropriately.
func (s *Store) EnsureIndex(opts storageplugins.EnsureIndexOptions) error {
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

	var g errgroup.Group
	for collectionName, models := range indices {
		g.Go(func() error {
			ctx, span := s.log.StartSpanWithName("CreateIndices", attribute.String("collection", collectionName))
			defer span.EndSpan()
			ctx, cancel := context.WithTimeout(ctx, s.timeout)
			defer cancel()

			c := s.getCollection(collectionName)
			_, err := c.Indexes().CreateMany(ctx, models)
			return err
		})
	}

	return g.Wait()
}

func (s *Store) Count(opts storageplugins.CountOptions) (int64, error) {
	ctx, span := s.log.StartSpan()
	defer span.EndSpan()
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	c := s.getCollection(opts.Collection)
	return c.CountDocuments(ctx, opts.Filter)
}

func (s *Store) Find(opts storageplugins.FindOptions) ([]bson.Raw, error) {
	ctx, span := s.log.StartSpan(attribute.String("collection", opts.Collection), makeFilterAttr(opts.Filter))
	defer span.EndSpan()
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	c := s.getCollection(opts.Collection)
	findOpts := s.buildFindOptions(opts)
	span.SetAttributes(attribute.String("query", fmt.Sprintf("%v", opts.Filter)))

	cur, err := c.Find(ctx, opts.Filter, findOpts)
	if err != nil {
		return nil, span.Error(errors.Wrapf(err, "find failed:\n%#v\n%#v", opts.Filter, findOpts))
	}

	var results []bson.Raw
	for cur.Next(ctx) {
		results = append(results, cur.Current)
	}

	return results, nil
}

func makeFilterAttr(value interface{}) attribute.KeyValue {
	return attribute.String("filter", fmt.Sprintf("%v", value))
}

func (s *Store) buildFindOptions(opts storageplugins.FindOptions) *options.FindOptions {
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

	return query
}

func (s *Store) Insert(opts storageplugins.InsertOptions) error {
	ctx, span := s.log.StartSpan(attribute.String("collection", opts.Collection))
	defer span.EndSpan()
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	c := s.getCollection(opts.Collection)
	_, err := c.InsertMany(ctx, opts.Documents)
	return err
}

func (s *Store) Patch(opts storageplugins.PatchOptions) error {
	ctx, span := s.log.StartSpan(attribute.String("collection", opts.Collection))
	defer span.EndSpan()
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	c := s.getCollection(opts.Collection)
	_, err := c.UpdateOne(ctx, opts.QueryDocument, opts.Transformation)
	return err
}

func (s *Store) Remove(opts storageplugins.RemoveOptions) error {
	ctx, span := s.log.StartSpan(attribute.String("collection", opts.Collection), makeFilterAttr(opts.Filter), attribute.Bool("all", opts.All))
	defer span.EndSpan()
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	c := s.getCollection(opts.Collection)
	if opts.All {
		_, err := c.DeleteMany(ctx, opts.Filter)
		return err
	}
	_, err := c.DeleteOne(ctx, opts.Filter)
	return err
}

func (s *Store) Update(opts storageplugins.UpdateOptions) error {
	ctx, span := s.log.StartSpan(attribute.String("collection", opts.Collection), makeFilterAttr(opts.Filter))
	defer span.EndSpan()
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	c := s.getCollection(opts.Collection)
	_, err := c.ReplaceOne(ctx, opts.Filter, opts.Document, &options.ReplaceOptions{Upsert: &opts.Upsert})
	return err
}

func (s *Store) getCollection(collection string) *mongo.Collection {
	return s.client.Database(s.database).Collection(collection)
}

// RemoveDatabase removes the current database.
func (s *Store) RemoveDatabase() error {
	ctx, span := s.log.StartSpan(attribute.String("database", s.database))
	defer span.EndSpan()
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	span.Debugf("Dropping database %s!", s.database)
	return s.client.Database(s.database).Drop(ctx)
}
