package porter

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPorter_AddMixin(t *testing.T) {
	t.Skip("not implemented")
	p := NewTestPorter(t)
	p.TestConfig.SetupPorterHome()

	err := p.addMixins(nil)
	require.NoError(t, err)
}
