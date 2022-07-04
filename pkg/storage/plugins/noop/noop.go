package noop

import (
	"context"

	"get.porter.sh/porter/pkg/storage/plugins"
	"go.mongodb.org/mongo-driver/bson"
)

var (
	_ plugins.StorageProtocol = Store{}
)

// Store implements a noop storage plugin.
type Store struct{}

// NewStore creates a new storage plugin that does nothing.
func NewStore() Store {
	return Store{}
}

// Aggregate does nothing.
func (s Store) Aggregate(ctx context.Context, opts plugins.AggregateOptions) ([]bson.Raw, error) {
	return nil, nil
}

// EnsureIndex does nothing.
func (s Store) EnsureIndex(ctx context.Context, opts plugins.EnsureIndexOptions) error {
	return nil
}

// Count does nothing.
func (s Store) Count(ctx context.Context, opts plugins.CountOptions) (int64, error) {
	return 0, nil
}

// Find does nothing.
func (s Store) Find(ctx context.Context, opts plugins.FindOptions) ([]bson.Raw, error) {
	return nil, nil
}

// Insert does nothing
func (s Store) Insert(ctx context.Context, opts plugins.InsertOptions) error {
	return nil
}

// Patch does nothing
func (s Store) Patch(ctx context.Context, opts plugins.PatchOptions) error {
	return nil
}

// Remove does nothing
func (s Store) Remove(ctx context.Context, opts plugins.RemoveOptions) error {
	return nil
}

// Update does nothing
func (s Store) Update(ctx context.Context, opts plugins.UpdateOptions) error {
	return nil
}

// RemoveDatabase does nothing.
func (s Store) RemoveDatabase(ctx context.Context) error {
	return nil
}
