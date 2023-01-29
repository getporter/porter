package linter

import (
	"context"
	"testing"

	"get.porter.sh/porter/pkg/manifest"
	"get.porter.sh/porter/pkg/mixin"
	"get.porter.sh/porter/pkg/portercontext"
	"github.com/stretchr/testify/require"
)

func TestLinter_Lint(t *testing.T) {
	ctx := context.Background()
	t.Run("no results", func(t *testing.T) {
		cxt := portercontext.NewTestContext(t)
		mixins := mixin.NewTestMixinProvider()
		l := New(cxt.Context, mixins)
		m := &manifest.Manifest{
			Mixins: []manifest.MixinDeclaration{
				{
					Name: "exec",
				},
			},
		}
		mixins.LintResults = nil

		results, err := l.Lint(ctx, m)
		require.NoError(t, err, "Lint failed")
		require.Len(t, results, 0, "linter should have returned 0 results")
	})

	t.Run("has results", func(t *testing.T) {
		cxt := portercontext.NewTestContext(t)
		mixins := mixin.NewTestMixinProvider()
		l := New(cxt.Context, mixins)
		m := &manifest.Manifest{
			Mixins: []manifest.MixinDeclaration{
				{
					Name: "exec",
				},
			},
		}
		mixins.LintResults = Results{
			{
				Level: LevelWarning,
				Code:  "exec-101",
				Title: "warning stuff isn't working",
			},
		}

		results, err := l.Lint(ctx, m)
		require.NoError(t, err, "Lint failed")
		require.Len(t, results, 1, "linter should have returned 1 result")
		require.Equal(t, mixins.LintResults, results, "unexpected lint results")
	})

	t.Run("mixin doesn't support lint", func(t *testing.T) {
		cxt := portercontext.NewTestContext(t)
		mixins := mixin.NewTestMixinProvider()
		l := New(cxt.Context, mixins)
		m := &manifest.Manifest{
			Mixins: []manifest.MixinDeclaration{
				{
					Name: "nope",
				},
			},
		}

		results, err := l.Lint(ctx, m)
		require.NoError(t, err, "Lint failed")
		require.Len(t, results, 0, "linter should ignore mixins that doesn't support the lint command")
	})

	testcases := []struct {
		Name			string
		ParameterName	string
	}{
		{
			Name: "does not use a reserved prefix",
			ParameterName: "porter-debug",
		},
		{
			Name: "is case insensitive and does not use reserved prefix even if mixed case ",
			ParameterName: "poRteR_lint",
		},
		{
			Name: "is case insensitive and does not use reserved prefix even if upper case ",
			ParameterName: "PORTER_DEBUG",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.Name, func(t *testing.T) {
			cxt := portercontext.NewTestContext(t)
			mixins := mixin.NewTestMixinProvider()
			l := New(cxt.Context, mixins)
			param := map[string]manifest.ParameterDefinition {
				"A": {
					Name: tc.ParameterName,
				},
			}
	
			m := &manifest.Manifest{
				Parameters: param,
			}
			mixins.LintResults = Results{
				{
					Level: LevelError,
					Location: Location {
						Action: "",
						Mixin: "",
						StepNumber: 0,
						StepDescription: "",
				
					},
					Code: "porter-100",
					Title: "Reserved name warning",
					Message: tc.ParameterName + " has a reserved prefix. Parameters cannot start with porter- or porter_",
					URL: "https://getporter.org/reference/linter/#porter-100",
				},
			}
	
			results, err := l.Lint(ctx, m)
			require.NoError(t, err, "Lint failed")
			require.Len(t, results, 1, "linter should have returned 1 result")
			require.Equal(t, mixins.LintResults, results, "unexpected lint results")
		})
	}

	t.Run("linter runs successfully if parameter does not use a reserved prefix", func(t *testing.T) {
		cxt := portercontext.NewTestContext(t)
		mixins := mixin.NewTestMixinProvider()
		l := New(cxt.Context, mixins)
		param := map[string]manifest.ParameterDefinition {
			"A": {
				Name: "successful",
			},
		}

		m := &manifest.Manifest{
			Parameters: param,
		}
		mixins.LintResults = Results{
			{
				Level: LevelError,
				Code:  "exec-101",
				Title: "warning stuff isn't working",
			},
		}

		results, err := l.Lint(ctx, m)
		require.NoError(t, err, "Lint failed")
		require.Len(t, results, 0, "linter should have returned 1 result")
	})
}
