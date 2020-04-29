package porter

import (
	"testing"

	"get.porter.sh/porter/pkg/manifest"
	"gopkg.in/AlecAivazis/survey.v1"
	"gopkg.in/AlecAivazis/survey.v1/core"
	"gopkg.in/AlecAivazis/survey.v1/terminal"

	"get.porter.sh/porter/pkg/secrets"

	"github.com/cnabio/cnab-go/credentials"
	"github.com/hinshun/vt10x"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPorter_applyDefaultOptions(t *testing.T) {
	p := NewTestPorter(t)
	p.TestConfig.SetupPorterHome()
	err := p.Create()
	require.NoError(t, err)

	opts := &InstallOptions{
		BundleLifecycleOpts{
			sharedOptions: sharedOptions{
				bundleFileOptions: bundleFileOptions{
					File: "porter.yaml",
				},
			},
		},
	}
	err = opts.Validate([]string{}, p.Context)
	require.NoError(t, err)

	p.Debug = true
	err = p.applyDefaultOptions(&opts.sharedOptions)
	require.NoError(t, err)

	assert.NotNil(t, p.Manifest, "Manifest should be loaded")
	assert.NotEqual(t, &manifest.Manifest{}, p.Manifest, "Manifest should not be empty")
	assert.Equal(t, p.Manifest.Name, opts.Name, "opts.Name should be set using the available manifest")

	debug, set := opts.combinedParameters["porter-debug"]
	assert.True(t, set)
	assert.Equal(t, "true", debug)
}

func TestPorter_applyDefaultOptions_NoManifest(t *testing.T) {
	p := NewTestPorter(t)

	opts := &InstallOptions{}
	err := opts.Validate([]string{}, p.Context)
	require.NoError(t, err)

	err = p.applyDefaultOptions(&opts.sharedOptions)
	require.NoError(t, err)

	assert.Equal(t, "", opts.Name, "opts.Name should be empty because the manifest was not available to default from")
	assert.Equal(t, &manifest.Manifest{}, p.Manifest, "p.Manifest should be initialized to an empty manifest")
}

func TestPorter_applyDefaultOptions_DebugOff(t *testing.T) {
	p := NewTestPorter(t)
	p.TestConfig.SetupPorterHome()
	err := p.Create()
	require.NoError(t, err)

	opts := &InstallOptions{}
	opts.File = "porter.yaml"
	err = opts.Validate([]string{}, p.Context)
	require.NoError(t, err)

	p.Debug = false
	err = p.applyDefaultOptions(&opts.sharedOptions)
	require.NoError(t, err)

	assert.Equal(t, p.Manifest.Name, opts.Name)

	_, set := opts.combinedParameters["porter-debug"]
	assert.False(t, set)
}

func TestPorter_applyDefaultOptions_ParamSet(t *testing.T) {
	p := NewTestPorter(t)
	p.TestConfig.SetupPorterHome()
	err := p.Create()
	require.NoError(t, err)

	opts := InstallOptions{}
	opts.Params = []string{"porter-debug=false"}

	err = opts.Validate([]string{}, p.Context)
	require.NoError(t, err)

	p.Debug = true
	err = p.applyDefaultOptions(&opts.sharedOptions)
	require.NoError(t, err)

	debug, set := opts.combinedParameters["porter-debug"]
	assert.True(t, set)
	assert.Equal(t, "false", debug)
}

func TestInstallOptions_validateParams(t *testing.T) {
	p := NewTestPorter(t)
	opts := InstallOptions{}
	opts.Params = []string{"A=1", "B=2"}

	err := opts.validateParams(p.Context)
	require.NoError(t, err)

	assert.Len(t, opts.Params, 2)
}

func TestInstallOptions_validateInstanceName(t *testing.T) {
	testcases := []struct {
		name      string
		args      []string
		wantClaim string
		wantError string
	}{
		{"none", nil, "", ""},
		{"name set", []string{"wordpress"}, "wordpress", ""},
		{"too many args", []string{"wordpress", "extra"}, "", "only one positional argument may be specified, the bundle instance name, but multiple were received: [wordpress extra]"},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			opts := InstallOptions{}
			err := opts.validateInstanceName(tc.args)

			if tc.wantError == "" {
				require.NoError(t, err)
				assert.Equal(t, tc.wantClaim, opts.Name)
			} else {
				require.EqualError(t, err, tc.wantError)
			}
		})
	}
}

