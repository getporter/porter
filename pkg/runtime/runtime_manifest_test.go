package runtime

import (
	"context"
	"fmt"
	"io"
	"sort"
	"testing"

	"get.porter.sh/porter/pkg"
	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/manifest"
	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/tests"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-go/bundle/definition"
	"github.com/cnabio/cnab-to-oci/relocation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func runtimeManifestFromStepYaml(t *testing.T, pCtx *portercontext.TestContext, stepYaml string) *RuntimeManifest {
	mContent := []byte(stepYaml)
	require.NoError(t, pCtx.FileSystem.WriteFile("/cnab/app/porter.yaml", mContent, pkg.FileModeWritable))
	m, err := manifest.ReadManifest(pCtx.Context, "/cnab/app/porter.yaml")
	require.NoError(t, err, "ReadManifest failed")
	cfg := NewConfigFor(pCtx.Context)
	return NewRuntimeManifest(cfg, cnab.ActionInstall, m)
}

func TestResolveMapParam(t *testing.T) {
	ctx := context.Background()
	pCtx := portercontext.NewTestContext(t)
	pCtx.Setenv("PERSON", "Ralpha")

	mContent := `schemaVersion: 1.0.0-alpha.2
parameters:
- name: person
- name: place
  applyTo: [install]

install:
- mymixin:
    Parameters:
      Thing: ${ bundle.parameters.person }
`
	rm := runtimeManifestFromStepYaml(t, pCtx, mContent)
	s := rm.Install[0]

	err := rm.ResolveStep(ctx, 0, s)
	require.NoError(t, err)

	require.IsType(t, map[string]interface{}{}, s.Data["mymixin"], "Data.mymixin has incorrect type")
	mixin := s.Data["mymixin"].(map[string]interface{})
	require.IsType(t, mixin["Parameters"], map[string]interface{}{}, "Data.mymixin.Parameters has incorrect type")
	pms := mixin["Parameters"].(map[string]interface{})
	require.IsType(t, "string", pms["Thing"], "Data.mymixin.Parameters.Thing has incorrect type")
	val := pms["Thing"].(string)

	assert.Equal(t, "Ralpha", val)
	assert.NotContains(t, "place", pms, "parameters that don't apply to the current action should not be resolved")

	err = rm.Initialize(ctx)
	require.NoError(t, err)
}
func TestStateBagUnpack(t *testing.T) {
	ctx := context.Background()
	pCtx := portercontext.NewTestContext(t)
	pCtx.Setenv("PERSON", "Ralpha")

	mContent := `schemaVersion: 1.0.0-alpha.2
parameters:
- name: person
- name: place
  applyTo: [install]

install:
- mymixin:
    Parameters:
      Thing: ${ bundle.parameters.person }
state:
- name: foo
  path: foo/state.json
`
	tests := []struct {
		name         string
		stateContent string
		expErr       error
	}{
		{
			name:         "/porter/state.tgz is empty file",
			stateContent: "",
			expErr:       nil,
		},
		{
			name:         "/porter/state.tgz has null string",
			stateContent: "null",
			expErr:       nil,
		},
		{
			name:         "/porter/state.tgz has newline",
			stateContent: "\n",
			expErr:       io.ErrUnexpectedEOF,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rm := runtimeManifestFromStepYaml(t, pCtx, mContent)
			require.NoError(t, pCtx.FileSystem.WriteFile("/porter/state.tgz", []byte(test.stateContent), pkg.FileModeWritable))
			s := rm.Install[0]

			err := rm.ResolveStep(ctx, 0, s)
			require.NoError(t, err)

			err = rm.Initialize(ctx)
			if test.expErr == nil {
				require.NoError(t, err)
			} else {
				require.Contains(t, err.Error(), test.expErr.Error())
			}
			pCtx.FileSystem.Remove("/porter/state.tgz")
		})
	}
}

func TestResolvePathParam(t *testing.T) {
	ctx := context.Background()
	pCtx := portercontext.NewTestContext(t)

	mContent := `schemaVersion: 1.0.0-alpha.2
parameters:
- name: person
  path: person.txt

install:
- mymixin:
    Parameters:
      Thing: ${ bundle.parameters.person }
`
	rm := runtimeManifestFromStepYaml(t, pCtx, mContent)
	s := rm.Install[0]

	err := rm.ResolveStep(ctx, 0, s)
	require.NoError(t, err)

	require.IsType(t, map[string]interface{}{}, s.Data["mymixin"], "Data.mymixin has incorrect type")
	mixin := s.Data["mymixin"].(map[string]interface{})
	require.IsType(t, mixin["Parameters"], map[string]interface{}{}, "Data.mymixin.Parameters has incorrect type")
	pms := mixin["Parameters"].(map[string]interface{})
	require.IsType(t, "string", pms["Thing"], "Data.mymixin.Parameters.Thing has incorrect type")
	val := pms["Thing"].(string)

	assert.Equal(t, "person.txt", val)
}

