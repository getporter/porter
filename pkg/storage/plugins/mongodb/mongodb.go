package mongodb

import (
	"context"
	"fmt"
	"net/url"
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
}

// NewStore creates a new storage engine that uses MongoDB.
func NewStore(c *portercontext.Context, cfg PluginConfig) *Store {
	return &Store{
		Context: c,
		url:     cfg.URL,
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

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(s.url))

	s.client = client
	return client.Ping(context.Background(), readpref.Primary())
}

func (s *Store) Close() error {
	if s.client != nil {
		s.client.Disconnect(context.Background())
		s.client = nil
	}
	return nil
}

// Ping the connected session to check if everything is okay.
func (s *Store) Ping() error {
	return s.client.Ping(context.Background(), readpref.Primary())
}

func (s *Store) Aggregate(opts plugins.AggregateOptions) ([]bson.Raw, error) {
	// TODO(carolynvs): wrap each call with session.refresh  on error and a single retry
	c := s.getCollection(opts.Collection)
	cxt := context.Background()
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
	c := s.getCollection(opts.Collection)
	cxt := context.Background()

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
	c := s.getCollection(opts.Collection)
	cxt := context.Background()
	return c.CountDocuments(cxt, opts.Filter)
}

func (s *Store) Find(opts plugins.FindOptions) ([]bson.Raw, error) {
	c := s.getCollection(opts.Collection)
	cxt := context.Background()

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
	c := s.getCollection(opts.Collection)
	cxt := context.Background()
	_, err := c.InsertMany(cxt, opts.Documents)
	return err
}

func (s *Store) Patch(opts plugins.PatchOptions) error {
	c := s.getCollection(opts.Collection)
	cxt := context.Background()
	_, err := c.UpdateOne(cxt, opts.QueryDocument, opts.Transformation)
	return err
}

func (s *Store) Remove(opts plugins.RemoveOptions) error {
	c := s.getCollection(opts.Collection)
	cxt := context.Background()

	if opts.All {
		_, err := c.DeleteMany(cxt, opts.Filter)
		return err
	}
	_, err := c.DeleteOne(cxt, opts.Filter)
	return err
}

func (s *Store) Update(opts plugins.UpdateOptions) error {
	c := s.getCollection(opts.Collection)
	cxt := context.Background()
	_, err := c.ReplaceOne(cxt, opts.Filter, opts.Document, &options.ReplaceOptions{Upsert: &opts.Upsert})
	return err
}

func (s *Store) getCollection(collection string) *mongo.Collection {
	return s.client.Database(s.database).Collection(collection)
}

// RemoveDatabase removes the current database.
func (s *Store) RemoveDatabase() error {
	cxt := context.Background()

	fmt.Fprintf(s.Err, "Dropping database %s!\n", s.database)
	return s.client.Database(s.database).Drop(cxt)
}

func parseConnectionString(dialStr string) (host string, port string, database string, err error) {
	u, err := url.Parse(dialStr)
	if err != nil {
		return "", "", "", err
	}

	if u.Path != "" {
		return u.Host, u.Port(), strings.Trim(u.Path, "/"), nil
	}
	// If this returns empty, then the driver is supposed to substitute in the
	// default database.
	return "", "", "", nil
}
