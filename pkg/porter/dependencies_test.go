package porter

import (
	"context"
	"encoding/json"
	"testing"

	"get.porter.sh/porter/pkg/cnab"
	cnabtooci "get.porter.sh/porter/pkg/cnab/cnab-to-oci"
	depsv1ext "get.porter.sh/porter/pkg/cnab/extensions/dependencies/v1"
	v2ext "get.porter.sh/porter/pkg/cnab/extensions/dependencies/v2"
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

// bundleWithV2DanglingWiring marshals a minimal bundle.Bundle that declares
// a single v2 dependency whose credentials mapping references a sibling
// alias that doesn't exist, into JSON.
func bundleWithV2DanglingWiring(t *testing.T, depRepo string) []byte {
	t.Helper()
	deps := v2ext.Dependencies{
		Requires: map[string]v2ext.Dependency{
			"app": {
				Bundle: depRepo,
				Credentials: map[string]string{
					"conn": "${bundle.dependencies.doesnotexist.outputs.foo}",
				},
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
		RequiredExtensions: []string{cnab.DependenciesV2ExtensionKey},
		Custom: map[string]interface{}{
			cnab.DependenciesV2ExtensionKey: deps,
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
	p.Data.Dependencies.VersionStrategy = config.DependencyVersionStrategyMin
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
	p.Data.Dependencies.VersionStrategy = config.DependencyVersionStrategyMaxMinor
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

func TestIdentifyDependencies_MaxPatchRestrictsToPatchLevel(t *testing.T) {
	t.Parallel()

	p := NewTestPorter(t)
	defer p.Close()
	// v1.2.4 and v1.2.5 are patch upgrades of v1.2.3; v1.3.0 and v2.0.0 are not.
	p.TestRegistry.MockListTags = staticTags([]string{"v1.2.4", "v1.2.5", "v1.3.0", "v2.0.0"})

	// Default ref includes the version tag; range is broad (>=1.0.0).
	bunData := bundleWithV1Ranges(t, "example.com/mysql:v1.2.3", []string{">=1.0.0"})
	require.NoError(t, p.FileSystem.WriteFile("/bundle.json", bunData, 0600))

	opts := NewInstallOptions()
	opts.DependenciesVersionStrategy = config.DependencyVersionStrategyMaxPatch

	e := newExecWithCNABFile(p, "/bundle.json", &opts)
	err := e.identifyDependencies(context.Background())
	require.NoError(t, err)
	require.Len(t, e.deps, 1)
	assert.Equal(t, "example.com/mysql:v1.2.5", e.deps[0].Reference,
		"max-patch should stay within the same major.minor as the default version")
}

func TestIdentifyDependencies_MaxMinorRestrictsToMinorLevel(t *testing.T) {
	t.Parallel()

	p := NewTestPorter(t)
	defer p.Close()
	// v1.3.0 is a minor upgrade of v1.2.3; v2.0.0 is a major upgrade.
	p.TestRegistry.MockListTags = staticTags([]string{"v1.2.4", "v1.3.0", "v2.0.0"})

	bunData := bundleWithV1Ranges(t, "example.com/mysql:v1.2.3", []string{">=1.0.0"})
	require.NoError(t, p.FileSystem.WriteFile("/bundle.json", bunData, 0600))

	opts := NewInstallOptions()
	opts.DependenciesVersionStrategy = config.DependencyVersionStrategyMaxMinor

	e := newExecWithCNABFile(p, "/bundle.json", &opts)
	err := e.identifyDependencies(context.Background())
	require.NoError(t, err)
	require.Len(t, e.deps, 1)
	assert.Equal(t, "example.com/mysql:v1.3.0", e.deps[0].Reference,
		"max-minor should stay within the same major as the default version")
}

// TestIdentifyDependencies_V2DanglingWiringFailsBeforePull confirms
// identifyDependencies -- the first step of Prepare(), itself the first step
// of ExecuteAction(), shared by install/upgrade/invoke/reconcile -- hard
// fails a v2 bundle with an unresolvable wiring reference before e.deps is
// ever populated, i.e. before any dependency (or the root bundle) could run.
func TestIdentifyDependencies_V2DanglingWiringFailsBeforePull(t *testing.T) {
	t.Parallel()

	p := NewTestPorter(t)
	defer p.Close()
	p.TestRegistry.MockPullBundle = newMockPullBundle(map[string]cnab.ExtendedBundle{
		// intentionally empty: the dangling wiring reference must be caught
		// before any dependency bundle needs to be pulled.
	})

	bunData := bundleWithV2DanglingWiring(t, "example.com/myapp:v1.0.0")
	require.NoError(t, p.FileSystem.WriteFile("/bundle.json", bunData, 0600))

	opts := NewInstallOptions()

	e := newExecWithCNABFile(p, "/bundle.json", &opts)
	err := e.identifyDependencies(context.Background())
	require.Error(t, err)
	var wiringErr ErrDanglingWiringReference
	require.ErrorAs(t, err, &wiringErr)
	assert.Equal(t, "app", wiringErr.DependencyAlias)
	assert.Nil(t, e.deps, "deps must not be populated when wiring validation fails")
}

// newDependencyExecutionerFor builds a dependencyExecutioner around parent
// without going through Prepare(), which requires pulling a real bundle.
// runDependencyv2/finalizeExecute only read e.parentArgs.Installation,
// e.parentOpts.Namespace and e.parentAction, so those are set directly.
func newDependencyExecutionerFor(p *TestPorter, parent storage.Installation, action BundleAction) *dependencyExecutioner {
	e := newDependencyExecutioner(p.Porter, parent, action)
	e.parentArgs.Installation = parent
	return e
}

// TestRunDependencyV2_AddsReferenceOnCreate confirms that creating a new v2
// (SharingMode) dependency installation records the parent's reference to
// it -- issue #2608, goal 1 (create).
func TestRunDependencyV2_AddsReferenceOnCreate(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := NewTestPorter(t)
	defer p.Close()

	parent := storage.NewInstallation("", "myinfra")
	opts := NewInstallOptions()
	opts.Namespace = ""
	opts.Name = "myinfra"

	e := newDependencyExecutionerFor(p, parent, opts)
	dep := &queuedDependency{
		DependencyLock: cnab.DependencyLock{
			Alias:        "db",
			SharingMode:  true,
			SharingGroup: "group1",
		},
	}

	err := e.runDependencyv2(ctx, dep)
	require.NoError(t, err)

	depInstallation, err := p.Installations.GetInstallation(ctx, "", "db")
	require.NoError(t, err)
	assert.Equal(t,
		[]storage.InstallationReference{{Installation: "/myinfra", Dependency: "db"}},
		depInstallation.Status.References,
		"creating a v2 dependency should record a reference from its parent")
}

// TestRunDependencyV2_AddsReferenceOnReuse confirms that reusing an
// already-installed v2 dependency installation to satisfy an upgrade
// records the parent's reference to it too -- issue #2608, goal 1 (reuse),
// goal 2 (finding a parent's dependencies on upgrade).
func TestRunDependencyV2_AddsReferenceOnReuse(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := NewTestPorter(t)
	defer p.Close()

	p.TestInstallations.CreateInstallation(storage.NewInstallation("", "db"), func(i *storage.Installation) {
		i.SetLabel(sharingGroupLabel, "group1")
		i.Status.Installed = &now
	})

	parent := storage.NewInstallation("", "webapp")
	// GetAction() == "upgrade" makes runDependencyv2 return right after
	// recording the reference (matching sharing group + already installed),
	// without needing to actually execute the dependency's bundle.
	opts := NewUpgradeOptions()
	opts.Namespace = ""
	opts.Name = "webapp"

	e := newDependencyExecutionerFor(p, parent, opts)
	dep := &queuedDependency{
		DependencyLock: cnab.DependencyLock{
			Alias:        "db",
			SharingMode:  true,
			SharingGroup: "group1",
		},
	}

	err := e.runDependencyv2(ctx, dep)
	require.NoError(t, err)

	depInstallation, err := p.Installations.GetInstallation(ctx, "", "db")
	require.NoError(t, err)
	assert.Equal(t,
		[]storage.InstallationReference{{Installation: "/webapp", Dependency: "db"}},
		depInstallation.Status.References,
		"reusing a v2 dependency should record a reference from the installation reusing it")
}

// TestRunDependencyV2_Uninstall_DropsReferenceButNeverDeletes confirms that
// uninstalling a parent installation only ever drops its own reference to a
// v2 (SharingMode) dependency -- it never runs the dependency's own
// uninstall action or deletes its installation record, regardless of
// whether any other installation still references it, since a referencing
// installation never owns the shared dependency's lifecycle (it may not
// have been the one that originally installed it). Issue #2608, goal 1.
func TestRunDependencyV2_Uninstall_DropsReferenceButNeverDeletes(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name            string
		otherReferences bool
	}{
		{name: "no other references after dropping the parent's own"},
		{name: "still referenced by another installation", otherReferences: true},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			p := NewTestPorter(t)
			defer p.Close()

			p.TestInstallations.CreateInstallation(storage.NewInstallation("", "db"), func(i *storage.Installation) {
				i.SetLabel(sharingGroupLabel, "group1")
				i.Status.Installed = &now
				i.AddReference("/myinfra", "db")
				if tc.otherReferences {
					i.AddReference("/webapp", "db")
				}
			})

			parent := storage.NewInstallation("", "myinfra")
			opts := NewUninstallOptions()
			opts.Namespace = ""
			opts.Name = "myinfra"

			e := newDependencyExecutionerFor(p, parent, opts)
			dep := &queuedDependency{
				DependencyLock: cnab.DependencyLock{
					Alias:        "db",
					SharingMode:  true,
					SharingGroup: "group1",
				},
			}

			err := e.runDependencyv2(ctx, dep)
			require.NoError(t, err)

			depInstallation, err := p.Installations.GetInstallation(ctx, "", "db")
			require.NoError(t, err, "the shared dependency must never be deleted by a referencing installation's uninstall")
			assert.True(t, depInstallation.IsInstalled(), "the shared dependency must never be uninstalled by a referencing installation's uninstall")
			assert.NotContains(t, depInstallation.Status.References, storage.InstallationReference{Installation: "/myinfra", Dependency: "db"},
				"the parent's own reference should be dropped")

			wantReferences := []storage.InstallationReference(nil)
			if tc.otherReferences {
				wantReferences = []storage.InstallationReference{{Installation: "/webapp", Dependency: "db"}}
			}
			assert.Equal(t, wantReferences, depInstallation.Status.References)
		})
	}
}

// TestRunDependencyV2_Uninstall_NotFoundIsANoOp confirms uninstalling a
// parent whose v2 dependency installation doesn't exist (or was already
// deleted) doesn't create one just to immediately do nothing with it.
func TestRunDependencyV2_Uninstall_NotFoundIsANoOp(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := NewTestPorter(t)
	defer p.Close()

	parent := storage.NewInstallation("", "myinfra")
	opts := NewUninstallOptions()
	opts.Namespace = ""
	opts.Name = "myinfra"

	e := newDependencyExecutionerFor(p, parent, opts)
	dep := &queuedDependency{
		DependencyLock: cnab.DependencyLock{
			Alias:       "db",
			SharingMode: true,
		},
	}

	err := e.runDependencyv2(ctx, dep)
	require.NoError(t, err)

	_, err = p.Installations.GetInstallation(ctx, "", "db")
	require.ErrorIs(t, err, storage.ErrNotFound{}, "uninstalling should not create a dependency installation record that never existed")
}

// TestExecute_V2Install_ReusingSharedDependency_RecordsReference confirms
// that Execute -- the actual production entry point for running
// dependencies (see action.go, uninstall.go), never runDependencyv2 called
// directly -- still records a reference when installing a new parent that
// reuses an already-installed shared v2 dependency. sharedActionResolver
// makes Execute return before ever reaching runDependencyv2 in this case,
// since the dependency's own install action must not be re-run.
func TestExecute_V2Install_ReusingSharedDependency_RecordsReference(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := NewTestPorter(t)
	defer p.Close()

	p.TestInstallations.CreateInstallation(storage.NewInstallation("", "db"), func(i *storage.Installation) {
		i.SetLabel(sharingGroupLabel, "group1")
		i.Status.Installed = &now
	})

	parent := storage.NewInstallation("", "webapp")
	opts := NewInstallOptions()
	opts.Namespace = ""
	opts.Name = "webapp"

	e := newDependencyExecutionerFor(p, parent, opts)
	e.deps = []*queuedDependency{
		{
			DependencyLock: cnab.DependencyLock{
				Alias:        "db",
				SharingMode:  true,
				SharingGroup: "group1",
			},
		},
	}

	err := e.Execute(ctx)
	require.NoError(t, err)

	depInstallation, err := p.Installations.GetInstallation(ctx, "", "db")
	require.NoError(t, err)
	assert.Equal(t,
		[]storage.InstallationReference{{Installation: "/webapp", Dependency: "db"}},
		depInstallation.Status.References,
		"installing a parent that reuses an existing shared dependency should record a reference")
}

// TestExecute_V2Uninstall_ReleasingSharedDependency_DropsReference confirms
// that Execute drops a parent's reference to a shared v2 dependency on
// uninstall, even though sharedActionResolver returns before
// runDependencyv2 is ever reached (this installation never owns the shared
// dependency's lifecycle).
func TestExecute_V2Uninstall_ReleasingSharedDependency_DropsReference(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := NewTestPorter(t)
	defer p.Close()

	p.TestInstallations.CreateInstallation(storage.NewInstallation("", "db"), func(i *storage.Installation) {
		i.SetLabel(sharingGroupLabel, "group1")
		i.Status.Installed = &now
		i.AddReference("/webapp", "db")
	})

	parent := storage.NewInstallation("", "webapp")
	opts := NewUninstallOptions()
	opts.Namespace = ""
	opts.Name = "webapp"

	e := newDependencyExecutionerFor(p, parent, opts)
	e.deps = []*queuedDependency{
		{
			DependencyLock: cnab.DependencyLock{
				Alias:        "db",
				SharingMode:  true,
				SharingGroup: "group1",
			},
		},
	}

	err := e.Execute(ctx)
	require.NoError(t, err)

	depInstallation, err := p.Installations.GetInstallation(ctx, "", "db")
	require.NoError(t, err)
	assert.True(t, depInstallation.IsInstalled(), "uninstalling a parent must never uninstall a shared dependency it merely references")
	assert.Empty(t, depInstallation.Status.References, "uninstalling a parent should drop its reference to a shared dependency")
}

// TestFinalizeExecute_UnconditionalDeleteForV1 confirms finalizeExecute
// (only ever reached by v1 dependencies -- see runDependencyv2, which
// handles v2/SharingMode uninstalls itself before finalizeExecute is
// called) keeps its original unconditional-delete behavior on
// uninstall --delete. v1 dependency names are scoped to a single parent
// (BuildPrerequisiteInstallationName) and can never be shared, so this is
// always safe.
func TestFinalizeExecute_UnconditionalDeleteForV1(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := NewTestPorter(t)
	defer p.Close()

	depInstallation := p.TestInstallations.CreateInstallation(storage.NewInstallation("", "myinfra-db"))

	parent := storage.NewInstallation("", "myinfra")
	opts := NewUninstallOptions()
	opts.Namespace = ""
	opts.Name = "myinfra"
	opts.ForceDelete = true // swallow the CNAB execute error from the unconfigured bundle below

	e := newDependencyExecutionerFor(p, parent, opts)
	e.depArgs.Installation = depInstallation

	dep := &queuedDependency{
		DependencyLock: cnab.DependencyLock{
			Alias:       "db",
			SharingMode: false,
		},
	}

	err := e.finalizeExecute(ctx, dep)
	require.NoError(t, err)

	_, err = p.Installations.GetInstallation(ctx, "", "myinfra-db")
	require.ErrorIs(t, err, storage.ErrNotFound{}, "expected the v1 dependency installation to be deleted unconditionally")
}
