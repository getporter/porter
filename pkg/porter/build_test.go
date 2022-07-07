package porter

import (
	"testing"

	"get.porter.sh/porter/pkg/manifest"
	"get.porter.sh/porter/pkg/mixin"
	"get.porter.sh/porter/pkg/pkgmgmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPorter_GetUsedMixins(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Close()

	// Add an extra mixin that isn't used by the bundle
	testMixins := p.Mixins.(*mixin.TestMixinProvider)
	testMixins.TestPackageManager.Packages = append(testMixins.TestPackageManager.Packages, &pkgmgmt.Metadata{
		Name: "mymixin",
		VersionInfo: pkgmgmt.VersionInfo{
			Version: "v0.1.0",
			Commit:  "defxyz",
			Author:  "It was Me",
		},
	})

	m := &manifest.Manifest{
		Mixins: []manifest.MixinDeclaration{
			{Name: "exec"},
		},
	}

	results, err := p.getUsedMixins(p.RootContext, m)
	require.NoError(t, err, "getUsedMixins failed")
	assert.Len(t, results, 1)
	assert.Equal(t, map[string]int{"exec": 1}, testMixins.GetCalled(), "expected the exec mixin to be called once")
}
