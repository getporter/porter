package migrations

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	"get.porter.sh/porter/pkg"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/encoding"
	"get.porter.sh/porter/pkg/storage"
	"get.porter.sh/porter/pkg/tracing"
)

// initCacheTTL is how long we trust a previously verified storage
// connection before re-checking its schema and indices.
const initCacheTTL = time.Hour

// initCacheFileName is the name of the init check cache file, stored in
// PORTER_HOME/cache.
const initCacheFileName = "init.json"

// initCacheFile persists the result of the storage schema and index checks
// performed by Manager.Connect, keyed by connectionHash, so that repeated
// CLI invocations against the same storage backend can skip redundant db
// calls. It is safe to delete or edit this file by hand to force a
// re-check; it will simply be recreated.
type initCacheFile struct {
	Storage map[string]initCacheEntry `json:"storage"`
}

// initCacheEntry records that a storage connection's schema and indices
// were verified at a point in time.
type initCacheEntry struct {
	// PorterVersion that performed the check. A version mismatch
	// invalidates the entry since supported schemas/indices can change
	// between releases.
	PorterVersion string `json:"porterVersion"`

	// Schema recorded the last time this connection was verified.
	Schema storage.Schema `json:"schema"`

	// CheckedAt is when this connection was last verified.
	CheckedAt time.Time `json:"checkedAt"`
}

// IsValid reports whether this entry can still be trusted without
// re-checking the database.
func (e initCacheEntry) IsValid() bool {
	return e.PorterVersion == pkg.Version && time.Since(e.CheckedAt) < initCacheTTL
}

// connectionHash returns a stable identifier for the storage connection
// that will be used, so init checks can be cached per-backend and
// automatically invalidated when the configuration changes. Only the hash
// is ever persisted to disk, not the underlying config, so secrets such as
// connection strings are never written to the cache file.
func connectionHash(c *config.Config) (string, error) {
	pluginKey := c.Data.DefaultStoragePlugin
	var pluginConfig interface{}

	if name := c.Data.DefaultStorage; name != "" {
		plugin, err := c.GetStorage(name)
		if err != nil {
			return "", err
		}
		if plugin.PluginSubKey != "" {
			pluginKey = plugin.PluginSubKey
		}
		pluginConfig = plugin.Config
	}

	data, err := json.Marshal(struct {
		Key    string
		Config interface{}
	}{Key: pluginKey, Config: pluginConfig})
	if err != nil {
		return "", fmt.Errorf("could not hash storage connection config: %w", err)
	}

	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}

// initCachePath returns the location of the init check cache file.
func (m *Manager) initCachePath() (string, error) {
	home, err := m.GetHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "cache", initCacheFileName), nil
}

// loadInitCacheEntry returns the cached entry for hash, when present and
// still valid. Any error reading or parsing the cache file is treated as a
// cache miss; it is never fatal to the caller.
func (m *Manager) loadInitCacheEntry(ctx context.Context, hash string) (initCacheEntry, bool) {
	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()

	path, err := m.initCachePath()
	if err != nil {
		return initCacheEntry{}, false
	}

	exists, err := m.FileSystem.Exists(path)
	if err != nil || !exists {
		return initCacheEntry{}, false
	}

	var cacheFile initCacheFile
	if err := encoding.UnmarshalFile(m.FileSystem, path, &cacheFile); err != nil {
		span.Debug("Ignoring unreadable storage init cache file: " + err.Error())
		return initCacheEntry{}, false
	}

	entry, ok := cacheFile.Storage[hash]
	if !ok || !entry.IsValid() {
		return initCacheEntry{}, false
	}

	return entry, true
}

// saveInitCacheEntry records that the connection identified by hash was
// just verified. Errors are logged and swallowed; failing to persist the
// cache should never fail the command that triggered the check.
func (m *Manager) saveInitCacheEntry(ctx context.Context, hash string, schema storage.Schema) {
	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()

	path, err := m.initCachePath()
	if err != nil {
		span.Debug("Unable to persist storage init cache: " + err.Error())
		return
	}

	var cacheFile initCacheFile
	if exists, existsErr := m.FileSystem.Exists(path); existsErr == nil && exists {
		// Best effort: if the existing file is corrupt, overwrite it rather than failing.
		_ = encoding.UnmarshalFile(m.FileSystem, path, &cacheFile)
	}
	if cacheFile.Storage == nil {
		cacheFile.Storage = map[string]initCacheEntry{}
	}

	cacheFile.Storage[hash] = initCacheEntry{
		PorterVersion: pkg.Version,
		Schema:        schema,
		CheckedAt:     time.Now(),
	}

	if err := encoding.MarshalFile(m.FileSystem, path, cacheFile); err != nil {
		span.Debug("Unable to persist storage init cache: " + err.Error())
	}
}
