package porter

import (
	"context"
	"testing"

	"get.porter.sh/porter/pkg/claims"
	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/portercontext"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogsShowOptions_Validate(t *testing.T) {

	t.Run("installation specified", func(t *testing.T) {
		c := portercontext.NewTestContext(t)
		opts := LogsShowOptions{}
		opts.Name = "mybun"

		err := opts.Validate(c.Context)
		require.NoError(t, err)
	})

	t.Run("installation defaulted", func(t *testing.T) {
		c := portercontext.NewTestContext(t)
		c.AddTestFile("testdata/porter.yaml", "porter.yaml")

		opts := LogsShowOptions{}

		err := opts.Validate(c.Context)
		require.NoError(t, err)
		assert.NotEmpty(t, opts.File) // it should pick up that there is one present, the name is defaulted when the action is run just like install
	})

	t.Run("run specified", func(t *testing.T) {
		c := portercontext.NewTestContext(t)
		opts := LogsShowOptions{}
		opts.ClaimID = "abc123"

		err := opts.Validate(c.Context)
		require.NoError(t, err)
	})

	t.Run("both specified", func(t *testing.T) {
		c := portercontext.NewTestContext(t)
		opts := LogsShowOptions{}
		opts.Name = "mybun"
		opts.ClaimID = "abc123"

		err := opts.Validate(c.Context)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "either --installation or --run should be specified, not both")
	})

	t.Run("neither specified", func(t *testing.T) {
		c := portercontext.NewTestContext(t)
		opts := LogsShowOptions{}

		err := opts.Validate(c.Context)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "either --installation or --run is required")
	})
}

func TestPorter_ShowInstallationLogs(t *testing.T) {
	t.Run("no logs found", func(t *testing.T) {
		p := NewTestPorter(t)
		defer p.Close()

		i := p.TestClaims.CreateInstallation(claims.NewInstallation("", "test"))
		c := p.TestClaims.CreateRun(i.NewRun(cnab.ActionInstall))
		p.TestClaims.CreateResult(c.NewResult(cnab.StatusSucceeded))

		var opts LogsShowOptions
		opts.Name = "test"
		err := p.ShowInstallationLogs(context.Background(), &opts)
		require.Error(t, err, "ShowInstallationLogs should have failed")
		assert.Contains(t, err.Error(), "no logs found")
	})

	t.Run("has logs", func(t *testing.T) {
		const testLogs = "some mighty fine logs"

		p := NewTestPorter(t)
		defer p.Close()

		i := p.TestClaims.CreateInstallation(claims.NewInstallation("", "test"))
		c := p.TestClaims.CreateRun(i.NewRun(cnab.ActionInstall))
		r := p.TestClaims.CreateResult(c.NewResult(cnab.StatusSucceeded), func(r *claims.Result) {
			r.OutputMetadata.SetGeneratedByBundle(cnab.OutputInvocationImageLogs, false)
		})

		p.TestClaims.CreateOutput(r.NewOutput(cnab.OutputInvocationImageLogs, []byte(testLogs)))

		var opts LogsShowOptions
		opts.Name = "test"
		err := p.ShowInstallationLogs(context.Background(), &opts)
		require.NoError(t, err, "ShowInstallationLogs failed")

		assert.Contains(t, p.TestConfig.TestContext.GetOutput(), testLogs)
	})
}