func TestMetadataAvailableForTemplating(t *testing.T) {
	ctx := context.Background()
	c := config.NewTestConfig(t)

	c.TestContext.AddTestFile("testdata/metadata-substitution.yaml", config.Name)
	m, err := manifest.LoadManifestFrom(context.Background(), c.Config, config.Name)
	require.NoError(t, err, "LoadManifestFrom")
	cfg := NewConfigFor(c.Context)
	rm := NewRuntimeManifest(cfg, cnab.ActionInstall, m)

	s := rm.Install[0]
	err = rm.ResolveStep(ctx, 0, s)
	require.NoError(t, err)

	pms, ok := s.Data["exec"].(map[string]interface{})
	require.True(t, ok)
	cmd := pms["command"].(string)
	assert.Equal(t, "echo \"name:porter-hello version:0.1.0 description:An example Porter configuration image:jeremyrickard/porter-hello:porter-39a022ca907e26c3d8fffabd4bb8dbbc\"", cmd)
}

func TestDependencyMetadataAvailableForTemplating(t *testing.T) {
	ctx := context.Background()
	c := config.NewTestConfig(t)
	c.TestContext.AddTestFile("testdata/dep-metadata-substitution.yaml", config.Name)

	m, err := manifest.LoadManifestFrom(context.Background(), c.Config, config.Name)
	require.NoError(t, err, "LoadManifestFrom")
	cfg := NewConfigFor(c.Context)
	rm := NewRuntimeManifest(cfg, cnab.ActionInstall, m)
	rm.bundles = map[string]cnab.ExtendedBundle{
		"mysql": cnab.NewBundle(bundle.Bundle{
			Name:        "Azure MySQL",
			Description: "Azure MySQL database as a service",
			Version:     "v1.0.0",
		}),
	}

	s := rm.Install[0]
	err = rm.ResolveStep(ctx, 0, s)
	require.NoError(t, err)

	pms, ok := s.Data["exec"].(map[string]interface{})
	require.True(t, ok)
	cmd := pms["command"].(string)
	assert.Equal(t, "echo \"dep name: Azure MySQL dep version: v1.0.0 dep description: Azure MySQL database as a service\"", cmd)
}

func TestResolveMapParamUnknown(t *testing.T) {
	ctx := context.Background()
	pCtx := portercontext.NewTestContext(t)

	mContent := `schemaVersion: 1.0.0
install:
- mymixin:
    Parameters:
      Thing: ${bundle.parameters.person}
`
	rm := runtimeManifestFromStepYaml(t, pCtx, mContent)
	s := rm.Install[0]

	err := rm.ResolveStep(ctx, 0, s)
	require.Error(t, err)
	tests.RequireErrorContains(t, err, "Missing variable \"person\"")
}

func TestResolveArrayUnknown(t *testing.T) {
	ctx := context.Background()
	pCtx := portercontext.NewTestContext(t)

	mContent := `schemaVersion: 1.0.0
parameters:
- name: name

install:
- exec:
    Arguments:
      - ${bundle.parameters.person}
`
	rm := runtimeManifestFromStepYaml(t, pCtx, mContent)
	s := rm.Install[0]

	err := rm.ResolveStep(ctx, 0, s)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `Missing variable "person"`)
}

func TestResolveArray(t *testing.T) {
	ctx := context.Background()
	pCtx := portercontext.NewTestContext(t)
	pCtx.Setenv("PERSON", "Ralpha")

	mContent := `schemaVersion: 1.0.0
parameters:
- name: person

install:
- mymixin:
    Arguments:
    - ${ bundle.parameters.person }
`
	rm := runtimeManifestFromStepYaml(t, pCtx, mContent)
	s := rm.Install[0]

	err := rm.ResolveStep(ctx, 0, s)
	require.NoError(t, err)

	require.IsType(t, map[string]interface{}{}, s.Data["mymixin"], "Data.mymixin has incorrect type")
	mixin := s.Data["mymixin"].(map[string]interface{})
	require.IsType(t, mixin["Arguments"], []interface{}{}, "Data.mymixin.Arguments has incorrect type")
	args := mixin["Arguments"].([]interface{})

	assert.Equal(t, "Ralpha", args[0].(string))
}

func TestResolveSensitiveParameter(t *testing.T) {
	ctx := context.Background()
	pCtx := portercontext.NewTestContext(t)
	pCtx.Setenv("SENSITIVE_PARAM", "deliciou$dubonnet")
	pCtx.Setenv("REGULAR_PARAM", "regular param value")

	mContent := `schemaVersion: 1.0.0
parameters:
- name: sensitive_param
  sensitive: true
- name: regular_param

install:
- mymixin:
    Arguments:
    - ${ bundle.parameters.sensitive_param }
    - ${ bundle.parameters.regular_param }
`
	rm := runtimeManifestFromStepYaml(t, pCtx, mContent)
	s := rm.Install[0]

	// Prior to resolving step values, this method should return an empty string array
	assert.Equal(t, rm.GetSensitiveValues(), []string{})

	err := rm.ResolveStep(ctx, 0, s)
	require.NoError(t, err)

	require.IsType(t, map[string]interface{}{}, s.Data["mymixin"], "Data.mymixin has incorrect type")
	mixin := s.Data["mymixin"].(map[string]interface{})
	require.IsType(t, mixin["Arguments"], []interface{}{}, "Data.mymixin.Arguments has incorrect type")
	args := mixin["Arguments"].([]interface{})

	require.Len(t, args, 2)
	assert.Equal(t, "deliciou$dubonnet", args[0])
	assert.Equal(t, "regular param value", args[1])

	// There should now be one sensitive value tracked under the manifest
	assert.Equal(t, []string{"deliciou$dubonnet"}, rm.GetSensitiveValues())
}

