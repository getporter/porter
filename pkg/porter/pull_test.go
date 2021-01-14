package porter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBundlePullOptions_validtag(t *testing.T) {
	opts := BundlePullOptions{
		Reference: "deislabs/kubetest:1.0",
	}

	err := opts.validateReference()
	assert.NoError(t, err, "valid tag should not produce an error")
}

func TestBundlePullOptions_invalidtag(t *testing.T) {
	opts := BundlePullOptions{
		Reference: "deislabs/kubetest:1.0:ahjdljahsdj",
	}

	err := opts.validateReference()
	assert.Error(t, err, "invalid tag should produce an error")
}

func TestPull_checkForDeprecatedTagValue(t *testing.T) {
	t.Parallel()

	t.Run("tag not set", func(t *testing.T) {
		b := BundlePullOptions{}

		b.checkForDeprecatedTagValue()
		assert.Equal(t, "", b.Tag)
		assert.Equal(t, "", b.Reference)
	})

	t.Run("tag set", func(t *testing.T) {
		b := BundlePullOptions{
			Tag: "getporter/hello:v0.1.0",
		}

		b.checkForDeprecatedTagValue()
		assert.Equal(t, "getporter/hello:v0.1.0", b.Tag)
		assert.Equal(t, "getporter/hello:v0.1.0", b.Reference)
	})
}
