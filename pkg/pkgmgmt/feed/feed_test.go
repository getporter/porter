package feed

import (
	"context"
	"net/url"
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

func TestMixinFileset_FindDownloadURL(t *testing.T) {
	t.Run("darwin/arm64 fallback to amd64", func(t *testing.T) {
		link, _ := url.Parse("https://example.com/mymixin-darwin-amd64")

		fs := MixinFileset{
			Mixin: "mymixin",
			Files: []*MixinFile{
				{URL: link},
			},
		}

		result := fs.FindDownloadURL(context.Background(), "darwin", "arm64")
		assert.Contains(t, result.String(), "amd64", "When an arm64 binary is not available for mac, fallback to using an amd64")
	})

	t.Run("darwin/arm64 binary exists", func(t *testing.T) {
		link, _ := url.Parse("https://example.com/mymixin-darwin-arm64")

		fs := MixinFileset{
			Mixin: "mymixin",
			Files: []*MixinFile{
				{URL: link},
			},
		}

		result := fs.FindDownloadURL(context.Background(), "darwin", "arm64")
		assert.Contains(t, result.String(), "arm64", "When an arm64 binary is available, use it")
	})

	t.Run("non-darwin arm64 no special handling", func(t *testing.T) {
		link, _ := url.Parse("https://example.com/mymixin-myos-amd64")

		fs := MixinFileset{
			Mixin: "mymixin",
			Files: []*MixinFile{
				{URL: link},
			},
		}

		result := fs.FindDownloadURL(context.Background(), "myos", "arm64")
		assert.Nil(t, result)
	})
}
