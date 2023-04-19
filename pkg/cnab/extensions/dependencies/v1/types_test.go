package v1

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDependencies_ListBySequence(t *testing.T) {
	t.Parallel()

	sequenceMock := []string{"nginx", "storage", "mysql"}

	rawDeps := Dependencies{
		Sequence: sequenceMock,
		Requires: map[string]Dependency{
			"mysql": {
				Name:   "mysql",
				Bundle: "somecloud/mysql",
				Version: &DependencyVersion{
					AllowPrereleases: true,
					Ranges:           []string{"5.7.x"},
				},
			},
			"storage": {
				Name:   "storage",
				Bundle: "somecloud/blob-storage",
			},
			"nginx": {
				Name:   "nginx",
				Bundle: "localhost:5000/nginx:1.19",
			},
		},
	}

	orderedDeps := rawDeps.ListBySequence()

	assert.NotNil(t, orderedDeps, "Dependencies was not populated")
	assert.Len(t, orderedDeps, 3, "Dependencies.Requires is the wrong length")

	assert.NotNil(t, orderedDeps[0], "expected Dependencies.Requires to have an entry for 'storage")
	assert.NotNil(t, orderedDeps[1], "expected Dependencies.Requires to have an entry for 'mysql'")
	assert.NotNil(t, orderedDeps[2], "expected Dependencies.Requires to have an entry for 'nginx'")

	// assert the bundles are sorted as sequenced
	assert.Equal(t, sequenceMock[0], orderedDeps[0].Name, "unexpected order of the dependencies")
	assert.Equal(t, sequenceMock[1], orderedDeps[1].Name, "unexpected order of the dependencies")
	assert.Equal(t, sequenceMock[2], orderedDeps[2].Name, "unexpected order of the dependencies")
}
