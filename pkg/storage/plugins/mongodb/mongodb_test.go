package mongodb

import (
	"testing"

	"get.porter.sh/porter/pkg/context"
	"github.com/stretchr/testify/assert"
)

func TestParseDatabase(t *testing.T) {
	tc := context.NewTestContext(t)
	t.Run("db specified", func(t *testing.T) {
		mongo := NewStore(tc.Context, PluginConfig{URL: "mongodb://localhost:27017/test/"})
		mongo.Connect()
		defer mongo.Close()
		assert.Equal(t, "test", mongo.database)
	})

	t.Run("default db", func(t *testing.T) {
		mongo := NewStore(tc.Context, PluginConfig{URL: "mongodb://localhost:27017"})
		mongo.Connect()
		defer mongo.Close()
		assert.Equal(t, "porter", mongo.database)
	})
}
