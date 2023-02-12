package cnab

import (
	"os"
	"strconv"
	"testing"

	"github.com/cnabio/cnab-go/bundle"
	"github.com/stretchr/testify/require"
)

func ReadTestBundle(t *testing.T, path string) ExtendedBundle {
	bunD, err := os.ReadFile(path)
	require.NoError(t, err, "ReadFile failed for %s", path)

	bun, err := bundle.Unmarshal(bunD)
	require.NoError(t, err, "Unmarshal failed for bundle at %s", path)

	return NewBundle(*bun)
}

// TestIDGenerator returns a sequential set of ids (default starting at 0)
// Used for predictable IDs for tests.
type TestIDGenerator struct {
	NextID int
}

func (g *TestIDGenerator) NewID() string {
	id := g.NextID
	g.NextID++
	return strconv.Itoa(id)
}
