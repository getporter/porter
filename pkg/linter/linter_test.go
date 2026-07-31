package linter

import (
	"context"
	"fmt"
	"testing"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/experimental"
	"get.porter.sh/porter/pkg/manifest"
	"get.porter.sh/porter/pkg/mixin"
	"get.porter.sh/porter/pkg/pkgmgmt"
	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/tests"
	"github.com/Masterminds/semver/v3"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/stretchr/testify/require"
)

func TestLinter_Lint(t *testing.T) {
	ctx := context.Background()
	testConfig := config.NewTestConfig(t).Config

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

		results, err := l.Lint(ctx, m, testConfig, nil)
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

		results, err := l.Lint(ctx, m, testConfig, nil)
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

		results, err := l.Lint(ctx, m, testConfig, nil)
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

			results, err := l.Lint(ctx, m, testConfig, nil)
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

		results, err := l.Lint(ctx, m, testConfig, nil)
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

		results, err := l.Lint(ctx, m, testConfig, nil)
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
	testConfig := config.NewTestConfig(t).Config

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
			results, err := l.Lint(ctx, m, testConfig, nil)
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
	testConfig := config.NewTestConfig(t).Config

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

			results, err := l.Lint(ctx, m, testConfig, nil)
			require.NoError(t, err, "Lint failed")
			require.Len(t, results, 0, "linter should have returned 1 result")
		})
	}
}

func TestLinter_DependencyMultipleTimes(t *testing.T) {
	testConfig := config.NewTestConfig(t).Config

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

		results, err := l.Lint(context.Background(), m, testConfig, nil)
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

		results, err := l.Lint(context.Background(), m, testConfig, nil)
		require.NoError(t, err, "Lint failed")
		require.Len(t, results, 0, "linter should have returned 0 result")
	})
	t.Run("no dependencies", func(t *testing.T) {
		cxt := portercontext.NewTestContext(t)
		mixins := mixin.NewTestMixinProvider()
		l := New(cxt.Context, mixins)

		m := &manifest.Manifest{}

		results, err := l.Lint(context.Background(), m, testConfig, nil)
		require.NoError(t, err, "Lint failed")
		require.Len(t, results, 0, "linter should have returned 0 result")
	})
}

func TestLinter_DependencyMappings(t *testing.T) {
	testConfig := config.NewTestConfig(t).Config

	newDepBundle := func() cnab.ExtendedBundle {
		return cnab.ExtendedBundle{Bundle: bundle.Bundle{
			Parameters: map[string]bundle.Parameter{
				"PARAM_A": {Definition: "param-a"},
			},
			Credentials: map[string]bundle.Credential{
				"CRED_A": {},
			},
		}}
	}

	t.Run("undefined parameter mapping", func(t *testing.T) {
		cxt := portercontext.NewTestContext(t)
		mixins := mixin.NewTestMixinProvider()
		l := New(cxt.Context, mixins)

		m := &manifest.Manifest{
			Dependencies: manifest.Dependencies{
				Requires: []*manifest.Dependency{
					{Name: "mysql", Parameters: map[string]string{"NOT_DEFINED": "value"}},
				},
			},
		}
		depBundles := map[string]cnab.ExtendedBundle{"mysql": newDepBundle()}

		expectedResult := Results{
			{
				Code:    "porter-103",
				Title:   "Dependency error",
				Message: "dependencies.mysql.parameters.NOT_DEFINED is not defined as a parameter on the dependency bundle",
				URL:     "https://porter.sh/reference/linter/#porter-103",
			},
		}

		results, err := l.Lint(context.Background(), m, testConfig, depBundles)
		require.NoError(t, err, "Lint failed")
		require.Equal(t, expectedResult, results, "unexpected lint results")
	})

	t.Run("undefined credential mapping", func(t *testing.T) {
		cxt := portercontext.NewTestContext(t)
		mixins := mixin.NewTestMixinProvider()
		l := New(cxt.Context, mixins)

		m := &manifest.Manifest{
			Dependencies: manifest.Dependencies{
				Requires: []*manifest.Dependency{
					{Name: "mysql", Credentials: map[string]string{"NOT_DEFINED": "value"}},
				},
			},
		}
		depBundles := map[string]cnab.ExtendedBundle{"mysql": newDepBundle()}

		expectedResult := Results{
			{
				Code:    "porter-104",
				Title:   "Dependency error",
				Message: "dependencies.mysql.credentials.NOT_DEFINED is not defined as a credential on the dependency bundle",
				URL:     "https://porter.sh/reference/linter/#porter-104",
			},
		}

		results, err := l.Lint(context.Background(), m, testConfig, depBundles)
		require.NoError(t, err, "Lint failed")
		require.Equal(t, expectedResult, results, "unexpected lint results")
	})

	t.Run("valid mappings", func(t *testing.T) {
		cxt := portercontext.NewTestContext(t)
		mixins := mixin.NewTestMixinProvider()
		l := New(cxt.Context, mixins)

		m := &manifest.Manifest{
			Dependencies: manifest.Dependencies{
				Requires: []*manifest.Dependency{
					{
						Name:        "mysql",
						Parameters:  map[string]string{"PARAM_A": "value"},
						Credentials: map[string]string{"CRED_A": "value"},
					},
				},
			},
		}
		depBundles := map[string]cnab.ExtendedBundle{"mysql": newDepBundle()}

		results, err := l.Lint(context.Background(), m, testConfig, depBundles)
		require.NoError(t, err, "Lint failed")
		require.Len(t, results, 0, "linter should have returned 0 results")
	})

	t.Run("unresolved dependency is skipped", func(t *testing.T) {
		cxt := portercontext.NewTestContext(t)
		mixins := mixin.NewTestMixinProvider()
		l := New(cxt.Context, mixins)

		m := &manifest.Manifest{
			Dependencies: manifest.Dependencies{
				Requires: []*manifest.Dependency{
					{Name: "mysql", Parameters: map[string]string{"NOT_DEFINED": "value"}},
				},
			},
		}

		results, err := l.Lint(context.Background(), m, testConfig, nil)
		require.NoError(t, err, "Lint failed")
		require.Len(t, results, 0, "linter should skip validation for unresolved dependencies")
	})
}

