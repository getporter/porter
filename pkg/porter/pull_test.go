package porter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBundlePullOptions_validtag(t *testing.T) {
	opts := BundlePullOptions{
		Reference: "deislabs/kubetest:1.0",
	}

	err := opts.Validate()
	require.NoError(t, err, "valid tag should not produce an error")
	assert.Equal(t, opts.Reference, opts.GetReference().String())
}

func TestBundlePullOptions_invalidtag(t *testing.T) {
	opts := BundlePullOptions{
		Reference: "deislabs/kubetest:1.0:ahjdljahsdj",
	}

	err := opts.Validate()
	require.Error(t, err, "invalid tag should produce an error")
}
