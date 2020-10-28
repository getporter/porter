package porter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBundlePullOptions_validtag(t *testing.T) {
	opts := BundlePullOptions{
		Tag: "deislabs/kubetest:1.0",
	}

	err := opts.validateTag()
	assert.NoError(t, err, "valid tag should not produce an error")
}

func TestBundlePullOptions_invalidtag(t *testing.T) {
	opts := BundlePullOptions{
		Tag: "deislabs/kubetest:1.0:ahjdljahsdj",
	}

	err := opts.validateTag()
	assert.Error(t, err, "invalid tag should produce an error")
}
