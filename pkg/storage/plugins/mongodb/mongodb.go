package mongodb

import (
	"fmt"
	"net/url"
	"strings"

	"get.porter.sh/porter/pkg/context"
	"get.porter.sh/porter/pkg/storage/plugins"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/pkg/errors"
)

var (
	_               plugins.StoragePlugin = &Store{}
	ErrNotConnected                       = errors.New("cannot execute command against the mongodb plugin because the session is closed (or was never connected)")
)

// Store implements the Porter plugin.StoragePlugin interface for mongodb.
type Store struct {
	*context.Context

	url      string
	database string
	session  *mgo.Session
}

// NewStore creates a new storage engine that uses MongoDB.
func NewStore(c *context.Context, cfg PluginConfig) *Store {
	return &Store{
		Context: c,
		url:     cfg.URL,
	}
}

func (s *Store) Connect() error {
	if s.session != nil {
		return nil
	}

	host, port, db, err := parseConnectionString(s.url)
	if err != nil {
		return err
	}
	if db == "" {
		db = "porter"
	}
	s.database = db

	if s.Debug {
		fmt.Fprintf(s.Err, "Connecting to mongo database %s at %s on %s\n", s.database, host, port)
	}
	session, err := mgo.Dial(s.url)
	if err != nil {
		return err
	}
	s.session = session
	return nil
}

func (s *Store) Close() error {
	if s.session != nil {
		s.session.Close()
		s.session = nil
	}
	return nil
}

// Ping the connected session to check if everything is okay.
func (s *Store) Ping() error {
	session, err := s.copySession()
	if err != nil {
		return err
	}
	defer session.Close()

	return session.Ping()
}

func (s *Store) Aggregate(opts plugins.AggregateOptions) ([]bson.Raw, error) {
	session, err := s.copySession()
	if err != nil {
		return nil, err
	}
	defer session.Close()

	// TODO(carolynvs): wrap each call with session.refresh  on error and a single retry
	c := s.getCollection(session, opts.Collection)
	var results []bson.Raw
	err = c.Pipe(opts.Pipeline).All(&results)
	return results, err
}

// EnsureIndexes makes sure that the specified indexes exist and are
// defined appropriately.
func (s *Store) EnsureIndex(opts plugins.EnsureIndexOptions) error {
	session, err := s.copySession()
	if err != nil {
		return err
	}
	defer session.Close()

	c := s.getCollection(session, opts.Collection)
	return c.EnsureIndex(opts.Index)
}

func (s *Store) Count(opts plugins.CountOptions) (int, error) {
	session, err := s.copySession()
	if err != nil {
		return 0, err
	}
	defer session.Close()

	c := s.getCollection(session, opts.Collection)
	return c.Find(opts.Filter).Count()
}

func (s *Store) Find(opts plugins.FindOptions) ([]bson.Raw, error) {
	session, err := s.copySession()
	if err != nil {
		return nil, err
	}
	defer session.Close()

	query := s.buildQuery(session, opts)

	var results []bson.Raw
	err = query.All(&results)
	if err != nil {
		return nil, errors.Wrapf(err, "find failed:\n%#v", opts)
	}

	return results, nil
}

func (s *Store) buildQuery(session *mgo.Session, opts plugins.FindOptions) *mgo.Query {
	c := s.getCollection(session, opts.Collection)

	var query *mgo.Query
	if opts.Filter != nil {
		query = c.Find(opts.Filter)
	} else {
		query = c.Find(bson.M{})
	}

	if opts.Select != nil {
		query = query.Select(opts.Select)
	}

	if opts.Limit > 0 {
		query = query.Limit(opts.Limit)
	}

	if opts.Skip > 0 {
		query = query.Skip(opts.Skip)
	}

	if len(opts.Sort) > 0 {
		query = query.Sort(opts.Sort...)
	}
	return query
}

func (s *Store) Insert(opts plugins.InsertOptions) error {
	session, err := s.copySession()
	if err != nil {
		return err
	}
	defer session.Close()

	c := s.getCollection(session, opts.Collection)
	return c.Insert(opts.Documents...)
}

func (s *Store) Patch(opts plugins.PatchOptions) error {
	session, err := s.copySession()
	if err != nil {
		return err
	}
	defer session.Close()

	c := s.getCollection(session, opts.Collection)
	return c.Update(opts.QueryDocument, opts.Transformation)
}

func (s *Store) Remove(opts plugins.RemoveOptions) error {
	session, err := s.copySession()
	if err != nil {
		return err
	}
	defer session.Close()

	c := s.getCollection(session, opts.Collection)
	if opts.All {
		_, err := c.RemoveAll(opts.Filter)
		return err
	}
	return c.Remove(opts.Filter)
}

func (s *Store) Update(opts plugins.UpdateOptions) error {
	session, err := s.copySession()
	if err != nil {
		return err
	}
	defer session.Close()

	c := s.getCollection(session, opts.Collection)
	if opts.Upsert {
		_, err := c.Upsert(opts.Filter, opts.Document)
		return err
	}
	return c.Update(opts.Filter, opts.Document)
}

func (s *Store) copySession() (*mgo.Session, error) {
	if s.session == nil {
		return nil, ErrNotConnected
	}

	return s.session.Copy(), nil
}

func (s *Store) getCollection(session *mgo.Session, collection string) *mgo.Collection {
	return session.DB(s.database).C(collection)
}

// RemoveDatabase removes the current database.
func (s *Store) RemoveDatabase() error {
	session, err := s.copySession()
	if err != nil {
		return err
	}
	defer session.Close()

	fmt.Fprintf(s.Err, "Dropping database %s!\n", s.database)
	return session.DB(s.database).DropDatabase()
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
