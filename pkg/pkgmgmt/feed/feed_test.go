package feed

import (
	"testing"

	"get.porter.sh/porter/pkg/portercontext"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMixinFeed_Search_Latest(t *testing.T) {
	tc := portercontext.NewTestContext(t)
	f := NewMixinFeed(tc.Context)

	f.Index["helm"] = make(map[string]*MixinFileset)
	f.Index["helm"]["canary"] = &MixinFileset{
		Mixin:   "helm",
		Version: "canary",
	}
	f.Index["helm"]["v1.2.3"] = &MixinFileset{
		Mixin:   "helm",
		Version: "v1.2.3",
	}
	f.Index["helm"]["v1.2.4"] = &MixinFileset{
		Mixin:   "helm",
		Version: "v1.2.4",
	}
	f.Index["helm"]["v2-latest"] = &MixinFileset{
		Mixin:   "helm",
		Version: "v2-latest",
	}
	f.Index["helm"]["v2.0.0-alpha.1"] = &MixinFileset{
		Mixin:   "helm",
		Version: "v2.0.0-alpha.1",
	}

	result := f.Search("helm", "latest")
	require.NotNil(t, result)
	assert.Equal(t, "v1.2.4", result.Version)

	// Now try to get v2-latest specifically
	result = f.Search("helm", "v2-latest")
	require.NotNil(t, result)
	assert.Equal(t, "v2-latest", result.Version)
}

func TestMixinFeed_Search_Canary(t *testing.T) {
	tc := portercontext.NewTestContext(t)
	f := NewMixinFeed(tc.Context)

	f.Index["helm"] = make(map[string]*MixinFileset)
	f.Index["helm"]["canary"] = &MixinFileset{
		Mixin:   "helm",
		Version: "canary",
	}
	f.Index["helm"]["v1.2.4"] = &MixinFileset{
		Mixin:   "helm",
		Version: "v1.2.4",
	}
	f.Index["helm"]["v2-canary"] = &MixinFileset{
		Mixin:   "helm",
		Version: "v2-canary",
	}

	result := f.Search("helm", "canary")
	require.NotNil(t, result)
	assert.Equal(t, "canary", result.Version)

	// Now try to get v2-canary specifically
	result = f.Search("helm", "v2-canary")
	require.NotNil(t, result)
	assert.Equal(t, "v2-canary", result.Version)
}
