package porter

import (
	"context"
	"testing"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/pkg/secrets"
	"get.porter.sh/porter/pkg/storage"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-go/bundle/definition"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInstallOptions_validateInstallationName(t *testing.T) {
	testcases := []struct {
		name      string
		args      []string
		wantClaim string
		wantError string
	}{
		{"none", nil, "", ""},
		{"name set", []string{"wordpress"}, "wordpress", ""},
		{"too many args", []string{"wordpress", "extra"}, "", "only one positional argument may be specified, the installation name, but multiple were received: [wordpress extra]"},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			opts := NewInstallOptions()
			err := opts.validateInstallationName(tc.args)

			if tc.wantError == "" {
				require.NoError(t, err)
				assert.Equal(t, tc.wantClaim, opts.Name)
			} else {
				require.EqualError(t, err, tc.wantError)
			}
		})
	}
}

func TestInstallOptions_validateDriver(t *testing.T) {
	testcases := []struct {
		name       string
		driver     string
		wantDriver string
		wantError  string
	}{
		{"debug", "debug", DebugDriver, ""},
		{"docker", "docker", DockerDriver, ""},
		{"invalid driver provided", "dbeug", "", "unsupported driver or driver not found in PATH: dbeug"},
	}

	cxt := portercontext.NewTestContext(t)
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			opts := NewInstallOptions()
			opts.Driver = tc.driver
			err := opts.validateDriver(cxt.Context)

			if tc.wantError == "" {
				require.NoError(t, err)
				assert.Equal(t, tc.wantDriver, opts.Driver)
			} else {
				require.EqualError(t, err, tc.wantError)
			}
		})
	}
}

func TestPorter_applyActionOptionsToInstallation(t *testing.T) {
	setup := func() (context.Context, *TestPorter, *storage.Installation) {
		ctx := context.Background()
		p := NewTestPorter(t)

		p.TestParameters.InsertParameterSet(ctx, storage.ParameterSet{
			ParameterSetSpec: storage.ParameterSetSpec{
				Name: "newps1",
				Parameters: []secrets.Strategy{
					{Name: "logLevel", Source: secrets.Source{Key: "value", Value: "11"}},
				},
			},
		})

		return ctx, p, &storage.Installation{
			Bundle: storage.OCIReferenceParts{
				Repository: "example.com/mybuns",
				Version:    "1.0.0",
			},
			ParameterSets:  []string{"oldps1"},
			CredentialSets: []string{"oldcs1", "oldcs2"},
		}
	}

	t.Run("replace previous sets", func(t *testing.T) {
		ctx, p, inst := setup()

		// We should replace the previously used sets since we specified different ones
		opts := NewBundleExecutionOptions()
		opts.Reference = kahnlatest.String()
		opts.bundleRef = &cnab.BundleReference{
			Reference: kahnlatest,
			Definition: cnab.NewBundle(bundle.Bundle{
				Credentials: map[string]bundle.Credential{
					"userid": {},
				},
				Parameters: map[string]bundle.Parameter{
					"logLevel": {Definition: "logLevel"},
				},
				Definitions: map[string]*definition.Schema{
					"logLevel": {Type: "string"},
				},
			}),
		}

		opts.ParameterSets = []string{"newps1"}
		opts.CredentialIdentifiers = []string{"newcs1"}

		require.NoError(t, opts.Validate(ctx, nil, p.Porter))
		err := p.applyActionOptionsToInstallation(ctx, inst, opts)
		require.NoError(t, err, "applyActionOptionsToInstallation failed")

		require.Equal(t, opts.ParameterSets, inst.ParameterSets, "expected the installation to replace the credential sets with those specified")
		require.Equal(t, opts.CredentialIdentifiers, inst.CredentialSets, "expected the installation to replace the credential sets with those specified")
	})

	t.Run("reuse previous sets", func(t *testing.T) {
		ctx, p, inst := setup()

		// We should reuse the previously used sets since we specified different ones
		opts := NewBundleExecutionOptions()
		opts.Reference = kahnlatest.String()
		opts.bundleRef = &cnab.BundleReference{
			Reference: kahnlatest,
			Definition: cnab.NewBundle(bundle.Bundle{
				Credentials: map[string]bundle.Credential{
					"userid": {},
				},
				Parameters: map[string]bundle.Parameter{
					"logLevel": {Definition: "logLevel"},
				},
				Definitions: map[string]*definition.Schema{
					"logLevel": {Type: "string"},
				},
			}),
		}
		opts.ParameterSets = []string{}
		opts.CredentialIdentifiers = []string{}

		require.NoError(t, opts.Validate(ctx, nil, p.Porter))
		err := p.applyActionOptionsToInstallation(ctx, inst, opts)
		require.NoError(t, err, "applyActionOptionsToInstallation failed")

		require.Equal(t, []string{"oldps1"}, inst.ParameterSets, "expected the installation to reuse the previous credential sets")
		require.Equal(t, []string{"oldcs1", "oldcs2"}, inst.CredentialSets, "expected the installation to reuse the previous credential sets")
	})
}
