package porter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInvokeOptions_Validate_ActionRequired(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Teardown()

	opts := NewInvokeOptions()

	err := opts.Validate(nil, p.Porter)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "--action is required")
}
