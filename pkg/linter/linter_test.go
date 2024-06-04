package linter

import (
	"context"
	"fmt"
	"testing"

	"get.porter.sh/porter/pkg/manifest"
	"get.porter.sh/porter/pkg/mixin"
	"get.porter.sh/porter/pkg/pkgmgmt"
	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/tests"
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
		Name          string
		ParameterName string
	}{
		{
			Name:          "does not use a reserved prefix",
			ParameterName: "porter-debug",
		},
		{
			Name:          "is case insensitive and does not use reserved prefix even if mixed case ",
			ParameterName: "poRteR_lint",
		},
		{
			Name:          "is case insensitive and does not use reserved prefix even if upper case ",
			ParameterName: "PORTER_DEBUG",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.Name, func(t *testing.T) {
			cxt := portercontext.NewTestContext(t)
			mixins := mixin.NewTestMixinProvider()
			l := New(cxt.Context, mixins)
			param := map[string]manifest.ParameterDefinition{
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
					Location: Location{
						Action:          "",
						Mixin:           "",
						StepNumber:      0,
						StepDescription: "",
					},
					Code:    "porter-100",
					Title:   "Reserved name error",
					Message: tc.ParameterName + " has a reserved prefix. Parameters cannot start with porter- or porter_",
					URL:     "https://porter.sh/reference/linter/#porter-100",
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
		param := map[string]manifest.ParameterDefinition{
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

	t.Run("lint messages does not mention mixins in message not coming from mixin", func(t *testing.T) {
		cxt := portercontext.NewTestContext(t)
		mixins := mixin.NewTestMixinProvider()
		l := New(cxt.Context, mixins)
		param := map[string]manifest.ParameterDefinition{
			"A": {
				Name: "porter_test",
			},
		}

		m := &manifest.Manifest{
			Parameters: param,
		}

		results, err := l.Lint(ctx, m)
		require.NoError(t, err, "Lint failed")
		require.Len(t, results, 1, "linter should have returned 1 result")
		require.NotContains(t, results[0].String(), ": 0th step in the mixin ()")
	})
}

func TestLinter_Lint_ParameterDoesNotApplyTo(t *testing.T) {
	ctx := context.Background()
	testCases := []struct {
		action   string
		setSteps func(*manifest.Manifest, manifest.Steps)
	}{
		{"install", func(m *manifest.Manifest, steps manifest.Steps) { m.Install = steps }},
		{"upgrade", func(m *manifest.Manifest, steps manifest.Steps) { m.Upgrade = steps }},
		{"uninstall", func(m *manifest.Manifest, steps manifest.Steps) { m.Uninstall = steps }},
		{"customAction", func(m *manifest.Manifest, steps manifest.Steps) {
			m.CustomActions = make(map[string]manifest.Steps)
			m.CustomActions["customAction"] = steps
		}},
	}

	for _, tc := range testCases {
		t.Run(tc.action, func(t *testing.T) {
			cxt := portercontext.NewTestContext(t)
			mixins := mixin.NewTestMixinProvider()
			l := New(cxt.Context, mixins)

			param := map[string]manifest.ParameterDefinition{
				"doesNotApply": {
					Name:    "doesNotApply",
					ApplyTo: []string{"dummy"},
				},
			}
			steps := manifest.Steps{
				&manifest.Step{
					Data: map[string]interface{}{
						"exec": map[string]interface{}{
							"description": "exec step",
							"parameters": []string{
								"\"${ bundle.parameters.doesNotApply }\"",
							},
						},
					},
				},
			}
			m := &manifest.Manifest{
				SchemaVersion:     "1.0.1",
				TemplateVariables: []string{"bundle.parameters.doesNotApply"},
				Parameters:        param,
			}
			tc.setSteps(m, steps)

			lintResults := Results{
				{
					Level: LevelError,
					Location: Location{
						Action:          tc.action,
						Mixin:           "exec",
						StepNumber:      1,
						StepDescription: "exec step",
					},
					Code:    "porter-101",
					Title:   "Parameter does not apply to action",
					Message: fmt.Sprintf("Parameter doesNotApply does not apply to %s action", tc.action),
					URL:     "https://porter.sh/docs/references/linter/#porter-101",
				},
			}
			results, err := l.Lint(ctx, m)
			require.NoError(t, err, "Lint failed")
			require.Len(t, results, 1, "linter should have returned 1 result")
			require.Equal(t, lintResults, results, "unexpected lint results")
		})
	}
}

func TestLinter_Lint_ParameterAppliesTo(t *testing.T) {
	ctx := context.Background()
	testCases := []struct {
		action   string
		setSteps func(*manifest.Manifest, manifest.Steps)
	}{
		{"install", func(m *manifest.Manifest, steps manifest.Steps) { m.Install = steps }},
		{"upgrade", func(m *manifest.Manifest, steps manifest.Steps) { m.Upgrade = steps }},
		{"uninstall", func(m *manifest.Manifest, steps manifest.Steps) { m.Uninstall = steps }},
		{"customAction", func(m *manifest.Manifest, steps manifest.Steps) {
			m.CustomActions = make(map[string]manifest.Steps)
			m.CustomActions["customAction"] = steps
		}},
	}

	for _, tc := range testCases {
		t.Run(tc.action, func(t *testing.T) {
			cxt := portercontext.NewTestContext(t)
			mixins := mixin.NewTestMixinProvider()
			l := New(cxt.Context, mixins)

			param := map[string]manifest.ParameterDefinition{
				"appliesTo": {
					Name:    "appliesTo",
					ApplyTo: []string{tc.action},
				},
			}
			steps := manifest.Steps{
				&manifest.Step{
					Data: map[string]interface{}{
						"exec": map[string]interface{}{
							"description": "exec step",
							"parameters": []string{
								"\"${ bundle.parameters.appliesTo }\"",
							},
						},
					},
				},
			}
			m := &manifest.Manifest{
				SchemaVersion:     "1.0.1",
				TemplateVariables: []string{"bundle.parameters.appliesTo"},
				Parameters:        param,
			}
			tc.setSteps(m, steps)

			results, err := l.Lint(ctx, m)
			require.NoError(t, err, "Lint failed")
			require.Len(t, results, 0, "linter should have returned 1 result")
		})
	}
}

func TestLinter_DependencyMultipleTimes(t *testing.T) {
	t.Run("dependency defined multiple times", func(t *testing.T) {
		cxt := portercontext.NewTestContext(t)
		mixins := mixin.NewTestMixinProvider()
		l := New(cxt.Context, mixins)

		m := &manifest.Manifest{
			Dependencies: manifest.Dependencies{
				Requires: []*manifest.Dependency{
					{Name: "mysql"},
					{Name: "mysql"},
				},
			},
		}

		expectedResult := Results{
			{
				Code:    "porter-102",
				Title:   "Dependency error",
				Message: "The dependency mysql is defined multiple times",
				URL:     "https://porter.sh/reference/linter/#porter-102",
			},
		}

		results, err := l.Lint(context.Background(), m)
		require.NoError(t, err, "Lint failed")
		require.Len(t, results, 1, "linter should have returned 1 result")
		require.Equal(t, expectedResult, results, "unexpected lint results")
	})
	t.Run("no dependency defined multiple times", func(t *testing.T) {
		cxt := portercontext.NewTestContext(t)
		mixins := mixin.NewTestMixinProvider()
		l := New(cxt.Context, mixins)

		m := &manifest.Manifest{
			Dependencies: manifest.Dependencies{
				Requires: []*manifest.Dependency{
					{Name: "mysql"},
					{Name: "mongo"},
				},
			},
		}

		results, err := l.Lint(context.Background(), m)
		require.NoError(t, err, "Lint failed")
		require.Len(t, results, 0, "linter should have returned 0 result")
	})
	t.Run("no dependencies", func(t *testing.T) {
		cxt := portercontext.NewTestContext(t)
		mixins := mixin.NewTestMixinProvider()
		l := New(cxt.Context, mixins)

		m := &manifest.Manifest{}

		results, err := l.Lint(context.Background(), m)
		require.NoError(t, err, "Lint failed")
		require.Len(t, results, 0, "linter should have returned 0 result")
	})
}

func TestLinter_Lint_MissingMixin(t *testing.T) {
	cxt := portercontext.NewTestContext(t)
	mixins := mixin.NewTestMixinProvider()
	l := New(cxt.Context, mixins)

	mixinName := "made-up-mixin-that-is-not-installed"

	m := &manifest.Manifest{
		Mixins: []manifest.MixinDeclaration{
			{
				Name: mixinName,
			},
		},
	}

	mixins.RunAssertions = append(mixins.RunAssertions, func(mixinCxt *portercontext.Context, mixinName string, commandOpts pkgmgmt.CommandOptions) error {
		return fmt.Errorf("%s not installed", mixinName)
	})

	_, err := l.Lint(context.Background(), m)
	require.Error(t, err, "Linting should return an error")
	tests.RequireOutputContains(t, err.Error(), fmt.Sprintf("%s is not currently installed", mixinName))
}
