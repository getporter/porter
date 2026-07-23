package porter

import (
	"context"
	"errors"
	"testing"

	"get.porter.sh/porter/pkg/cnab"
	cnabtooci "get.porter.sh/porter/pkg/cnab/cnab-to-oci"
	v2 "get.porter.sh/porter/pkg/cnab/extensions/dependencies/v2"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/stretchr/testify/require"
)

// failIfPulled returns a MockPullBundle that fails the test immediately if
// invoked, for asserting that composeRequiredInterface didn't attempt a
// pull in cases where it shouldn't need to.
func failIfPulled(t *testing.T) func(context.Context, cnab.OCIReference, cnabtooci.RegistryOptions) (cnab.BundleReference, error) {
	t.Helper()
	return func(_ context.Context, ref cnab.OCIReference, _ cnabtooci.RegistryOptions) (cnab.BundleReference, error) {
		t.Fatalf("unexpected pull of %s", ref.String())
		return cnab.BundleReference{}, nil
	}
}

func TestGraphBuilder_ComposeRequiredInterface(t *testing.T) {
	t.Parallel()

	const dbRef = "localhost:5000/mysql:v1.0.0"

	baseDep := func(transform func(d *v2.Dependency)) v2.Dependency {
		d := v2.Dependency{Bundle: dbRef, Outputs: map[string]string{"connstr": "${outputs.connstr}"}}
		if transform != nil {
			transform(&d)
		}
		return d
	}

	t.Run("nil interface: requirement is exactly requiredOutputNames, no pull attempted", func(t *testing.T) {
		t.Parallel()

		p := NewTestPorter(t)
		defer p.Close()
		p.TestRegistry.MockPullBundle = failIfPulled(t)

		builder := NewGraphBuilder(p.Porter, 10)
		got, err := builder.composeRequiredInterface(context.Background(), "db", baseDep(nil), nil, ExplainOpts{})
		require.NoError(t, err)
		require.Equal(t, cnab.InterfaceRequirement{Outputs: []string{"connstr"}}, got)
	})

	t.Run("document only: unions document names, no pull attempted", func(t *testing.T) {
		t.Parallel()

		p := NewTestPorter(t)
		defer p.Close()
		p.TestRegistry.MockPullBundle = failIfPulled(t)

		dep := baseDep(func(d *v2.Dependency) {
			d.Interface = &v2.DependencyInterface{Document: v2.DependencyInterfaceDocument{
				Outputs:     map[string]bundle.Output{"port": {}},
				Parameters:  map[string]bundle.Parameter{"logLevel": {}},
				Credentials: map[string]bundle.Credential{"token": {}},
			}}
		})

		builder := NewGraphBuilder(p.Porter, 10)
		got, err := builder.composeRequiredInterface(context.Background(), "db", dep, nil, ExplainOpts{})
		require.NoError(t, err)
		require.Equal(t, cnab.InterfaceRequirement{
			Outputs:     []string{"connstr", "port"},
			Parameters:  []string{"logLevel"},
			Credentials: []string{"token"},
		}, got)
	})

	t.Run("reference only: pulls the referenced bundle and unions its names", func(t *testing.T) {
		t.Parallel()

		const interfaceRef = "localhost:5000/mysql-interface:v1.0.0"

		p := NewTestPorter(t)
		defer p.Close()
		p.TestRegistry.MockPullBundle = newMockPullBundle(map[string]cnab.ExtendedBundle{
			interfaceRef: {Bundle: bundle.Bundle{
				Name:    "mysql-interface",
				Version: "1.0.0",
				Outputs: map[string]bundle.Output{"port": {}},
			}},
		})

		dep := baseDep(func(d *v2.Dependency) {
			d.Interface = &v2.DependencyInterface{Reference: interfaceRef}
		})

		builder := NewGraphBuilder(p.Porter, 10)
		got, err := builder.composeRequiredInterface(context.Background(), "db", dep, nil, ExplainOpts{})
		require.NoError(t, err)
		require.Equal(t, cnab.InterfaceRequirement{
			Outputs: []string{"connstr", "port"},
		}, got)
	})

	t.Run("reference and document both set: returns the sentinel error, no pull attempted", func(t *testing.T) {
		t.Parallel()

		p := NewTestPorter(t)
		defer p.Close()
		p.TestRegistry.MockPullBundle = failIfPulled(t)

		dep := baseDep(func(d *v2.Dependency) {
			d.Interface = &v2.DependencyInterface{
				Reference: "localhost:5000/mysql-interface:v1.0.0",
				Document:  v2.DependencyInterfaceDocument{Outputs: map[string]bundle.Output{"port": {}}},
			}
		})

		builder := NewGraphBuilder(p.Porter, 10)
		_, err := builder.composeRequiredInterface(context.Background(), "db", dep, nil, ExplainOpts{})
		require.ErrorIs(t, err, errInterfaceReferenceAndDocument)
	})

	t.Run("ID only: requirement is exactly requiredOutputNames, no pull attempted", func(t *testing.T) {
		t.Parallel()

		p := NewTestPorter(t)
		defer p.Close()
		p.TestRegistry.MockPullBundle = failIfPulled(t)

		dep := baseDep(func(d *v2.Dependency) {
			d.Interface = &v2.DependencyInterface{ID: "mysql"}
		})

		builder := NewGraphBuilder(p.Porter, 10)
		got, err := builder.composeRequiredInterface(context.Background(), "db", dep, nil, ExplainOpts{})
		require.NoError(t, err)
		require.Equal(t, cnab.InterfaceRequirement{Outputs: []string{"connstr"}}, got)
	})

	t.Run("reference pull failure propagates as a plain error, not the sentinel", func(t *testing.T) {
		t.Parallel()

		p := NewTestPorter(t)
		defer p.Close()
		p.TestRegistry.MockPullBundle = newMockPullBundle(map[string]cnab.ExtendedBundle{})

		dep := baseDep(func(d *v2.Dependency) {
			d.Interface = &v2.DependencyInterface{Reference: "localhost:5000/missing-interface:v1.0.0"}
		})

		builder := NewGraphBuilder(p.Porter, 10)
		_, err := builder.composeRequiredInterface(context.Background(), "db", dep, nil, ExplainOpts{})
		require.Error(t, err)
		require.False(t, errors.Is(err, errInterfaceReferenceAndDocument))
	})
}