func TestResolveCredential(t *testing.T) {
	ctx := context.Background()
	pCtx := portercontext.NewTestContext(t)
	pCtx.Setenv("PASSWORD", "deliciou$dubonnet")

	mContent := `schemaVersion: 1.0.0
credentials:
- name: password
  env: PASSWORD

install:
- mymixin:
    Arguments:
    - ${ bundle.credentials.password }
`
	rm := runtimeManifestFromStepYaml(t, pCtx, mContent)
	s := rm.Install[0]

	// Prior to resolving step values, this method should return an empty string array
	assert.Equal(t, rm.GetSensitiveValues(), []string{})

	err := rm.ResolveStep(ctx, 0, s)
	require.NoError(t, err)

	require.IsType(t, map[string]interface{}{}, s.Data["mymixin"], "Data.mymixin has incorrect type")
	mixin := s.Data["mymixin"].(map[string]interface{})
	require.IsType(t, mixin["Arguments"], []interface{}{}, "Data.mymixin.Arguments has incorrect type")
	args := mixin["Arguments"].([]interface{})

	assert.Equal(t, "deliciou$dubonnet", args[0])
	// There should now be a sensitive value tracked under the manifest
	assert.Equal(t, []string{"deliciou$dubonnet"}, rm.GetSensitiveValues())
}

func TestResolveStep_DependencyOutput(t *testing.T) {
	ctx := context.Background()
	pCtx := portercontext.NewTestContext(t)
	pCtx.Setenv("PORTER_MYSQL_PASSWORD_DEP_OUTPUT", "password")
	pCtx.Setenv("PORTER_MYSQL_ROOT_PASSWORD_DEP_OUTPUT", "mysql-password")

	mContent := `schemaVersion: 1.0.0
dependencies:
  requires: 
  - name: mysql
    bundle:
      reference: "getporter/porter-mysql"

install:
- mymixin:
    Arguments:
    - ${ bundle.dependencies.mysql.outputs.password }
    - ${ bundle.dependencies.mysql.outputs.root-password }
`
	rm := runtimeManifestFromStepYaml(t, pCtx, mContent)
	ps := cnab.ParameterSources{}
	ps.SetParameterFromDependencyOutput("porter-mysql-password", "mysql", "password")
	ps.SetParameterFromDependencyOutput("porter-mysql-root-password", "mysql", "root-password")
	rm.bundle = cnab.NewBundle(bundle.Bundle{
		Custom: map[string]interface{}{
			cnab.ParameterSourcesExtensionKey: ps,
		},
		RequiredExtensions: []string{cnab.ParameterSourcesExtensionKey},
	})

	rm.bundles = map[string]cnab.ExtendedBundle{
		"mysql": cnab.NewBundle(bundle.Bundle{
			Outputs: map[string]bundle.Output{
				"password": {
					Definition: "password",
				},
				"root-password": {
					Definition: "root-password",
				},
			},
			Definitions: map[string]*definition.Schema{
				"password":      {WriteOnly: makeBoolPtr(true)},
				"root-password": {WriteOnly: makeBoolPtr(true)},
			},
		}),
	}

	s := rm.Install[0]
	err := rm.ResolveStep(ctx, 0, s)
	require.NoError(t, err)

	require.IsType(t, map[string]interface{}{}, s.Data["mymixin"], "Data.mymixin has incorrect type")
	mixin := s.Data["mymixin"].(map[string]interface{})
	require.IsType(t, mixin["Arguments"], []interface{}{}, "Data.mymixin.Arguments has incorrect type")
	args := mixin["Arguments"].([]interface{})

	assert.Equal(t, []interface{}{"password", "mysql-password"}, args, "Incorrect template args passed to the mixin step")

	// There should now be a sensitive value tracked under the manifest
	gotSensitiveValues := rm.GetSensitiveValues()
	sort.Strings(gotSensitiveValues)
	assert.Equal(t, []string{"mysql-password", "password"}, gotSensitiveValues, "Incorrect values were marked as sensitive")
}

func TestResolveInMainDict(t *testing.T) {
	ctx := context.Background()
	c := config.NewTestConfig(t)

	c.TestContext.AddTestFile("testdata/param-test-in-block.yaml", config.Name)

	m, err := manifest.LoadManifestFrom(context.Background(), c.Config, config.Name)
	require.NoError(t, err, "could not load manifest")

	cfg := NewConfigFor(c.Context)
	rm := NewRuntimeManifest(cfg, cnab.ActionInstall, m)

	installStep := rm.Install[0]

	rm.config.Setenv("COMMAND", "echo hello world")
	err = rm.ResolveStep(ctx, 0, installStep)
	require.NoError(t, err)

	require.IsType(t, map[string]interface{}{}, installStep.Data["exec"], "Data.exec has the wrong type")
	exec := installStep.Data["exec"].(map[string]interface{})
	command := exec["command"]
	require.IsType(t, "string", command, "Data.exec.command has the wrong type")
	cmdVal := command.(string)

	assert.Equal(t, "echo hello world", cmdVal)
}

