package buildkit

import (
	"context"
	_ "embed"
	"testing"

	"get.porter.sh/porter/pkg/build"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/manifest"
	buildx "github.com/docker/buildx/build"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_parseBuildArgs(t *testing.T) {
	testcases := []struct {
		name      string
		inputArgs []string
		wantArgs  map[string]string
	}{
		{name: "valid args", inputArgs: []string{"A=1", "B=2=2", "C="},
			wantArgs: map[string]string{"A": "1", "B": "2=2", "C": ""}},
		{name: "missing equal sign", inputArgs: []string{"A"},
			wantArgs: map[string]string{}},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			var gotArgs = map[string]string{}
			parseBuildArgs(tc.inputArgs, gotArgs)
			assert.Equal(t, tc.wantArgs, gotArgs)
		})
	}
}

func Test_toNamedContexts(t *testing.T) {
	testcases := []struct {
		name      string
		inputArgs map[string]string
		wantArgs  map[string]buildx.NamedContext
	}{
		{name: "Basic conversion",
			inputArgs: map[string]string{"context1": "/path/to/context1", "context2": "/path/to/context2"},
			wantArgs:  map[string]buildx.NamedContext{"context1": {Path: "/path/to/context1"}, "context2": {Path: "/path/to/context2"}}},
		{name: "Single entry",
			inputArgs: map[string]string{"singlecontext": "/single/path"},
			wantArgs:  map[string]buildx.NamedContext{"singlecontext": {Path: "/single/path"}}},
		{name: "Empty path",
			inputArgs: map[string]string{"singlecontext": ""},
			wantArgs:  map[string]buildx.NamedContext{"singlecontext": {Path: ""}}},
		{name: "Empty input map",
			inputArgs: map[string]string{},
			wantArgs:  map[string]buildx.NamedContext{}},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			var gotArgs = toNamedContexts(tc.inputArgs)
			assert.Equal(t, tc.wantArgs, gotArgs)
		})
	}
}

func Test_flattenMap(t *testing.T) {
	tt := []struct {
		desc string
		inp  map[string]interface{}
		out  map[string]string
		err  bool
	}{
		{
			desc: "one pair",
			inp: map[string]interface{}{
				"key": "value",
			},
			out: map[string]string{
				"key": "value",
			},
			err: false,
		},
		{
			desc: "nested input",
			inp: map[string]interface{}{
				"key": map[string]string{
					"nestedKey": "value",
				},
			},
			out: map[string]string{
				"key.nestedKey": "value",
			},
			err: false,
		},
		{
			desc: "nested input",
			inp: map[string]interface{}{
				"key1": map[string]interface{}{
					"key2": map[string]string{
						"key3": "value",
					},
				},
			},
			out: map[string]string{
				"key1.key2.key3": "value",
			},
			err: false,
		},
		{
			desc: "multiple nested input",
			inp: map[string]interface{}{
				"key11": map[string]interface{}{
					"key12": map[string]string{
						"key13": "value1",
					},
				},
				"key21": map[string]string{
					"key22": "value2",
				},
			},
			out: map[string]string{
				"key11.key12.key13": "value1",
				"key21.key22":       "value2",
			},
			err: false,
		},
		{
			// CNAB represents null parameters as empty strings, so we will do the same, e.g. ARG CUSTOM_FOO=
			desc: "nil is converted empty string",
			inp: map[string]interface{}{
				"a": nil,
			},
			out: map[string]string{
				"a": "",
			},
			err: false,
		},
		{
			desc: "int is converted to string representation",
			inp: map[string]interface{}{
				"a": 1,
			},
			out: map[string]string{
				"a": "1",
			},
			err: false,
		},
		{
			desc: "bool is converted to string representation",
			inp: map[string]interface{}{
				"a": true,
			},
			out: map[string]string{
				"a": "true",
			},
			err: false,
		},
		{
			desc: "array is converted to string representation",
			inp: map[string]interface{}{
				"a": []string{"beep", "boop"},
			},
			out: map[string]string{
				"a": `["beep","boop"]`,
			},
			err: false,
		},
	}

	for _, tc := range tt {
		t.Run(tc.desc, func(t *testing.T) {
			out, err := flattenMap(tc.inp)
			if tc.err {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.out, out)
		})
	}
}

func Test_detectCustomBuildArgsUsed(t *testing.T) {
	contents := `FROM scratch
ARG BAR=1
ARG CUSTOM_FOO=2
ARG CUSTOM_FOO1_BAR=nope
ARG CUSTOM_STUFF_THINGS
ARG NOT_CUSTOM_ARG

CMD ["echo", "stuff"]
`
	argNames, err := detectCustomBuildArgsUsed(contents)
	require.NoError(t, err, "detectCustomBuildArgsUsed failed")

	wantArgs := map[string]struct{}{
		"CUSTOM_FOO":          {},
		"CUSTOM_FOO1_BAR":     {},
		"CUSTOM_STUFF_THINGS": {},
	}
	require.Equal(t, wantArgs, argNames)
}

// This value is exactly the max sized argument allowed
//
//go:embed testdata/max-arg.txt
var maxArg string

func TestBuilder_determineBuildArgs(t *testing.T) {
	// This value goes over the limit of arg size
	oversizedArg := maxArg + "oopstoobig"

	ctx := context.Background()
	c := config.NewTestConfig(t)
	b := NewBuilder(c.Config)
	m := &manifest.Manifest{
		Custom: map[string]interface{}{
			// these can cause a problem when used as a build arg in the Dockerfile
			"BIG_LABEL":         oversizedArg,
			"ANOTHER_BIG_LABEL": oversizedArg,
		},
	}
	// First we use the standard template, that does not use any custom build arguments
	c.TestContext.AddTestFileFromRoot("pkg/templates/templates/create/template.buildkit.Dockerfile", "/.cnab/Dockerfile")

	// Try manually passing too big of a --build-arg, this should fail
	opts := build.BuildImageOptions{
		BuildArgs: []string{"BIG_BUILD_ARG=" + oversizedArg},
	}
	_, err := b.determineBuildArgs(ctx, m, opts)
	assert.ErrorContains(t, err, "BIG_BUILD_ARG is longer than the max")

	// Make the --build-arg the max length, so it passes
	// Try making a too big custom value in porter.yaml and using it in the Dockerfile so that it still fails
	opts.BuildArgs = []string{"BIG_BUILD_ARG=" + maxArg}
	c.TestContext.AddTestFile("testdata/custom-build-arg.Dockerfile", "/.cnab/Dockerfile")
	_, err = b.determineBuildArgs(ctx, m, opts)
	require.ErrorContains(t, err, "CUSTOM_BIG_LABEL is longer than the max")

	// Get everything to pass by making all big args the max length
	m.Custom["BIG_LABEL"] = maxArg
	args, err := b.determineBuildArgs(ctx, m, opts)
	require.NoError(t, err, "determineBuildArgs should pass now that all args are at the max length")
	wantArgs := map[string]string{
		"BUNDLE_DIR":       "/cnab/app",
		"BIG_BUILD_ARG":    maxArg,
		"CUSTOM_BIG_LABEL": maxArg}
	require.Equal(t, wantArgs, args, "incorrect arguments returned")
}
