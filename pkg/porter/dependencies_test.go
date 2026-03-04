package porter

import (
	"context"
	"encoding/json"
	"testing"

	"get.porter.sh/porter/pkg/cnab"
	cnabtooci "get.porter.sh/porter/pkg/cnab/cnab-to-oci"
	depsv1ext "get.porter.sh/porter/pkg/cnab/extensions/dependencies/v1"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/storage"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// bundleWithV1Ranges marshals a minimal bundle.Bundle that declares a single
// v1 dependency with the given version ranges into JSON.
func bundleWithV1Ranges(t *testing.T, depRepo string, ranges []string) []byte {
	t.Helper()
	deps := depsv1ext.Dependencies{
		Requires: map[string]depsv1ext.Dependency{
			"dep": {
				Bundle:  depRepo,
				Version: &depsv1ext.DependencyVersion{Ranges: ranges},
			},
		},
	}
	bun := bundle.Bundle{
		SchemaVersion: "1.2.0",
		Name:          "testbundle",
		Version:       "0.1.0",
		InvocationImages: []bundle.InvocationImage{
			{BaseImage: bundle.BaseImage{Image: "test/testbundle-installer:0.1.0", ImageType: "docker"}},
		},
		RequiredExtensions: []string{cnab.DependenciesV1ExtensionKey},
		Custom: map[string]interface{}{
			cnab.DependenciesV1ExtensionKey: deps,
		},
	}
	data, err := json.Marshal(bun)
	require.NoError(t, err)
	return data
}

// newExecWithCNABFile creates a dependencyExecutioner wired to a bundle file
// at the given virtual path and with the given install opts.
func newExecWithCNABFile(p *TestPorter, cnabFile string, opts *InstallOptions) *dependencyExecutioner {
	opts.CNABFile = cnabFile
	inst := storage.NewInstallation(opts.Namespace, opts.Name)
	return newDependencyExecutioner(p.Porter, inst, opts)
}

// staticTags returns a MockListTags func that always serves the given tags.
func staticTags(tags []string) func(context.Context, cnab.OCIReference, cnabtooci.RegistryOptions) ([]string, error) {
	return func(_ context.Context, _ cnab.OCIReference, _ cnabtooci.RegistryOptions) ([]string, error) {
		return tags, nil
	}
}

func TestIdentifyDependencies_ExactStrategyWithRangeErrors(t *testing.T) {
	t.Parallel()

	p := NewTestPorter(t)
	defer p.Close()

	bunData := bundleWithV1Ranges(t, "example.com/mysql", []string{">=1.0 <1.3"})
	require.NoError(t, p.FileSystem.WriteFile("/bundle.json", bunData, 0600))

	opts := NewInstallOptions()
	opts.DependenciesVersionStrategy = config.DependencyVersionStrategyExact

	e := newExecWithCNABFile(p, "/bundle.json", &opts)
	err := e.identifyDependencies(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "version strategy is")
}

func TestIdentifyDependencies_EmptyStrategyWithRangeErrors(t *testing.T) {
	t.Parallel()

	p := NewTestPorter(t)
	defer p.Close()

	bunData := bundleWithV1Ranges(t, "example.com/mysql", []string{">=1.0 <1.3"})
	require.NoError(t, p.FileSystem.WriteFile("/bundle.json", bunData, 0600))

	// No flag and no config → defaults to "exact"
	opts := NewInstallOptions()

	e := newExecWithCNABFile(p, "/bundle.json", &opts)
	err := e.identifyDependencies(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "version strategy is")
}

func TestIdentifyDependencies_MaxPatchResolvesHighest(t *testing.T) {
	t.Parallel()

	p := NewTestPorter(t)
	defer p.Close()
	p.TestRegistry.MockListTags = staticTags([]string{"v1.0", "v1.1", "v1.2", "v1.3"})

	bunData := bundleWithV1Ranges(t, "example.com/mysql", []string{">=1.0 <1.3"})
	require.NoError(t, p.FileSystem.WriteFile("/bundle.json", bunData, 0600))

	opts := NewInstallOptions()
	opts.DependenciesVersionStrategy = config.DependencyVersionStrategyMaxPatch

	e := newExecWithCNABFile(p, "/bundle.json", &opts)
	err := e.identifyDependencies(context.Background())
	require.NoError(t, err)
	require.Len(t, e.deps, 1)
	assert.Equal(t, "example.com/mysql:v1.2", e.deps[0].Reference)
}

func TestIdentifyDependencies_MinResolvesLowest(t *testing.T) {
	t.Parallel()

	p := NewTestPorter(t)
	defer p.Close()
	p.TestRegistry.MockListTags = staticTags([]string{"v1.0", "v1.1", "v1.2", "v1.3"})

	bunData := bundleWithV1Ranges(t, "example.com/mysql", []string{">=1.0 <1.3"})
	require.NoError(t, p.FileSystem.WriteFile("/bundle.json", bunData, 0600))

	opts := NewInstallOptions()
	opts.DependenciesVersionStrategy = config.DependencyVersionStrategyMin

	e := newExecWithCNABFile(p, "/bundle.json", &opts)
	err := e.identifyDependencies(context.Background())
	require.NoError(t, err)
	require.Len(t, e.deps, 1)
	assert.Equal(t, "example.com/mysql:v1.0", e.deps[0].Reference)
}

func TestIdentifyDependencies_StrategyFromConfig(t *testing.T) {
	t.Parallel()

	p := NewTestPorter(t)
	defer p.Close()
	p.Config.Data.Dependencies.VersionStrategy = config.DependencyVersionStrategyMin
	p.TestRegistry.MockListTags = staticTags([]string{"v1.0", "v1.1", "v1.2"})

	bunData := bundleWithV1Ranges(t, "example.com/mysql", []string{">=1.0 <1.3"})
	require.NoError(t, p.FileSystem.WriteFile("/bundle.json", bunData, 0600))

	// No flag — should fall back to global config
	opts := NewInstallOptions()

	e := newExecWithCNABFile(p, "/bundle.json", &opts)
	err := e.identifyDependencies(context.Background())
	require.NoError(t, err)
	require.Len(t, e.deps, 1)
	assert.Equal(t, "example.com/mysql:v1.0", e.deps[0].Reference,
		"config strategy min should pick lowest matching version")
}

func TestIdentifyDependencies_FlagOverridesConfig(t *testing.T) {
	t.Parallel()

	p := NewTestPorter(t)
	defer p.Close()
	// Config says max-minor (would pick v1.2)
	p.Config.Data.Dependencies.VersionStrategy = config.DependencyVersionStrategyMaxMinor
	p.TestRegistry.MockListTags = staticTags([]string{"v1.0", "v1.1", "v1.2"})

	bunData := bundleWithV1Ranges(t, "example.com/mysql", []string{">=1.0 <1.3"})
	require.NoError(t, p.FileSystem.WriteFile("/bundle.json", bunData, 0600))

	// Flag overrides to min (should pick v1.0)
	opts := NewInstallOptions()
	opts.DependenciesVersionStrategy = config.DependencyVersionStrategyMin

	e := newExecWithCNABFile(p, "/bundle.json", &opts)
	err := e.identifyDependencies(context.Background())
	require.NoError(t, err)
	require.Len(t, e.deps, 1)
	assert.Equal(t, "example.com/mysql:v1.0", e.deps[0].Reference,
		"flag should override config strategy")
}

func TestIdentifyDependencies_UpgradeUsesStrategy(t *testing.T) {
	t.Parallel()

	p := NewTestPorter(t)
	defer p.Close()
	p.TestRegistry.MockListTags = staticTags([]string{"v1.0", "v1.1", "v1.2", "v1.3"})

	bunData := bundleWithV1Ranges(t, "example.com/mysql", []string{">=1.0 <1.3"})
	require.NoError(t, p.FileSystem.WriteFile("/bundle.json", bunData, 0600))

	opts := NewUpgradeOptions()
	opts.CNABFile = "/bundle.json"
	opts.DependenciesVersionStrategy = config.DependencyVersionStrategyMaxMinor

	inst := storage.NewInstallation(opts.Namespace, opts.Name)
	e := newDependencyExecutioner(p.Porter, inst, opts)
	err := e.identifyDependencies(context.Background())
	require.NoError(t, err)
	require.Len(t, e.deps, 1)
	assert.Equal(t, "example.com/mysql:v1.2", e.deps[0].Reference,
		"upgrade with max-minor should pick highest matching version")
}
