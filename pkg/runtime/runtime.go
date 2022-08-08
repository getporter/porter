package runtime

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"

	"get.porter.sh/porter/pkg"
	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/manifest"
	"get.porter.sh/porter/pkg/pkgmgmt"
	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/pkg/yaml"
	"github.com/cnabio/cnab-to-oci/relocation"
	"github.com/hashicorp/go-multierror"
)

// PorterRuntime orchestrates executing a bundle and managing state.
type PorterRuntime struct {
	config          RuntimeConfig
	mixins          pkgmgmt.PackageManager
	RuntimeManifest *RuntimeManifest
}

func NewPorterRuntime(runtimeCfg RuntimeConfig, mixins pkgmgmt.PackageManager) *PorterRuntime {
	return &PorterRuntime{
		config: runtimeCfg,
		mixins: mixins,
	}
}

func (r *PorterRuntime) Execute(ctx context.Context, rm *RuntimeManifest) error {
	r.RuntimeManifest = rm

	installationName := r.config.Getenv(config.EnvInstallationName)
	bundleName := r.config.Getenv(config.EnvBundleName)
	fmt.Fprintf(r.config.Out, "executing %s action from %s (installation: %s)\n", r.RuntimeManifest.Action, bundleName, installationName)

	err := r.RuntimeManifest.Validate()
	if err != nil {
		return err
	}

	// Prepare prepares the runtime environment prior to step execution.
	// As an example, for parameters of type "file", we may need to decode file contents
	// on the filesystem before execution of the step/action
	err = r.RuntimeManifest.Initialize(ctx)
	if err != nil {
		return err
	}

	// Update the runtimeManifest images with the bundle.json and relocation mapping (if it's there)
	rtb, reloMap, err := r.getImageMappingFiles()
	if err != nil {
		return err
	}

	err = r.RuntimeManifest.ResolveInvocationImage(rtb, reloMap)
	if err != nil {
		return fmt.Errorf("unable to resolve bundle invocation images: %w", err)
	}
	err = r.RuntimeManifest.ResolveImages(rtb, reloMap)
	if err != nil {
		return fmt.Errorf("unable to resolve bundle images: %w", err)
	}

	err = r.config.FileSystem.MkdirAll(portercontext.MixinOutputsDir, pkg.FileModeDirectory)
	if err != nil {
		return fmt.Errorf("could not create outputs directory %s: %w", portercontext.MixinOutputsDir, err)
	}

	var bigErr *multierror.Error
	for stepIndex, step := range r.RuntimeManifest.GetSteps() {
		err = r.executeStep(ctx, stepIndex, step)
		if err != nil {
			bigErr = multierror.Append(bigErr, err)
			break
		}
	}

	err = r.RuntimeManifest.Finalize(ctx)
	if err != nil {
		bigErr = multierror.Append(bigErr, err)
	}

	return bigErr.ErrorOrNil()
}

func (r *PorterRuntime) executeStep(ctx context.Context, stepIndex int, step *manifest.Step) error {
	if step == nil {
		return nil
	}
	err := r.RuntimeManifest.ResolveStep(ctx, stepIndex, step)
	if err != nil {
		return fmt.Errorf("unable to resolve step: %w", err)
	}

	description, _ := step.GetDescription()
	if len(description) > 0 {
		fmt.Fprintln(r.config.Out, description)
	}

	// Hand over values needing masking in config output streams
	r.config.Context.SetSensitiveValues(r.RuntimeManifest.GetSensitiveValues())

	input := &ActionInput{
		action: r.RuntimeManifest.Action,
		Steps:  []*manifest.Step{step},
	}
	inputBytes, _ := yaml.Marshal(input)
	cmd := pkgmgmt.CommandOptions{
		Command: string(r.RuntimeManifest.Action),
		Input:   string(inputBytes),
		Runtime: true,
	}
	err = r.mixins.Run(ctx, r.config.Context, step.GetMixinName(), cmd)
	if err != nil {
		return fmt.Errorf("mixin execution failed: %w", err)
	}

	outputs, err := r.readMixinOutputs()
	if err != nil {
		return fmt.Errorf("could not read step outputs: %w", err)
	}

	err = r.RuntimeManifest.ApplyStepOutputs(outputs)
	if err != nil {
		return err
	}

	// Apply any Bundle Outputs declared in this step
	return r.applyStepOutputsToBundle(outputs)
}

// applyStepOutputsToBundle writes the provided step outputs to the proper location
// in the bundle execution environment.
func (r *PorterRuntime) applyStepOutputsToBundle(outputs map[string]string) error {
	for outputKey, outputValue := range outputs {
		bundleOutput, ok := r.RuntimeManifest.Outputs[outputKey]
		if !ok {
			continue
		}

		if r.shouldApplyOutput(bundleOutput) {
			outpath := filepath.Join(config.BundleOutputsDir, bundleOutput.Name)

			err := r.config.FileSystem.WriteFile(outpath, []byte(outputValue), pkg.FileModeWritable)
			if err != nil {
				return fmt.Errorf("unable to write output file %s: %w", outpath, err)
			}
		}
	}
	return nil
}

func (r *PorterRuntime) shouldApplyOutput(output manifest.OutputDefinition) bool {
	if len(output.ApplyTo) == 0 {
		return true
	}

	for _, applyTo := range output.ApplyTo {
		if string(r.RuntimeManifest.Action) == applyTo {
			return true
		}
	}
	return false
}

func (r *PorterRuntime) readMixinOutputs() (map[string]string, error) {
	outputs := map[string]string{}

	outfiles, err := r.config.FileSystem.ReadDir(portercontext.MixinOutputsDir)
	if err != nil {
		return nil, fmt.Errorf("could not list %s: %w", portercontext.MixinOutputsDir, err)
	}

	for _, outfile := range outfiles {
		if outfile.IsDir() {
			continue
		}
		outpath := filepath.Join(portercontext.MixinOutputsDir, outfile.Name())
		contents, err := r.config.FileSystem.ReadFile(outpath)
		if err != nil {
			return nil, fmt.Errorf("could not read output file %s: %w", outpath, err)
		}

		outputs[outfile.Name()] = string(contents)

		err = r.config.FileSystem.Remove(outpath)
		if err != nil {
			return nil, err
		}
	}

	return outputs, nil
}

func (r *PorterRuntime) getImageMappingFiles() (cnab.ExtendedBundle, relocation.ImageRelocationMap, error) {
	// TODO(carolynvs): switch to returning a BundleReference
	b, err := cnab.LoadBundle(r.config.Context, "/cnab/bundle.json")
	if err != nil {
		return cnab.ExtendedBundle{}, nil, err
	}

	var reloMap relocation.ImageRelocationMap
	if _, err := r.config.FileSystem.Stat("/cnab/app/relocation-mapping.json"); err == nil {
		reloBytes, err := r.config.FileSystem.ReadFile("/cnab/app/relocation-mapping.json")
		if err != nil {
			return cnab.ExtendedBundle{}, nil, fmt.Errorf("couldn't read relocation file: %w", err)
		}
		err = json.Unmarshal(reloBytes, &reloMap)
		if err != nil {
			return cnab.ExtendedBundle{}, nil, fmt.Errorf("couldn't load relocation file: %w", err)
		}
	}
	return b, reloMap, nil
}

func (r *PorterRuntime) NewRuntimeManifest(action string, m *manifest.Manifest) *RuntimeManifest {
	return NewRuntimeManifest(r.config, action, m)
}
