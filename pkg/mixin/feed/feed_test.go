package feed

import (
	"github.com/deislabs/porter/pkg/context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestMixinFeed_Search_Latest(t *testing.T) {
	tc := context.NewTestContext(t)
	f := NewMixinFeed(tc.Context)

	f.Index["helm"] = make(map[string]*MixinFileset)
	f.Index["helm"]["v1.2.3"] = &MixinFileset{
		Mixin: "helm",
		Version:"v1.2.3",
	}
	f.Index["helm"]["v1.2.4"] = &MixinFileset{
		Mixin: "helm",
		Version:"v1.2.4",
	}

	result := f.Search("helm", "latest")

	require.NotNil(t, result)

	assert.Equal(t, "v1.2.4", result.Version)
}