func TestResolveSliceWithAMap(t *testing.T) {
	ctx := context.Background()
	c := config.NewTestConfig(t)

	c.TestContext.AddTestFile("testdata/slice-test.yaml", config.Name)

	m, err := manifest.LoadManifestFrom(context.Background(), c.Config, config.Name)
	require.NoError(t, err, "could not load manifest")

	cfg := NewConfigFor(c.Context)
	rm := NewRuntimeManifest(cfg, cnab.ActionInstall, m)

	installStep := rm.Install[0]

	rm.config.Setenv("COMMAND", "echo hello world")
	err = rm.ResolveStep(ctx, 0, installStep)
	require.NoError(t, err)

	require.NotNil(t, installStep.Data)
	exec := installStep.Data["exec"].(map[string]interface{})
	require.NotNil(t, exec)
	flags := exec["flags"].(map[string]interface{})
	require.Len(t, flags, 1)
	assert.Equal(t, "echo hello world", flags["c"].(string))
}

func TestResolveMissingStepOutputs(t *testing.T) {
	ctx := context.Background()
	pCtx := portercontext.NewTestContext(t)

	mContent := `schemaVersion: 1.0.0
install:
- mymixin:
    Arguments:
    - jdbc://${bundle.outputs.database_url}:${bundle.outputs.database_port}
`
	rm := runtimeManifestFromStepYaml(t, pCtx, mContent)
	s := rm.Install[0]

	err := rm.ResolveStep(ctx, 0, s)
	tests.RequireErrorContains(t, err, `Missing variable "database_url"`)
}

func TestResolveSensitiveOutputs(t *testing.T) {
	ctx := context.Background()
	pCtx := portercontext.NewTestContext(t)
	mContent := `schemaVersion: 1.0.0
outputs:
- name: username
- name: password
  sensitive: true

install:
- mymixin:
    Arguments:
    - ${ bundle.outputs.username }
    - ${ bundle.outputs.password }
`
	rm := runtimeManifestFromStepYaml(t, pCtx, mContent)
	rm.outputs = map[string]string{
		"username": "sally",
		"password": "top$ecret!",
	}
	s := rm.Install[0]

	err := rm.ResolveStep(ctx, 0, s)
	require.NoError(t, err)

	require.IsType(t, s.Data["mymixin"], map[string]interface{}{}, "Data.mymixin has the wrong type")
	mixin := s.Data["mymixin"].(map[string]interface{})
	require.IsType(t, []interface{}{}, mixin["Arguments"], "Data.mymixin.Arguments has the wrong type")
	args := mixin["Arguments"].([]interface{})

	require.Len(t, args, 2)
	require.Equal(t, "sally", args[0])
	require.Equal(t, "top$ecret!", args[1])

	// There should be only one sensitive value being tracked
	require.Equal(t, []string{"top$ecret!"}, rm.GetSensitiveValues())
}

func TestManifest_ResolveBundleName(t *testing.T) {
	ctx := context.Background()
	pCtx := portercontext.NewTestContext(t)
	mContent := `schemaVersion: 1.0.0
name: mybuns

install:
- mymixin:
    Arguments:
    - ${ bundle.name }
`
	rm := runtimeManifestFromStepYaml(t, pCtx, mContent)
	s := rm.Install[0]

	err := rm.ResolveStep(ctx, 0, s)
	require.NoError(t, err)

	require.IsType(t, s.Data["mymixin"], map[string]interface{}{}, "Data.mymixin has the wrong type")
	mixin := s.Data["mymixin"].(map[string]interface{})
	require.IsType(t, []interface{}{}, mixin["Arguments"], "Data.mymixin.Arguments has the wrong type")
	args := mixin["Arguments"].([]interface{})

	assert.Equal(t, "mybuns", args[0].(string))
}

func TestReadManifest_Validate_BundleOutput(t *testing.T) {
	c := config.NewTestConfig(t)

	c.TestContext.AddTestFile("testdata/outputs/bundle-outputs.yaml", config.Name)

	wantOutputs := manifest.OutputDefinitions{
		"mysql-root-password": {
			Name: "mysql-root-password",
			Schema: definition.Schema{
				Description: "The root MySQL password",
				Type:        "string",
			},
		},
		"mysql-password": {
			Name: "mysql-password",
			Schema: definition.Schema{
				Type: "string",
			},
			ApplyTo: []string{
				"install",
				"upgrade",
			},
		},
	}

	m, err := manifest.LoadManifestFrom(context.Background(), c.Config, config.Name)
	require.NoError(t, err, "could not load manifest")

	require.Equal(t, wantOutputs, m.Outputs)
}

func TestReadManifest_Validate_BundleOutput_Error(t *testing.T) {
	c := config.NewTestConfig(t)

	c.TestContext.AddTestFile("testdata/outputs/bundle-outputs-error.yaml", config.Name)

	_, err := manifest.LoadManifestFrom(context.Background(), c.Config, config.Name)
	require.Error(t, err)
}

