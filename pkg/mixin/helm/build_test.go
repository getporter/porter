package helm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMixin_Build(t *testing.T) {
	m := NewTestMixin(t)

	err := m.Build()
	require.NoError(t, err)

	wantOutput := "RUN echo 'TODO: COPY HELM BINARY'"
	gotOutput := m.TestContext.GetOutput()
	assert.Equal(t, wantOutput, gotOutput)
}
