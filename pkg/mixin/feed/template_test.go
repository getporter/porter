package feed

import (
	"testing"

	"get.porter.sh/porter/pkg/context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateTemplate(t *testing.T) {
	tc := context.NewTestContext(t)

	err := CreateTemplate(tc.Context)

	require.NoError(t, err)
	success, _ := tc.Context.FileSystem.Exists("atom-template.xml")
	assert.True(t, success)
}