func TestDependencyV1_Validate(t *testing.T) {
	testcases := []struct {
		name       string
		dep        manifest.Dependency
		wantOutput string
		wantError  string
	}{
		{
			name:       "version in reference",
			dep:        manifest.Dependency{Name: "mysql", Bundle: manifest.BundleCriteria{Reference: "deislabs/azure-mysql:5.7"}},
			wantOutput: "",
			wantError:  "",
		}, {
			name:       "version ranges",
			dep:        manifest.Dependency{Name: "mysql", Bundle: manifest.BundleCriteria{Reference: "deislabs/azure-mysql", Version: "5.7.x-6"}},
			wantOutput: "",
			wantError:  "",
		}, {
			name:       "missing reference",
			dep:        manifest.Dependency{Name: "mysql", Bundle: manifest.BundleCriteria{Reference: ""}},
			wantOutput: "",
			wantError:  `reference is required for dependency "mysql"`,
		}, {
			name:       "version double specified",
			dep:        manifest.Dependency{Name: "mysql", Bundle: manifest.BundleCriteria{Reference: "deislabs/azure-mysql:5.7", Version: "5.7.x-6"}},
			wantOutput: "",
			wantError:  `reference for dependency "mysql" can only specify REGISTRY/NAME when version ranges are specified`,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			pCtx := portercontext.NewTestContext(t)

			err := tc.dep.Validate(pCtx.Context)

			if tc.wantError == "" {
				require.NoError(t, err)
			} else {
				require.Equal(t, tc.wantError, err.Error())
			}

			gotOutput := pCtx.GetOutput()
			if gotOutput != "" {
				require.Equal(t, tc.wantOutput, gotOutput)
			}
		})
	}
}

func TestManifest_ApplyStepOutputs(t *testing.T) {
	c := config.NewTestConfig(t)

	c.TestContext.AddTestFileFromRoot("pkg/manifest/testdata/porter-with-templating.yaml", config.Name)

	m, err := manifest.LoadManifestFrom(context.Background(), c.Config, config.Name)
	require.NoError(t, err, "could not load manifest")

	cfg := NewConfigFor(c.Context)
	rm := NewRuntimeManifest(cfg, cnab.ActionInstall, m)

	err = rm.ApplyStepOutputs(map[string]string{"name": "world"})
	require.NoError(t, err)

	assert.Contains(t, rm.outputs, "name")
	assert.Equal(t, "world", rm.outputs["name"])
}

func makeBoolPtr(value bool) *bool {
	return &value
}

func TestManifest_ResolveImageMap(t *testing.T) {
	ctx := context.Background()
	c := config.NewTestConfig(t)
	c.TestContext.AddTestFile("testdata/porter-images.yaml", config.Name)

	m, err := manifest.LoadManifestFrom(context.Background(), c.Config, config.Name)
	require.NoError(t, err, "could not load manifest")

	cfg := NewConfigFor(c.Context)
	rm := NewRuntimeManifest(cfg, cnab.ActionInstall, m)
	expectedImage, ok := m.ImageMap["something"]
	require.True(t, ok, "couldn't get expected image")
	expectedRef := fmt.Sprintf("%s@%s", expectedImage.Repository, expectedImage.Digest)
	step := rm.Install[0]
	err = rm.ResolveStep(ctx, 0, step)
	assert.NoError(t, err, "Should have successfully resolved step")
	s := step.Data["searcher"].(map[string]interface{})
	assert.NotNil(t, s)
	img, ok := s["image"]
	assert.True(t, ok, "should have found image")
	val := fmt.Sprintf("%v", img)
	assert.Equal(t, expectedRef, val)

	repo, ok := s["repo"]
	assert.True(t, ok, "should have found repo")
	val = fmt.Sprintf("%v", repo)
	assert.Equal(t, expectedImage.Repository, val)

	digest, ok := s["digest"]
	assert.True(t, ok, "should have found content digest")
	val = fmt.Sprintf("%v", digest)
	assert.Equal(t, expectedImage.Digest, val)

	tag, ok := s["tag"]
	assert.True(t, ok, "should have found tag")
	val = fmt.Sprintf("%v", tag)
	assert.Equal(t, expectedImage.Tag, val)
}

func TestManifest_ResolveImageMapMissingKey(t *testing.T) {
	// Try to access an images entry that doesn't exist
	ctx := context.Background()
	pCtx := portercontext.NewTestContext(t)
	mContent := `schemaVersion: 1.0.0-alpha.2
images:
  something:
    repository: "blah/blah"
    digest: "sha1234:cafebab"

install:
- mymixin:
    Arguments:
      - ${ bundle.images.notsomething.digest }
`
	rm := runtimeManifestFromStepYaml(t, pCtx, mContent)
	s := rm.Install[0]

	err := rm.ResolveStep(ctx, 0, s)
	tests.RequireErrorContains(t, err, `Missing variable "notsomething"`)
}

