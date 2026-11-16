package mongodb_test

import (
	"testing"

	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/pkg/storage/plugins"
	"get.porter.sh/porter/pkg/storage/plugins/plugintest"
	"get.porter.sh/porter/pkg/storage/plugins/testplugin"
)

// The mongodb backend must pass the same conformance suite as every other
// storage backend. The test plugin runs it against a real MongoDB, which it
// starts in a container when one isn't already running.
func TestBackend_Conformance(t *testing.T) {
	plugintest.RunConformance(t, func(t *testing.T) plugins.StorageBackend {
		tc := portercontext.NewTestContext(t)
		backend := testplugin.NewTestStoragePlugin(tc)
		t.Cleanup(func() { backend.Close() })
		return backend
	})
}
