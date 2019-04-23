package mixin

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInstallOptions_Validate_MixinName(t *testing.T) {
	opts := InstallOptions{}

	err := opts.validateMixinName([]string{"helm"})
	require.NoError(t, err)
	assert.Equal(t, "helm", opts.Name)
}

func TestInstallOptions_Validate_MissingMixinName(t *testing.T) {
	opts := InstallOptions{}

	err := opts.Validate(nil)
	assert.EqualError(t, err, "no mixin name was specified")
}

func TestInstallOptions_Validate_BadURL(t *testing.T) {
	opts := InstallOptions{
		URL: ":#",
	}

	err := opts.Validate([]string{"helm"})
	assert.EqualError(t, err, "invalid --url :#: parse :: missing protocol scheme")
}

func TestInstallOptions_Validate_BadFeedURL(t *testing.T) {
	opts := InstallOptions{
		FeedURL: ":#",
	}

	err := opts.Validate([]string{"helm"})
	assert.EqualError(t, err, "invalid --feed-url :#: parse :: missing protocol scheme")
}

func TestInstallOptions_Validate_MissingURL(t *testing.T) {
	opts := InstallOptions{
		URL: "",
	}

	err := opts.Validate([]string{"helm"})
	assert.EqualError(t, err, "either --url or --feed-url is required")
}

func TestInstallOptions_Validate_FeedUrlSpecified(t *testing.T) {
	opts := InstallOptions{
		FeedURL: "https://example.com/atom.xml",
	}

	err := opts.Validate([]string{"helm"})
	assert.NoError(t, err)
}

func TestInstallOptions_Validate_DefaultVersion(t *testing.T) {
	opts := InstallOptions{
		URL: "https://example.com",
	}

	opts.Validate([]string{"helm"})

	assert.Equal(t, "latest", opts.Version)
}