func TestResolveImage(t *testing.T) {
	tests := []struct {
		name      string
		reference string
		want      manifest.MappedImage
	}{
		{
			name:      "canonical reference",
			reference: "getporter/porter-hello@sha256:8b06c3da72dc9fa7002b9bc1f73a7421b4287c9cf0d3b08633287473707f9a63",
			want: manifest.MappedImage{
				Repository: "getporter/porter-hello",
				Digest:     "sha256:8b06c3da72dc9fa7002b9bc1f73a7421b4287c9cf0d3b08633287473707f9a63",
			},
		},
		{
			name:      "tagged reference",
			reference: "ghcr.io/getporter/examples/porter-hello:v0.2.0",
			want: manifest.MappedImage{
				Repository: "ghcr.io/getporter/examples/porter-hello",
				Tag:        "v0.2.0",
			},
		},
		{
			name:      "named reference",
			reference: "getporter/porter-hello",
			want: manifest.MappedImage{
				Repository: "getporter/porter-hello",
				Tag:        "latest",
			},
		},
		{
			name:      "the one with a hostname",
			reference: "deislabs.io/getporter/porter-hello",
			want: manifest.MappedImage{
				Repository: "deislabs.io/getporter/porter-hello",
				Tag:        "latest",
			},
		},
		{
			name:      "the one with a hostname and port",
			reference: "deislabs.io:9090/getporter/porter-hello:foo",
			want: manifest.MappedImage{
				Repository: "deislabs.io:9090/getporter/porter-hello",
				Tag:        "foo",
			},
		},
		{

			name:      "tagged and digested",
			reference: "ghcr.io/getporter/examples/porter-hello:v0.2.0@sha256:8b06c3da72dc9fa7002b9bc1f73a7421b4287c9cf0d3b08633287473707f9a63",
			want: manifest.MappedImage{
				Repository: "ghcr.io/getporter/examples/porter-hello",
				Tag:        "v0.2.0",
				Digest:     "sha256:8b06c3da72dc9fa7002b9bc1f73a7421b4287c9cf0d3b08633287473707f9a63",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := &manifest.MappedImage{}
			err := resolveImage(got, test.reference)
			require.NoError(t, err)
			assert.Equal(t, test.want.Repository, got.Repository)
			assert.Equal(t, test.want.Tag, got.Tag)
			assert.Equal(t, test.want.Digest, got.Digest)
		})
	}
}

func TestResolveImageErrors(t *testing.T) {
	tests := []struct {
		name      string
		reference string
		want      string
	}{
		{
			name:      "no algo digest",
			reference: "getporter/porter-hello@8b06c3da72dc9fa7002b9bc1f73a7421b4287c9cf0d3b08633287473707f9a63",
			want:      "invalid reference format",
		},
		{
			name:      "bad digest",
			reference: "getporter/porter-hello@sha256:8b06c3da72dc9fa7002b9bc1f73a7421b4287c9cf0d3b08633287473707f",
			want:      "invalid checksum digest length",
		},
		{
			name:      "bad digest algo",
			reference: "getporter/porter-hello@sha356:8b06c3da72dc9fa7002b9bc1f73a7421b4287c9cf0d3b08633287473707f9a63",
			want:      "unsupported digest algorithm",
		},
		{
			name:      "malformed tagged ref",
			reference: "getporter/porter-hello@latest",
			want:      "invalid reference format",
		},
		{
			name:      "too many ports tagged ref",
			reference: "deislabs:8080:8080/porter-hello:latest",
			want:      "invalid reference format",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := &manifest.MappedImage{}
			err := resolveImage(got, test.reference)
			require.Error(t, err)
			assert.Contains(t, err.Error(), test.want)
		})
	}
}

func TestResolveImageWithUpdatedBundle(t *testing.T) {
	pCtx := portercontext.NewTestContext(t)
	m := &manifest.Manifest{
		ImageMap: map[string]manifest.MappedImage{
			"machine": manifest.MappedImage{
				Repository: "deislabs/ghost",
				Tag:        "latest",
				Digest:     "sha256:75c495e5ce9c428d482973d72e3ce9925e1db304a97946c9aa0b540d7537e041",
			},
		},
	}

	img := bundle.Image{}
	img.Image = "blah/ghost:latest"
	img.Digest = "sha256:75c495e5ce9c428d482973d72e3ce9925e1db304a97946c9aa0b540d7537e041"
	bun := cnab.NewBundle(bundle.Bundle{
		Images: map[string]bundle.Image{
			"machine": img,
		},
	})

	reloMap := relocation.ImageRelocationMap{}

	cfg := NewConfigFor(pCtx.Context)
	rm := NewRuntimeManifest(cfg, cnab.ActionInstall, m)
	err := rm.ResolveImages(bun, reloMap)
	require.NoError(t, err)
	mi := rm.ImageMap["machine"]
	assert.Equal(t, "blah/ghost", mi.Repository)
}

