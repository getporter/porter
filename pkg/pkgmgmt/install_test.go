package pkgmgmt

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInstallOptions_ValidateName(t *testing.T) {
	t.Run("no name", func(t *testing.T) {
		opts := InstallOptions{}
		err := opts.validateName(nil)
		require.EqualError(t, err, "no name was specified")
	})
	t.Run("name specified", func(t *testing.T) {
		opts := InstallOptions{}
		err := opts.validateName([]string{"thename"})
		require.NoError(t, err)
		assert.Equal(t, "thename", opts.Name, "the package name was not captured")
	})
	t.Run("multiple names specified", func(t *testing.T) {
		opts := InstallOptions{}
		err := opts.validateName([]string{"name1", "name2"})
		require.EqualError(t, err, "only one positional argument may be specified, the name, but multiple were received: [name1 name2]")
	})
}

func TestInstallOptions_DefaultVersion(t *testing.T) {
	t.Run("none specified", func(t *testing.T) {
		opts := InstallOptions{}
		opts.defaultVersion()
		assert.Equal(t, "latest", opts.Version, "we should default to installing the latest version")
	})
	t.Run("version specified", func(t *testing.T) {
		opts := InstallOptions{Version: "canary"}
		opts.defaultVersion()
		assert.Equal(t, "canary", opts.Version, "defaultVersion should not overwrite the user's choice")
	})
}

func TestInstallOptions_ValidateFeedURL(t *testing.T) {
	t.Run("feed url unset, mirror unset", func(t *testing.T) {
		opts := InstallOptions{
			PackageType: "mixin",
		}
		err := opts.Validate([]string{"mypkg"})
		require.NoError(t, err)
		wantFeedURL := "https://cdn.porter.sh/mixins/atom.xml"
		assert.Equal(t, wantFeedURL, opts.FeedURL, "fallback to the default feed url when nothing is specified")

		parsedFeedURL := opts.GetParsedFeedURL()
		assert.Equal(t, wantFeedURL, parsedFeedURL.String(), "validateFeedURL should parse the feed url")
	})
	t.Run("feed url unset, mirror set", func(t *testing.T) {
		opts := InstallOptions{
			PackageType:            "plugin",
			PackageDownloadOptions: PackageDownloadOptions{Mirror: "https://example.com:81/porter"},
		}
		err := opts.Validate([]string{"mypkg"})
		require.NoError(t, err)
		wantFeedURL := "https://example.com:81/porter/plugins/atom.xml"
		assert.Equal(t, wantFeedURL, opts.FeedURL, "fallback to the mirror when nothing is specified")

		parsedFeedURL := opts.GetParsedFeedURL()
		assert.Equal(t, wantFeedURL, parsedFeedURL.String(), "validateFeedURL should parse the feed url")
	})
	t.Run("user specified feed url", func(t *testing.T) {
		opts := InstallOptions{
			PackageType: "mixin",
			FeedURL:     "https://example.com/atom.xml",
		}
		err := opts.Validate([]string{"mypkg"})
		require.NoError(t, err)

		parsedFeedURL := opts.GetParsedFeedURL()
		assert.Equal(t, opts.FeedURL, parsedFeedURL.String(), "validateFeedURL should parse the feed url")
	})
	t.Run("user specified url", func(t *testing.T) {
		opts := InstallOptions{
			PackageType: "plugin",
			URL:         "https://example.com/mymixin",
		}
		err := opts.Validate([]string{"mypkg"})
		require.NoError(t, err)
		assert.Nil(t, opts.parsedFeedURL, "validateFeedURL shouldn't try to parse an empty URL")
	})
	t.Run("invalid feed url specified", func(t *testing.T) {
		opts := InstallOptions{
			PackageType: "mixin",
			FeedURL:     "$://example.com",
		}
		err := opts.Validate([]string{"mypkg"})
		assert.Contains(t, err.Error(), fmt.Sprintf("invalid --feed-url %s", opts.FeedURL))
		assert.Contains(t, err.Error(), "first path segment in URL cannot contain colon")
	})
}

func TestInstallOptions_ValidateURL(t *testing.T) {
	t.Run("url unset, mirror unset", func(t *testing.T) {
		opts := InstallOptions{}
		err := opts.validateURL()
		require.NoError(t, err)
		assert.Nil(t, opts.parsedURL, "validateURL shouldn't try to parse an empty URL")
	})
	t.Run("url unset, mirror set", func(t *testing.T) {
		opts := InstallOptions{}
		err := opts.validateURL()
		require.NoError(t, err)
		assert.Nil(t, opts.parsedURL, "validateURL shouldn't try to parse an empty URL")
	})
	t.Run("url specified", func(t *testing.T) {
		opts := InstallOptions{
			URL: "https://example.com/mymixin",
		}
		err := opts.validateURL()
		require.NoError(t, err)
		parsedURL := opts.parsedURL
		assert.Equal(t, opts.URL, parsedURL.String(), "validateURL should parse the URL")
	})
	t.Run("invalid url specified", func(t *testing.T) {
		opts := InstallOptions{
			URL: "$://example.com",
		}
		err := opts.validateURL()
		assert.Contains(t, err.Error(), fmt.Sprintf("invalid --url %s", opts.URL))
		assert.Contains(t, err.Error(), "first path segment in URL cannot contain colon")
	})
}

func TestInstallOptions_Validate(t *testing.T) {
	t.Run("mixin", func(t *testing.T) {
		opts := InstallOptions{
			PackageType: "mixin",
		}
		err := opts.Validate([]string{"pkg"})
		require.NoError(t, err, "Validate failed")
		assert.NotEmpty(t, opts.FeedURL, "Validate should have defaulted the feed")
		assert.NotEmpty(t, opts.Version, "Validate should have defaulted the version")
		assert.NotEmpty(t, opts.Mirror, "Validate should have defaulted the mirror")
	})
	t.Run("plugin", func(t *testing.T) {
		opts := InstallOptions{
			PackageType: "mixin",
		}
		err := opts.Validate([]string{"pkg"})
		require.NoError(t, err, "Validate failed")
		assert.NotEmpty(t, opts.FeedURL, "Validate should have defaulted the feed")
		assert.NotEmpty(t, opts.Version, "Validate should have defaulted the version")
		assert.NotEmpty(t, opts.Mirror, "Validate should have defaulted the mirror")
	})
	t.Run("invalid package type", func(t *testing.T) {
		opts := InstallOptions{
			PackageType: "oops",
		}
		err := opts.Validate([]string{"pkg"})
		require.Error(t, err, "Validate should have failed")
		assert.Contains(t, err.Error(), `invalid package type "oops"`)
	})
}