func TestLinter_DependencyTemplateVariables(t *testing.T) {
	testConfig := config.NewTestConfig(t).Config
	testConfig.SetExperimentalFlags(experimental.FlagDependenciesV2)

	newDepBundle := func(outputs ...string) cnab.ExtendedBundle {
		outputDefs := make(map[string]bundle.Output, len(outputs))
		for _, o := range outputs {
			outputDefs[o] = bundle.Output{Definition: o}
		}
		return cnab.ExtendedBundle{Bundle: bundle.Bundle{Outputs: outputDefs}}
	}

	t.Run("unsupported prefix in parameters mapping", func(t *testing.T) {
		cxt := portercontext.NewTestContext(t)
		mixins := mixin.NewTestMixinProvider()
		l := New(cxt.Context, mixins)

		m := &manifest.Manifest{
			SchemaVersion: "1.0.1",
			Dependencies: manifest.Dependencies{
				Requires: []*manifest.Dependency{
					{Name: "mysql", Parameters: map[string]string{"dbstuff": "${env.FOO}"}},
				},
			},
		}

		results, err := l.Lint(context.Background(), m, testConfig, nil)
		require.NoError(t, err, "Lint failed")
		require.Len(t, results, 1)
		require.Equal(t, Code("porter-106"), results[0].Code)
	})

	t.Run("outputs shorthand used outside output mapping", func(t *testing.T) {
		cxt := portercontext.NewTestContext(t)
		mixins := mixin.NewTestMixinProvider()
		l := New(cxt.Context, mixins)

		m := &manifest.Manifest{
			SchemaVersion: "1.0.1",
			Dependencies: manifest.Dependencies{
				Requires: []*manifest.Dependency{
					{Name: "mysql", Credentials: map[string]string{"token": "${outputs.port}"}},
				},
			},
		}

		results, err := l.Lint(context.Background(), m, testConfig, nil)
		require.NoError(t, err, "Lint failed")
		require.Len(t, results, 1)
		require.Equal(t, Code("porter-106"), results[0].Code)
	})

	t.Run("unsupported installation field", func(t *testing.T) {
		cxt := portercontext.NewTestContext(t)
		mixins := mixin.NewTestMixinProvider()
		l := New(cxt.Context, mixins)

		m := &manifest.Manifest{
			SchemaVersion: "1.0.1",
			Dependencies: manifest.Dependencies{
				Requires: []*manifest.Dependency{
					{Name: "mysql", Parameters: map[string]string{"dbstuff": "${installation.foo}"}},
				},
			},
		}

		results, err := l.Lint(context.Background(), m, testConfig, nil)
		require.NoError(t, err, "Lint failed")
		require.Len(t, results, 1)
		require.Equal(t, Code("porter-106"), results[0].Code)
	})

	t.Run("valid installation fields", func(t *testing.T) {
		cxt := portercontext.NewTestContext(t)
		mixins := mixin.NewTestMixinProvider()
		l := New(cxt.Context, mixins)

		m := &manifest.Manifest{
			SchemaVersion: "1.0.1",
			Dependencies: manifest.Dependencies{
				Requires: []*manifest.Dependency{
					{Name: "mysql", Parameters: map[string]string{
						"a": "${installation.name}",
						"b": "${installation.namespace}",
						"c": "${installation.id}",
					}},
				},
			},
		}

		results, err := l.Lint(context.Background(), m, testConfig, nil)
		require.NoError(t, err, "Lint failed")
		require.Len(t, results, 0)
	})

	t.Run("undefined parent parameter", func(t *testing.T) {
		cxt := portercontext.NewTestContext(t)
		mixins := mixin.NewTestMixinProvider()
		l := New(cxt.Context, mixins)

		m := &manifest.Manifest{
			SchemaVersion: "1.0.1",
			Dependencies: manifest.Dependencies{
				Requires: []*manifest.Dependency{
					{Name: "mysql", Parameters: map[string]string{"dbstuff": "${bundle.parameters.stuff}"}},
				},
			},
		}

		results, err := l.Lint(context.Background(), m, testConfig, nil)
		require.NoError(t, err, "Lint failed")
		require.Len(t, results, 1)
		require.Equal(t, Code("porter-107"), results[0].Code)
	})

	t.Run("undefined parent credential", func(t *testing.T) {
		cxt := portercontext.NewTestContext(t)
		mixins := mixin.NewTestMixinProvider()
		l := New(cxt.Context, mixins)

		m := &manifest.Manifest{
			SchemaVersion: "1.0.1",
			Dependencies: manifest.Dependencies{
				Requires: []*manifest.Dependency{
					{Name: "mysql", Credentials: map[string]string{"token": "${bundle.credentials.token}"}},
				},
			},
		}

		results, err := l.Lint(context.Background(), m, testConfig, nil)
		require.NoError(t, err, "Lint failed")
		require.Len(t, results, 1)
		require.Equal(t, Code("porter-108"), results[0].Code)
	})

	t.Run("valid parameter and credential references", func(t *testing.T) {
		cxt := portercontext.NewTestContext(t)
		mixins := mixin.NewTestMixinProvider()
		l := New(cxt.Context, mixins)

		m := &manifest.Manifest{
			SchemaVersion: "1.0.1",
			Parameters:    manifest.ParameterDefinitions{"stuff": manifest.ParameterDefinition{Name: "stuff"}},
			Credentials:   manifest.CredentialDefinitions{"token": manifest.CredentialDefinition{Name: "token"}},
			Dependencies: manifest.Dependencies{
				Requires: []*manifest.Dependency{
					{
						Name:        "mysql",
						Parameters:  map[string]string{"dbstuff": "${bundle.parameters.stuff}"},
						Credentials: map[string]string{"token": "${bundle.credentials.token}"},
					},
				},
			},
		}

		results, err := l.Lint(context.Background(), m, testConfig, nil)
		require.NoError(t, err, "Lint failed")
		require.Len(t, results, 0)
	})

	t.Run("undefined output on self via shorthand", func(t *testing.T) {
		cxt := portercontext.NewTestContext(t)
		mixins := mixin.NewTestMixinProvider()
		l := New(cxt.Context, mixins)

		m := &manifest.Manifest{
			SchemaVersion: "1.0.1",
			Dependencies: manifest.Dependencies{
				Requires: []*manifest.Dependency{
					{Name: "mysql", Outputs: map[string]string{"connstr": "${outputs.port}"}},
				},
			},
		}
		depBundles := map[string]cnab.ExtendedBundle{"mysql": newDepBundle("user", "password")}

		results, err := l.Lint(context.Background(), m, testConfig, depBundles)
		require.NoError(t, err, "Lint failed")
		require.Len(t, results, 1)
		require.Equal(t, Code("porter-109"), results[0].Code)
	})

	t.Run("valid output on self via shorthand", func(t *testing.T) {
		cxt := portercontext.NewTestContext(t)
		mixins := mixin.NewTestMixinProvider()
		l := New(cxt.Context, mixins)

		m := &manifest.Manifest{
			SchemaVersion: "1.0.1",
			Dependencies: manifest.Dependencies{
				Requires: []*manifest.Dependency{
					{Name: "mysql", Outputs: map[string]string{"connstr": "${outputs.port}"}},
				},
			},
		}
		depBundles := map[string]cnab.ExtendedBundle{"mysql": newDepBundle("port")}

		results, err := l.Lint(context.Background(), m, testConfig, depBundles)
		require.NoError(t, err, "Lint failed")
		require.Len(t, results, 0)
	})

	t.Run("long-form reference to a different dependency's undefined output", func(t *testing.T) {
		cxt := portercontext.NewTestContext(t)
		mixins := mixin.NewTestMixinProvider()
		l := New(cxt.Context, mixins)

		m := &manifest.Manifest{
			SchemaVersion: "1.0.1",
			Dependencies: manifest.Dependencies{
				Requires: []*manifest.Dependency{
					{Name: "mysql"},
					{Name: "web", Parameters: map[string]string{"dbHost": "${bundle.dependencies.mysql.outputs.host}"}},
				},
			},
		}
		depBundles := map[string]cnab.ExtendedBundle{"mysql": newDepBundle("port")}

		results, err := l.Lint(context.Background(), m, testConfig, depBundles)
		require.NoError(t, err, "Lint failed")
		require.Len(t, results, 1)
		require.Equal(t, Code("porter-109"), results[0].Code)
	})

	t.Run("long-form reference to an unknown dependency", func(t *testing.T) {
		cxt := portercontext.NewTestContext(t)
		mixins := mixin.NewTestMixinProvider()
		l := New(cxt.Context, mixins)

		m := &manifest.Manifest{
			SchemaVersion: "1.0.1",
			Dependencies: manifest.Dependencies{
				Requires: []*manifest.Dependency{
					{Name: "web", Parameters: map[string]string{"dbHost": "${bundle.dependencies.unknown.outputs.host}"}},
				},
			},
		}

		results, err := l.Lint(context.Background(), m, testConfig, nil)
		require.NoError(t, err, "Lint failed")
		require.Len(t, results, 1)
		require.Equal(t, Code("porter-109"), results[0].Code)
	})

	t.Run("long-form reference satisfied by the other dependency's own outputs mapping", func(t *testing.T) {
		cxt := portercontext.NewTestContext(t)
		mixins := mixin.NewTestMixinProvider()
		l := New(cxt.Context, mixins)

		m := &manifest.Manifest{
			SchemaVersion: "1.0.1",
			Dependencies: manifest.Dependencies{
				Requires: []*manifest.Dependency{
					{Name: "mysql", Outputs: map[string]string{"connstr": "${outputs.port}"}},
					{Name: "web", Parameters: map[string]string{"dbConn": "${bundle.dependencies.mysql.outputs.connstr}"}},
				},
			},
		}
		// mysql's bundle isn't resolved here; "connstr" is only known from mysql's own outputs mapping keys.

		results, err := l.Lint(context.Background(), m, testConfig, nil)
		require.NoError(t, err, "Lint failed")
		require.Len(t, results, 0)
	})

	t.Run("unresolved dependency is skipped", func(t *testing.T) {
		cxt := portercontext.NewTestContext(t)
		mixins := mixin.NewTestMixinProvider()
		l := New(cxt.Context, mixins)

		m := &manifest.Manifest{
			SchemaVersion: "1.0.1",
			Dependencies: manifest.Dependencies{
				Requires: []*manifest.Dependency{
					{Name: "mysql", Outputs: map[string]string{"connstr": "${outputs.port}"}},
				},
			},
		}

		results, err := l.Lint(context.Background(), m, testConfig, nil)
		require.NoError(t, err, "Lint failed")
		require.Len(t, results, 0, "linter should skip output validation for unresolved dependencies")
	})

	t.Run("not enforced when dependencies v2 is disabled", func(t *testing.T) {
		cxt := portercontext.NewTestContext(t)
		mixins := mixin.NewTestMixinProvider()
		l := New(cxt.Context, mixins)
		v1Config := config.NewTestConfig(t).Config

		m := &manifest.Manifest{
			SchemaVersion: "1.0.1",
			Dependencies: manifest.Dependencies{
				Requires: []*manifest.Dependency{
					{Name: "mysql", Parameters: map[string]string{"dbstuff": "${env.FOO}"}},
				},
			},
		}

		results, err := l.Lint(context.Background(), m, v1Config, nil)
		require.NoError(t, err, "Lint failed")
		require.Len(t, results, 0, "dependency template variable checks should be skipped without FlagDependenciesV2")
	})
}