func TestInstallOptions_combineParameters(t *testing.T) {
	p := NewTestPorter(t)
	p.FileSystem = &afero.Afero{Fs: afero.NewOsFs()}

	opts := InstallOptions{
		BundleLifecycleOpts{
			sharedOptions: sharedOptions{
				ParamFiles: []string{
					"testdata/install/base-params.txt",
					"testdata/install/dev-params.txt",
				},
				Params: []string{"A=true", "E=puppies", "E=kitties"},
			},
		},
	}

	err := opts.validateParams(p.Context)
	require.NoError(t, err)

	gotParams := opts.combineParameters()

	wantParams := map[string]string{
		"A": "true",
		"B": "2",
		"C": "3",
		"D": "blue",
		"E": "kitties",
	}

	assert.Equal(t, wantParams, gotParams)
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

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			opts := InstallOptions{
				BundleLifecycleOpts{
					sharedOptions: sharedOptions{
						Driver: tc.driver,
					},
				},
			}
			err := opts.validateDriver()

			if tc.wantError == "" {
				require.NoError(t, err)
				assert.Equal(t, tc.wantDriver, opts.Driver)
			} else {
				require.EqualError(t, err, tc.wantError)
			}
		})
	}
}

func TestPorter_InstallBundle_WithDepsFromTag(t *testing.T) {
	p := NewTestPorter(t)

	cacheDir, _ := p.Cache.GetCacheDir()
	p.TestConfig.TestContext.AddTestDirectory("testdata/cache", cacheDir)

	// Make some fake credentials to give to the install operation, they won't be used because it's a dummy driver
	cs := credentials.CredentialSet{
		Name: "wordpress",
		Credentials: []credentials.CredentialStrategy{
			{
				Name: "kubeconfig",
				Source: credentials.Source{
					Key:   secrets.SourceSecret,
					Value: "kubeconfig",
				},
			},
		},
	}
	p.TestCredentials.TestSecrets.AddSecret("kubeconfig", "abc123")
	err := p.Credentials.Save(cs)
	require.NoError(t, err, "Credentials.Save failed")

	opts := InstallOptions{}
	opts.Tag = "getporter/wordpress:v0.1.2"
	opts.CredentialIdentifiers = []string{"wordpress"}
	opts.Params = []string{"wordpress-password=mypassword"}
	err = opts.Validate(nil, p.Context)
	require.NoError(t, err, "Validate install options failed")

	err = p.InstallBundle(opts)
	require.NoError(t, err, "InstallBundle failed")
}

func TestPorter_InstallBundle_Interactive(t *testing.T) {
	p := NewTestPorter(t)
	p.TestConfig.TestContext.AddTestFile("testdata/bundle.json", "/bundle.json")
	p.TestCredentials.TestSecrets.AddSecret("my-first-cred", "my-first-cred-value")
	p.TestCredentials.TestSecrets.AddSecret("my-second-cred", "my-second-cred-value")

	core.DisableColor = true
	c, _, err := vt10x.NewVT10XConsole()
	defer c.Close()
	tstdio := terminal.Stdio{c.Tty(), c.Tty(), c.Tty()}
	p.SurveyAskOpts = survey.WithStdio(tstdio.In, tstdio.Out, tstdio.Err)

	opts := InstallOptions{}
	// opts.Tag = "getporter/wordpress:v0.1.2"
	opts.CNABFile = "/bundle.json"
	opts.Name = "HELLO_CUSTOM"
	// opts.CredentialIdentifiers = []string{"wordpress"}
	// opts.Params = []string{"wordpress-password=mypassword"}
	err = opts.Validate(nil, p.Context)
	require.NoError(t, err, "Validate install options failed")

	donec := make(chan struct{})
	go func() {
		defer close(donec)

		c.ExpectString("Choose an option")
		c.Send(string(terminal.KeyEnter))

		c.ExpectString("Enter credential identifier name")
		c.SendLine("credset_for_HELLO_CUSTOM")

		c.ExpectString("How would you like to set credential \"my-first-cred\"")
		c.Send(string(terminal.KeyArrowUp))
		c.Send(string(terminal.KeyArrowUp)) // secret
		c.Send(string(terminal.KeyEnter))
		c.ExpectString("Enter the secret that will be used to set credential \"my-first-cred\"")
		c.SendLine("my-first-cred")

		c.ExpectString("How would you like to set credential \"my-second-cred\"")
		c.Send(string(terminal.KeyArrowUp))
		c.Send(string(terminal.KeyArrowUp)) // secret
		c.Send(string(terminal.KeyEnter))

		c.ExpectString("Enter the secret that will be used to set credential \"my-second-cred\" ")
		c.SendLine("my-second-cred")

		c.ExpectEOF()
	}()

	err = p.InstallBundle(opts)

	c.Tty().Close()
	<-donec

	require.NoError(t, err, "InstallBundle failed")
}