func TestResolveImageWithUpdatedMismatchedBundle(t *testing.T) {
	pCtx := portercontext.NewTestContext(t)
	m := &manifest.Manifest{
		ImageMap: map[string]manifest.MappedImage{
			"machine": manifest.MappedImage{
				Repository: "deislabs/ghost",
				Tag:        "latest",
				Digest:     "sha256:75c495e5ce9c428d482973d72e3ce9925e1db304a97946c9aa0b540d7537e041",
			},
		},
	}

	img := bundle.Image{}
	img.Image = "blah/ghost:latest"
	img.Digest = "sha256:75c495e5ce9c428d482973d72e3ce9925e1db304a97946c9aa0b540d7537e041"
	bun := cnab.NewBundle(bundle.Bundle{
		Images: map[string]bundle.Image{
			"ghost": img,
		},
	})

	reloMap := relocation.ImageRelocationMap{}

	cfg := NewConfigFor(pCtx.Context)
	rm := NewRuntimeManifest(cfg, cnab.ActionInstall, m)
	err := rm.ResolveImages(bun, reloMap)
	assert.Error(t, err)
	assert.EqualError(t, err, fmt.Sprintf("unable to find image in porter manifest: %s", "ghost"))

}

func TestResolveImageWithRelo(t *testing.T) {
	pCtx := portercontext.NewTestContext(t)
	m := &manifest.Manifest{
		ImageMap: map[string]manifest.MappedImage{
			"machine": manifest.MappedImage{
				Repository: "gabrtv/microservice",
				Tag:        "latest",
				Digest:     "sha256:cca460afa270d4c527981ef9ca4989346c56cf9b20217dcea37df1ece8120687",
			},
		},
	}

	img := bundle.Image{}
	img.Image = "gabrtv/microservice@sha256:cca460afa270d4c527981ef9ca4989346c56cf9b20217dcea37df1ece8120687"
	img.Digest = "sha256:cca460afa270d4c527981ef9ca4989346c56cf9b20217dcea37df1ece8120687"
	bun := cnab.NewBundle(bundle.Bundle{
		Images: map[string]bundle.Image{
			"machine": img,
		},
	})

	reloMap := relocation.ImageRelocationMap{
		"gabrtv/microservice@sha256:cca460afa270d4c527981ef9ca4989346c56cf9b20217dcea37df1ece8120687": "my.registry/microservice@sha256:cca460afa270d4c527981ef9ca4989346c56cf9b20217dcea37df1ece8120687",
	}

	cfg := NewConfigFor(pCtx.Context)
	rm := NewRuntimeManifest(cfg, cnab.ActionInstall, m)
	err := rm.ResolveImages(bun, reloMap)
	require.NoError(t, err)
	mi := rm.ImageMap["machine"]
	assert.Equal(t, "my.registry/microservice", mi.Repository)
}

func TestResolveImageRelocationNoMatch(t *testing.T) {
	pCtx := portercontext.NewTestContext(t)
	m := &manifest.Manifest{
		ImageMap: map[string]manifest.MappedImage{
			"machine": manifest.MappedImage{
				Repository: "deislabs/ghost",
				Tag:        "latest",
				Digest:     "sha256:75c495e5ce9c428d482973d72e3ce9925e1db304a97946c9aa0b540d7537e041",
			},
		},
	}

	img := bundle.Image{}
	img.Image = "deislabs/ghost:latest"
	img.Digest = "sha256:75c495e5ce9c428d482973d72e3ce9925e1db304a97946c9aa0b540d7537e041"
	bun := cnab.NewBundle(bundle.Bundle{
		Images: map[string]bundle.Image{
			"machine": img,
		},
	})

	reloMap := relocation.ImageRelocationMap{
		"deislabs/nogood:latest": "cnabio/ghost:latest",
	}

	cfg := NewConfigFor(pCtx.Context)
	rm := NewRuntimeManifest(cfg, cnab.ActionInstall, m)
	err := rm.ResolveImages(bun, reloMap)
	require.NoError(t, err)
	assert.Equal(t, "deislabs/ghost", rm.ImageMap["machine"].Repository)
}

func TestResolveStepEncoding(t *testing.T) {
	ctx := context.Background()
	pCtx := portercontext.NewTestContext(t)

	wantValue := `{"test":"value"}`
	pCtx.Setenv("TEST", wantValue)

	mContent := `schemaVersion: 1.0.0
parameters:
- name: test
  env: TEST

install:
- mymixin:
    Flags:
      c: '${bundle.parameters.test}'
`
	rm := runtimeManifestFromStepYaml(t, pCtx, mContent)
	s := rm.Install[0]

	err := rm.ResolveStep(ctx, 0, s)
	require.NoError(t, err)

	require.IsType(t, s.Data["mymixin"], map[string]interface{}{}, "Data.mymixin has the wrong type")
	mixin := s.Data["mymixin"].(map[string]interface{})
	require.IsType(t, map[string]interface{}{}, mixin["Flags"], "Data.mymixin.Flags has the wrong type")
	flags := mixin["Flags"].(map[string]interface{})

	assert.Equal(t, flags["c"], wantValue)
}

func TestResolveInstallation(t *testing.T) {
	ctx := context.Background()
	pCtx := portercontext.NewTestContext(t)
	pCtx.Setenv(config.EnvPorterInstallationNamespace, "mynamespace")
	pCtx.Setenv(config.EnvPorterInstallationName, "mybun")

	mContent := `schemaVersion: 1.0.0
install:
- mymixin:
    ns: ${ installation.namespace }
    release: ${ installation.name }
`
	rm := runtimeManifestFromStepYaml(t, pCtx, mContent)
	s := rm.Install[0]

	err := rm.ResolveStep(ctx, 0, s)
	require.NoError(t, err)

	require.IsType(t, map[string]interface{}{}, s.Data["mymixin"], "Data.mymixin has the wrong type")
	mixin := s.Data["mymixin"].(map[string]interface{})

	assert.Equal(t, "mynamespace", mixin["ns"], "installation.namespace was not rendered")
	assert.Equal(t, "mybun", mixin["release"], "installation.name was not rendered")
}