func TestLinter_Lint_MissingMixin(t *testing.T) {
	cxt := portercontext.NewTestContext(t)
	mixins := mixin.NewTestMixinProvider()
	l := New(cxt.Context, mixins)
	testConfig := config.NewTestConfig(t).Config

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

	_, err := l.Lint(context.Background(), m, testConfig, nil)
	require.Error(t, err, "Linting should return an error")
	tests.RequireOutputContains(t, err.Error(), fmt.Sprintf("%s is not currently installed", mixinName))
}

func TestLinter_Lint_MixinVersions(t *testing.T) {
	cxt := portercontext.NewTestContext(t)
	mixinProvider := mixin.NewTestMixinProvider()
	l := New(cxt.Context, mixinProvider)
	testConfig := config.NewTestConfig(t).Config

	exampleMixinVersion := mixin.ExampleMixinSemver.String()

	// build up some test semvers
	patchDifferenceSemver := fmt.Sprintf("%d.%d.%d", mixin.ExampleMixinSemver.Major(), mixin.ExampleMixinSemver.Minor(), mixin.ExampleMixinSemver.Patch()+1)
	anyPatchAccepted := fmt.Sprintf("%d.%d.x", mixin.ExampleMixinSemver.Major(), mixin.ExampleMixinSemver.Minor())
	lessThanNextMajor := fmt.Sprintf("<%d.%d", mixin.ExampleMixinSemver.Major()+1, mixin.ExampleMixinSemver.Minor())

	exampleMixinVersionConstraint, _ := semver.NewConstraint(exampleMixinVersion)
	patchDifferenceSemverConstraint, _ := semver.NewConstraint(patchDifferenceSemver)
	anyPatchAcceptedConstraint, _ := semver.NewConstraint(anyPatchAccepted)
	lessThanNextMajorConstraint, _ := semver.NewConstraint(lessThanNextMajor)

	testCases := []struct {
		name        string
		errExpected bool
		mixins      []manifest.MixinDeclaration
	}{
		{"exact-semver", false, []manifest.MixinDeclaration{
			{
				Name:    mixin.ExampleMixinName,
				Version: exampleMixinVersionConstraint,
			},
		}},
		{"different-patch", true, []manifest.MixinDeclaration{
			{
				Name:    mixin.ExampleMixinName,
				Version: patchDifferenceSemverConstraint,
			},
		}},
		{"accept-different-patch", false, []manifest.MixinDeclaration{
			{
				Name:    mixin.ExampleMixinName,
				Version: anyPatchAcceptedConstraint,
			},
		}},
		{"accept-less-than-versions", false, []manifest.MixinDeclaration{
			{
				Name:    mixin.ExampleMixinName,
				Version: lessThanNextMajorConstraint,
			},
		}},
		{"no-version-provided", false, []manifest.MixinDeclaration{
			{
				Name: mixin.ExampleMixinName,
			},
		}},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			m := &manifest.Manifest{
				Mixins: testCase.mixins,
			}
			results, err := l.Lint(context.Background(), m, testConfig, nil)
			if testCase.errExpected {
				require.Error(t, err, "Linting should return an error")
				tests.RequireOutputContains(t, err.Error(), fmt.Sprintf(
					"mixin %s is installed at version v%s but your bundle requires version %s",
					mixin.ExampleMixinName,
					exampleMixinVersion,
					testCase.mixins[0].Version.String(),
				))
			} else {
				require.NoError(t, err, "Linting should not return an error")
			}
			require.Len(t, results, 0, "linter should have returned 0 result")
		})
	}

}
