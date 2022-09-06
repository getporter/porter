package porter

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPorter_PrintManifestSchema(t *testing.T) {
	ctx := context.Background()
	p := NewTestPorter(t)
	defer p.Close()

	err := p.PrintManifestSchema(ctx)
	require.NoError(t, err)

	p.CompareGoldenFile("testdata/schema.json", p.TestConfig.TestContext.GetOutput())
}
