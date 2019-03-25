package porter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInstallOptions_Prepare(t *testing.T) {
	opts := InstallOptions{
		RawParams: []string{"A=1", "B=2"},
	}

	err := opts.Prepare()
	require.NoError(t, err)

	assert.Len(t, opts.Params, 2)
}
