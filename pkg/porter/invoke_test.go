package porter

import (
	"testing"

	"github.com/hinshun/vt10x"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/AlecAivazis/survey.v1"
	"gopkg.in/AlecAivazis/survey.v1/terminal"
)

func TestInvokeOptions_Validate_ActionRequired(t *testing.T) {
	p := NewTestPorter(t)
	opts := InvokeOptions{}

	err := opts.Validate(nil, p.Context)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "--action is required")
}

func TestPorter_Invoke_ChooseCred(t *testing.T) {
	p := NewTestPorter(t)
	p.TestConfig.TestContext.AddTestFile("testdata/bundle.json", "/bundle.json")
	p.TestCredentials.AddTestCredentialsDirectory("testdata/test-creds")
	p.TestCredentials.TestSecrets.AddSecret("my-first-cred", "my-first-cred-value")
	p.TestCredentials.TestSecrets.AddSecret("my-second-cred", "my-second-cred-value")

	c, _, _ := vt10x.NewVT10XConsole()
	defer c.Close()
	tstdio := terminal.Stdio{c.Tty(), c.Tty(), c.Tty()}
	p.SurveyAskOpts = survey.WithStdio(tstdio.In, tstdio.Out, tstdio.Err)

	installOpts := InstallOptions{}
	installOpts.CNABFile = "/bundle.json"
	installOpts.Name = "HELLO_CUSTOM"
	installOpts.CredentialIdentifiers = []string{"cred_set_HELLO_CUSTOM"}

	err := p.InstallBundle(installOpts)

	require.NoError(t, err, "InstallBundle failed")

	invokeOpts := InvokeOptions{}
	invokeOpts.CNABFile = "/bundle.json"
	invokeOpts.Name = "HELLO_CUSTOM"
	invokeOpts.Action = "zombies"

	donec := make(chan struct{})
	go func() {
		defer close(donec)

		c.ExpectString("Choose an option")
		c.Send(string(terminal.KeyEnter)) // select "choose credential set"

		c.ExpectString("Choose a set of credentials to use while installing this bundle")
		c.Send("cred_set_HELLO_CUSTOM")   // search
		c.Send(string(terminal.KeySpace)) // select
		c.Send(string(terminal.KeyEnter))

		c.ExpectEOF()
	}()

	err = p.InvokeBundle(invokeOpts)

	c.Tty().Close()
	<-donec

	require.NoError(t, err, "action invocation failed")
}
