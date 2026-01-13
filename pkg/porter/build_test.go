package porter

import (
	"fmt"
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
	require.ErrorContains(t, err, fmt.Sprintf("could not find porter.yaml in the current directory %s, make sure you are in the right directory or specify the porter manifest with --file", o.Dir))
}

func TestPorter_NoErrorWhenPorterYamlIsPresent(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Close()

	o := BuildOptions{
		BundleDefinitionOptions: BundleDefinitionOptions{},
	}
	err := p.FileSystem.WriteFile("porter.yaml", []byte(""), pkg.FileModeWritable)
	require.NoError(t, err)

	err = o.Validate(p.Porter)
	require.NoError(t, err, "validate BuildOptions failed")
}

func TestBuildOptions_Validate_PorterFileConflict(t *testing.T) {
	t.Run("porter file exists", func(t *testing.T) {
		p := NewTestPorter(t)
		defer p.Close()

		// Create porter.yaml first to pass that validation
		err := p.FileSystem.WriteFile("porter.yaml", []byte(""), pkg.FileModeWritable)
		require.NoError(t, err)

		// Create a file named "porter"
		err = p.FileSystem.WriteFile("porter", []byte(""), pkg.FileModeWritable)
		require.NoError(t, err)

		o := BuildOptions{
			BundleDefinitionOptions: BundleDefinitionOptions{},
		}

		err = o.Validate(p.Porter)
		require.Error(t, err)
		require.ErrorContains(t, err, "a file or directory named \"porter\" exists")
		require.ErrorContains(t, err, "will conflict with Porter's internal directory structure")
	})

	t.Run("porter directory exists", func(t *testing.T) {
		p := NewTestPorter(t)
		defer p.Close()

		// Create porter.yaml first to pass that validation
		err := p.FileSystem.WriteFile("porter.yaml", []byte(""), pkg.FileModeWritable)
		require.NoError(t, err)

		// Create a directory named "porter"
		err = p.FileSystem.Mkdir("porter", pkg.FileModeDirectory)
		require.NoError(t, err)

		o := BuildOptions{
			BundleDefinitionOptions: BundleDefinitionOptions{},
		}

		err = o.Validate(p.Porter)
		require.Error(t, err)
		require.ErrorContains(t, err, "a file or directory named \"porter\" exists")
		require.ErrorContains(t, err, "will conflict with Porter's internal directory structure")
	})
}
