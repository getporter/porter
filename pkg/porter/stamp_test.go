package porter

import (
	"context"
	"testing"

	"get.porter.sh/porter/pkg/linter"
	"get.porter.sh/porter/pkg/mixin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPorter_EnsureLocalBundleIsUpToDate_LintFailureHint(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Close()

	p.TestConfig.TestContext.AddTestFile("testdata/porter.yaml", "porter.yaml")

	mixins := p.Mixins.(*mixin.TestMixinProvider)
	mixins.LintResults = linter.Results{
		{Level: linter.LevelError},
	}

	var opts BuildOptions
	opts.File = "porter.yaml"
	require.NoError(t, opts.Validate(p.Porter))

	_, err := p.ensureLocalBundleIsUpToDate(context.Background(), opts)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Run 'porter build --no-lint'")
	assert.NotContains(t, err.Error(), "Rerun with --no-lint to ignore the errors",
		"the raw ErrLintFailed message must not be appended, or the unsupported hint reappears")
}
