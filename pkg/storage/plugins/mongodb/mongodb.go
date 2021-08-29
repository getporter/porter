package mongodb

import (
	"context"
	"fmt"
	"strings"
	"time"

	portercontext "get.porter.sh/porter/pkg/context"
	"get.porter.sh/porter/pkg/storage/plugins"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
)

var (
	_               plugins.StoragePlugin = &Store{}
	ErrNotConnected                       = errors.New("cannot execute command against the mongodb plugin because the session is closed (or was never connected)")
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
	return &Store{
		Context: c,
		url:     cfg.URL,
		timeout: cfg.Timeout,
	}
}

func (s *Store) Connect() error {
	if s.client != nil {
		return nil
	}

	connStr, err := connstring.ParseAndValidate(s.url)
	if err != nil {
		return errors.Wrapf(err, "invalid mongodb connection string %s", s.url)
	}

	if connStr.Database == "" {
		s.database = "porter"
	} else {
		s.database = strings.TrimSuffix(connStr.Database, "/")
	}

	if s.Debug {
		fmt.Fprintf(s.Err, "Connecting to mongo database %s at %s\n", s.database, connStr.Hosts)
	}

	cxt, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	client, err := mongo.Connect(cxt, options.Client().ApplyURI(s.url))
	if err != nil {
		return err
	}

	s.client = client
	return s.Ping()
}

func (s *Store) Close() error {
	if s.client != nil {
		cxt, cancel := context.WithTimeout(context.Background(), s.timeout)
		defer cancel()

		s.client.Disconnect(cxt)
		s.client = nil
	}
	return nil
}

// Ping the connected session to check if everything is okay.
func (s *Store) Ping() error {
	cxt, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	return s.client.Ping(cxt, readpref.Primary())
}

func (s *Store) Aggregate(opts plugins.AggregateOptions) ([]bson.Raw, error) {
	cxt, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	// TODO(carolynvs): wrap each call with session.refresh  on error and a single retry
	c := s.getCollection(opts.Collection)
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
func (s *Store) EnsureIndex(opts plugins.EnsureIndexOptions) error {
	cxt, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	c := s.getCollection(opts.Collection)
	indexOpts := mongo.IndexModel{
		Keys:    opts.Keys,
		Options: options.Index(),
	}
	indexOpts.Options.SetUnique(opts.Unique)
	indexOpts.Options.SetBackground(true)

	_, err := c.Indexes().CreateOne(cxt, indexOpts)
	return err
}

func (s *Store) Count(opts plugins.CountOptions) (int64, error) {
	cxt, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	c := s.getCollection(opts.Collection)
	return c.CountDocuments(cxt, opts.Filter)
}

func (s *Store) Find(opts plugins.FindOptions) ([]bson.Raw, error) {
	cxt, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	c := s.getCollection(opts.Collection)
	findOpts := s.buildFindOptions(opts)
	cur, err := c.Find(cxt, opts.Filter, findOpts)
	if err != nil {
		return nil, errors.Wrapf(err, "find failed:\n%#v\n%#v", opts.Filter, findOpts)
	}

	var results []bson.Raw
	for cur.Next(cxt) {
		results = append(results, cur.Current)
	}

	return results, nil
}

func (s *Store) buildFindOptions(opts plugins.FindOptions) *options.FindOptions {
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

func (s *Store) Insert(opts plugins.InsertOptions) error {
	cxt, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	c := s.getCollection(opts.Collection)
	_, err := c.InsertMany(cxt, opts.Documents)
	return err
}

func (s *Store) Patch(opts plugins.PatchOptions) error {
	cxt, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	c := s.getCollection(opts.Collection)
	_, err := c.UpdateOne(cxt, opts.QueryDocument, opts.Transformation)
	return err
}

func (s *Store) Remove(opts plugins.RemoveOptions) error {
	cxt, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	c := s.getCollection(opts.Collection)
	if opts.All {
		_, err := c.DeleteMany(cxt, opts.Filter)
		return err
	}
	_, err := c.DeleteOne(cxt, opts.Filter)
	return err
}

func (s *Store) Update(opts plugins.UpdateOptions) error {
	cxt, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	c := s.getCollection(opts.Collection)
	_, err := c.ReplaceOne(cxt, opts.Filter, opts.Document, &options.ReplaceOptions{Upsert: &opts.Upsert})
	return err
}

func (s *Store) getCollection(collection string) *mongo.Collection {
	return s.client.Database(s.database).Collection(collection)
}

// RemoveDatabase removes the current database.
func (s *Store) RemoveDatabase() error {
	cxt, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	fmt.Fprintf(s.Err, "Dropping database %s!\n", s.database)
	return s.client.Database(s.database).Drop(cxt)
}
