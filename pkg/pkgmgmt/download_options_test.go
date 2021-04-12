package pkgmgmt

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPackageDownloadOptions_Validate(t *testing.T) {
	t.Run("unset", func(t *testing.T) {
		opts := PackageDownloadOptions{}
		require.NoError(t, opts.Validate())
		assert.Equal(t, DefaultPackageMirror, opts.Mirror)
	})

	t.Run("valid url", func(t *testing.T) {
		exampleMirror := "https://example.com"
		opts := PackageDownloadOptions{Mirror: exampleMirror}
		require.NoError(t, opts.Validate())
		assert.Equal(t, exampleMirror, opts.Mirror)
	})

	t.Run("invalid url", func(t *testing.T) {
		opts := PackageDownloadOptions{Mirror: "$://example.com"}
		require.Error(t, opts.Validate())
	})
}

func TestPackageDownloadOptions_GetMirror(t *testing.T) {
	t.Run("mirror unset", func(t *testing.T) {
		opts := PackageDownloadOptions{}
		mirror := opts.GetMirror()
		assert.Equal(t, DefaultPackageMirror, mirror.String())
	})

	t.Run("mirror set", func(t *testing.T) {
		exampleMirror := "https://example.com"
		opts := PackageDownloadOptions{Mirror: exampleMirror}
		require.NoError(t, opts.Validate())

		result := opts.GetMirror()
		assert.Equal(t, exampleMirror, result.String())
	})
}
