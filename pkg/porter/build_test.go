package porter

import (
	"testing"

	"get.porter.sh/porter/pkg"
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
	assert.Equal(t, 1, testMixins.GetCalled("exec"), "expected the exec mixin to be called once")
}

func TestPorter_ErrorMessageOnMissingPorterYaml(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Close()

	o := BuildOptions{
		BundleDefinitionOptions: BundleDefinitionOptions{},
	}

	err := o.Validate(p.Porter)
	require.ErrorContains(t, err, "could not find porter.yaml in the current directory %s, make sure you are in the right directory or specify the porter manifest with --file")
}

func TestPorter_NoErrorWhenPorterYamlIsPresent(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Close()

	o := BuildOptions{
		BundleDefinitionOptions: BundleDefinitionOptions{},
	}
	p.FileSystem.WriteFile("porter.yaml", []byte(""), pkg.FileModeWritable)

	err := o.Validate(p.Porter)
	require.NoError(t, err, "validate BuildOptions failed")
}