func TestResolveCustomMetadata(t *testing.T) {
	ctx := context.Background()
	pCtx := portercontext.NewTestContext(t)

	mContent := `schemaVersion: 1.0.0
custom:
  foo: foobar
  myApp:
    featureFlags:
      featureA: true

install:
- mymixin:
    release: ${ bundle.custom.foo }
    featureA: ${ bundle.custom.myApp.featureFlags.featureA }
    notabool: "${ bundle.custom.myApp.featureFlags.featureA }"
`
	rm := runtimeManifestFromStepYaml(t, pCtx, mContent)
	s := rm.Install[0]

	err := rm.ResolveStep(ctx, 0, s)
	require.NoError(t, err)

	require.IsType(t, map[string]interface{}{}, s.Data["mymixin"], "Data.mymixin has the wrong type")
	mixin := s.Data["mymixin"].(map[string]interface{})

	err = rm.ResolveStep(ctx, 0, s)
	require.NoError(t, err, "ResolveStep failed")

	assert.Equal(t, "foobar", mixin["release"], "custom metadata was not rendered")
	assert.Equal(t, true, mixin["featureA"], "nested custom metadata was not rendered, an unquoted boolean should render as a bool")
	assert.Equal(t, "true", mixin["notabool"], "a quoted boolean should render as a string")
}

func TestResolveEnvironmentVariable(t *testing.T) {
	ctx := context.Background()
	pCtx := portercontext.NewTestContext(t)
	pCtx.Setenv("foo", "foo-value")
	pCtx.Setenv("BAR", "bar-value")

	mContent := `schemaVersion: 1.0.0
install:
- mymixin:
    someInput: ${ env.foo }
    moreInput: ${ env.BAR }
`
	rm := runtimeManifestFromStepYaml(t, pCtx, mContent)
	s := rm.Install[0]

	err := rm.ResolveStep(ctx, 0, s)
	require.NoError(t, err)

	require.IsType(t, map[string]interface{}{}, s.Data["mymixin"], "Data.mymixin has the wrong type")
	mixin := s.Data["mymixin"].(map[string]interface{})

	assert.Equal(t, "foo-value", mixin["someInput"], "expected lower-case foo env var was resolved")
	assert.Equal(t, "bar-value", mixin["moreInput"], "expected upper-case BAR env var was resolved")
}

func TestResolveInvocationImage(t *testing.T) {
	testcases := []struct {
		name                string
		bundleInvocationImg bundle.BaseImage
		relocationMap       relocation.ImageRelocationMap
		expectedImg         string
		wantErr             string
	}{
		{name: "success with no relocation map",
			bundleInvocationImg: bundle.BaseImage{Image: "blah/ghost:latest", ImageType: "docker", Digest: "sha256:75c495e5ce9c428d482973d72e3ce9925e1db304a97946c9aa0b540d7537e041"},
			expectedImg:         "blah/ghost:latest@sha256:75c495e5ce9c428d482973d72e3ce9925e1db304a97946c9aa0b540d7537e041",
		},
		{name: "success with relocation map",
			bundleInvocationImg: bundle.BaseImage{Image: "blah/ghost:latest", ImageType: "docker", Digest: "sha256:75c495e5ce9c428d482973d72e3ce9925e1db304a97946c9aa0b540d7537e041"},
			relocationMap:       relocation.ImageRelocationMap{"blah/ghost:latest@sha256:75c495e5ce9c428d482973d72e3ce9925e1db304a97946c9aa0b540d7537e041": "relocated-ghost@sha256:75c495e5ce9c428d482973d72e3ce9925e1db304a97946c9aa0b540d7537e041"},
			expectedImg:         "relocated-ghost@sha256:75c495e5ce9c428d482973d72e3ce9925e1db304a97946c9aa0b540d7537e041",
		},
		{name: "success with no update",
			expectedImg: "test/image:latest",
		},
		{name: "failure with invalid digest",
			bundleInvocationImg: bundle.BaseImage{Image: "blah/ghost:latest", ImageType: "docker", Digest: "123"},
			wantErr:             "unable to get invocation image reference with digest",
		},
	}

	pCtx := portercontext.NewTestContext(t)
	cfg := NewConfigFor(pCtx.Context)

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {

			bun := cnab.NewBundle(bundle.Bundle{
				InvocationImages: []bundle.InvocationImage{
					{BaseImage: tc.bundleInvocationImg},
				},
			})
			m := &manifest.Manifest{
				Image: "test/image:latest",
			}
			rm := NewRuntimeManifest(cfg, cnab.ActionInstall, m)

			err := rm.ResolveInvocationImage(bun, tc.relocationMap)
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.expectedImg, m.Image)
		})
	}

}
