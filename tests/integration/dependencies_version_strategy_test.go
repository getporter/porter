//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"get.porter.sh/porter/pkg"
	"get.porter.sh/porter/pkg/cnab"
	cnabtooci "get.porter.sh/porter/pkg/cnab/cnab-to-oci"
	depsv1ext "get.porter.sh/porter/pkg/cnab/extensions/dependencies/v1"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/porter"
	"get.porter.sh/porter/pkg/storage"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// depBundleJSON creates a minimal bundle.json with a v1 dependency that
// uses the given version ranges, writes it to the real filesystem, and
// returns its path.
func depBundleJSON(t *testing.T, dir, depRepo string, ranges []string) string {
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

	path := filepath.Join(dir, "bundle.json")
	require.NoError(t, os.WriteFile(path, data, pkg.FileModeWritable))
	return path
}

// mockTags returns a MockListTags function that always serves the given tags.
func mockTags(tags []string) func(context.Context, cnab.OCIReference, cnabtooci.RegistryOptions) ([]string, error) {
	return func(_ context.Context, _ cnab.OCIReference, _ cnabtooci.RegistryOptions) ([]string, error) {
		return tags, nil
	}
}

// TestInstall_DependencyVersionStrategy_ExactWithRangeRejects verifies
// that the default exact strategy rejects a bundle whose dependency
// specifies only a version range, through the full InstallBundle pipeline
// with a real config and real filesystem.
func TestInstall_DependencyVersionStrategy_ExactWithRangeRejects(t *testing.T) {
	t.Parallel()

	p := porter.NewTestPorter(t)
	ctx := p.SetupIntegrationTest()
	defer p.Close()

	// No config file — defaults to exact strategy.
	p.Registry = p.TestRegistry

	bundlePath := depBundleJSON(t, p.Getwd(), "example.com/mysql", []string{">=1.0 <1.3"})

	opts := porter.NewInstallOptions()
	opts.CNABFile = bundlePath
	require.NoError(t, opts.Validate(ctx, []string{}, p.Porter))

	err := p.InstallBundle(ctx, opts)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "version strategy is",
		"exact strategy + ranges-only dep must produce a clear version strategy error")
}

// TestInstall_DependencyVersionStrategy_MaxPatchSelectsHighest verifies
// that --dependencies-version-strategy=max-patch selects the highest
// matching semver tag from the registry during install, tested through
// the real InstallBundle pipeline and real filesystem.
func TestInstall_DependencyVersionStrategy_MaxPatchSelectsHighest(t *testing.T) {
	t.Parallel()

	p := porter.NewTestPorter(t)
	ctx := p.SetupIntegrationTest()
	defer p.Close()

	// Re-inject the mock registry so we can serve tags without a real OCI
	// registry. SetupIntegrationTest replaces p.Porter, so we must re-assign.
	p.Registry = p.TestRegistry
	p.TestRegistry.MockListTags = mockTags([]string{"v1.0", "v1.1", "v1.2"})

	// Capture which reference was pulled to verify the correct version.
	var pulledRef string
	p.TestRegistry.MockPullBundle = func(_ context.Context, ref cnab.OCIReference, _ cnabtooci.RegistryOptions) (cnab.BundleReference, error) {
		pulledRef = ref.String()
		return cnab.BundleReference{Reference: ref}, nil
	}

	bundlePath := depBundleJSON(t, p.Getwd(), "example.com/mysql", []string{">=1.0 <1.3"})

	opts := porter.NewInstallOptions()
	opts.CNABFile = bundlePath
	opts.DependenciesVersionStrategy = config.DependencyVersionStrategyMaxPatch
	require.NoError(t, opts.Validate(ctx, []string{}, p.Porter))

	// The install will fail after resolving deps (no real bundle to run),
	// but MockPullBundle is called during dep preparation so pulledRef is set.
	_ = p.InstallBundle(ctx, opts)

	assert.Equal(t, "example.com/mysql:v1.2", pulledRef,
		"max-patch should select the highest matching tag")
}

// TestUpgrade_DependencyVersionStrategy_MinSelectsLowest verifies that
// --dependencies-version-strategy=min selects the lowest matching semver
// tag from the registry during upgrade, tested through the real
// UpgradeBundle pipeline and real filesystem.
func TestUpgrade_DependencyVersionStrategy_MinSelectsLowest(t *testing.T) {
	t.Parallel()

	p := porter.NewTestPorter(t)
	ctx := p.SetupIntegrationTest()
	defer p.Close()

	p.Registry = p.TestRegistry
	p.TestRegistry.MockListTags = mockTags([]string{"v1.0", "v1.1", "v1.2"})

	var pulledRef string
	p.TestRegistry.MockPullBundle = func(_ context.Context, ref cnab.OCIReference, _ cnabtooci.RegistryOptions) (cnab.BundleReference, error) {
		pulledRef = ref.String()
		return cnab.BundleReference{Reference: ref}, nil
	}

	bundlePath := depBundleJSON(t, p.Getwd(), "example.com/mysql", []string{">=1.0 <1.3"})

	// UpgradeBundle requires an existing installation. Pre-create one marked as installed.
	now := time.Now()
	inst := storage.NewInstallation("", "testbundle")
	inst.Status.Installed = &now
	require.NoError(t, p.Installations.InsertInstallation(ctx, inst))

	opts := porter.NewUpgradeOptions()
	opts.Name = "testbundle"
	opts.CNABFile = bundlePath
	opts.DependenciesVersionStrategy = config.DependencyVersionStrategyMin
	require.NoError(t, opts.Validate(ctx, []string{}, p.Porter))

	// The upgrade will fail after resolving deps (no real bundle to run),
	// but MockPullBundle is called during dep preparation so pulledRef is set.
	_ = p.UpgradeBundle(ctx, opts)

	assert.Equal(t, "example.com/mysql:v1.0", pulledRef,
		"min strategy should select the lowest matching tag")
}
