package mongodb

import (
	"context"
	"testing"

	"get.porter.sh/porter/pkg/portercontext"
	"github.com/stretchr/testify/assert"
)

func TestParseDatabase(t *testing.T) {
	tc := portercontext.NewTestContext(t)
	t.Run("db specified", func(t *testing.T) {
		mongo := NewStore(tc.Context, PluginConfig{URL: "mongodb://localhost:27017/test/"})
		mongo.Connect(context.Background())
		defer mongo.Close(context.Background())
		assert.Equal(t, "test", mongo.database)
	})

	t.Run("default db", func(t *testing.T) {
		mongo := NewStore(tc.Context, PluginConfig{URL: "mongodb://localhost:27017"})
		mongo.Connect(context.Background())
		defer mongo.Close(context.Background())
		assert.Equal(t, "porter", mongo.database)
	})
}
