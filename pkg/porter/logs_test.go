package porter

import (
	"testing"

	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-go/claim"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPorter_ShowInstallationLogs(t *testing.T) {
	t.Run("no logs found", func(t *testing.T) {
		p := NewTestPorter(t)
		b := bundle.Bundle{}
		c := p.TestClaims.CreateClaim("test", claim.ActionInstall, b, nil)
		p.TestClaims.CreateResult(c, claim.StatusSucceeded)

		var opts LogsShowOptions
		opts.Name = "test"
		err := p.ShowInstallationLogs(&opts)
		require.Error(t, err, "ShowInstallationLogs should have failed")
		assert.Contains(t, err.Error(), "no logs found")
	})

	t.Run("has logs", func(t *testing.T) {
		const testLogs = "some mighty fine logs"

		p := NewTestPorter(t)
		b := bundle.Bundle{}
		c := p.TestClaims.CreateClaim("test", claim.ActionInstall, b, nil)
		r := p.TestClaims.CreateResult(c, claim.StatusSucceeded)
		p.TestClaims.CreateOutput(c, r, claim.OutputInvocationImageLogs, []byte(testLogs))
		r.OutputMetadata.SetGeneratedByBundle(claim.OutputInvocationImageLogs, false)
		p.TestClaims.SaveResult(r)

		var opts LogsShowOptions
		opts.Name = "test"
		err := p.ShowInstallationLogs(&opts)
		require.NoError(t, err, "ShowInstallationLogs failed")

		assert.Contains(t, p.TestConfig.TestContext.GetOutput(), testLogs)
	})
}
