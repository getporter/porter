package porter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInvokeOptions_Validate_ActionRequired(t *testing.T) {
	p := NewTestPorter(t)
	opts := InvokeOptions{}

	err := opts.Validate(nil, p.Context)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "--action is required")
}
