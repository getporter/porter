//go:build integration
// +build integration

package porter

import (
	"context"
	"path/filepath"
	"testing"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/storage"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-go/bundle/definition"
	"github.com/stretchr/testify/require"
)

var (
	kahnlatestHash = "fd4bbe38665531d10bb653140842a370"
)

func TestResolveBundleReference(t *testing.T) {
	t.Parallel()
	t.Run("current bundle source", func(t *testing.T) {
		t.Parallel()

		p := NewTestPorter(t)
		defer p.Close()

		p.AddTestBundleDir(filepath.Join(p.RepoRoot, "tests/testdata/mybuns"), true)

		opts := &BundleReferenceOptions{}
		require.NoError(t, opts.Validate(context.Background(), nil, p.Porter))
		ref, err := p.resolveBundleReference(context.Background(), opts)
		require.NoError(t, err)
		require.NotEmpty(t, opts.Name)
		require.NotEmpty(t, ref.Definition)
	})

	t.Run("cnab file", func(t *testing.T) {
		t.Parallel()

		p := NewTestPorter(t)
		defer p.Close()

		p.AddTestFile(filepath.Join(p.RepoRoot, "build/testdata/bundles/mysql/.cnab/bundle.json"), "bundle.json")

		opts := &BundleReferenceOptions{}
		opts.CNABFile = "bundle.json"
		require.NoError(t, opts.Validate(context.Background(), nil, p.Porter))
		ref, err := p.resolveBundleReference(context.Background(), opts)
		require.NoError(t, err)
		require.NotEmpty(t, opts.Name)
		require.NotEmpty(t, ref.Definition)
	})

	t.Run("reference", func(t *testing.T) {
		t.Parallel()

		p := NewTestPorter(t)
		defer p.Close()
		ctx := p.SetupIntegrationTest()

		opts := &BundleReferenceOptions{}
		opts.Reference = "ghcr.io/getporter/examples/porter-hello:v0.2.0"
		require.NoError(t, opts.Validate(ctx, nil, p.Porter))
		ref, err := p.resolveBundleReference(ctx, opts)
		require.NoError(t, err)
		require.NotEmpty(t, opts.Name)
		require.NotEmpty(t, ref.Definition)
		require.NotEmpty(t, ref.RelocationMap)
		require.NotEmpty(t, ref.Digest)
	})

	t.Run("installation name", func(t *testing.T) {
		t.Parallel()

		p := NewTestPorter(t)
		defer p.Close()

		i := p.TestInstallations.CreateInstallation(storage.NewInstallation("dev", "example"))
		p.TestInstallations.CreateRun(i.NewRun(cnab.ActionInstall), func(r *storage.Run) {
			r.BundleReference = kahnlatest.String()
			r.Bundle = buildExampleBundle()
			r.BundleDigest = kahnlatestHash
		})
		opts := &BundleReferenceOptions{}
		opts.Name = "example"
		opts.Namespace = "dev"
		require.NoError(t, opts.Validate(context.Background(), nil, p.Porter))
		ref, err := p.resolveBundleReference(context.Background(), opts)
		require.NoError(t, err)
		require.NotEmpty(t, opts.Name)
		require.NotEmpty(t, ref.Definition)
		require.NotEmpty(t, ref.Digest)
	})
}

func buildExampleBundle() bundle.Bundle {
	bun := bundle.Bundle{
		SchemaVersion:    bundle.GetDefaultSchemaVersion(),
		InvocationImages: []bundle.InvocationImage{{BaseImage: bundle.BaseImage{Image: "example.com/foo:v1.0.0"}}},
		Actions: map[string]bundle.Action{
			"blah": {
				Stateless: true,
			},
			"other": {
				Stateless: false,
				Modifies:  true,
			},
		},
		Definitions: map[string]*definition.Schema{
			"porter-debug-parameter": {
				Comment:     "porter-internal",
				ID:          "https://getporter.org/generated-bundle/#porter-debug",
				Default:     false,
				Description: "Print debug information from Porter when executing the bundle",
				Type:        "boolean",
			},
		},
		Parameters: map[string]bundle.Parameter{
			"porter-debug": {
				Definition:  "porter-debug-parameter",
				Description: "Print debug information from Porter when executing the bundle",
				Destination: &bundle.Location{
					EnvironmentVariable: "PORTER_DEUBG",
				},
			},
		},
	}
	return bun
}
