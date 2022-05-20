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
	"github.com/pkg/errors"
)

// PorterRuntime orchestrates executing a bundle and managing state.
type PorterRuntime struct {
	*portercontext.Context
	mixins          pkgmgmt.PackageManager
	RuntimeManifest *RuntimeManifest
}

func NewPorterRuntime(cxt *portercontext.Context, mixins pkgmgmt.PackageManager) *PorterRuntime {
	return &PorterRuntime{
		Context: cxt,
		mixins:  mixins,
	}
}

func (r *PorterRuntime) Execute(ctx context.Context, rm *RuntimeManifest) error {
	r.RuntimeManifest = rm

	installationName := r.Getenv(config.EnvInstallationName)
	bundleName := r.Getenv(config.EnvBundleName)
	fmt.Fprintf(r.Out, "executing %s action from %s (installation: %s)\n", r.RuntimeManifest.Action, bundleName, installationName)

	err := r.RuntimeManifest.Validate()
	if err != nil {
		return err
	}

	// Prepare prepares the runtime environment prior to step execution.
	// As an example, for parameters of type "file", we may need to decode file contents
	// on the filesystem before execution of the step/action
	err = r.RuntimeManifest.Initialize()
	if err != nil {
		return err
	}

	// Update the runtimeManifest images with the bundle.json and relocation mapping (if it's there)
	rtb, reloMap, err := r.getImageMappingFiles()
	if err != nil {
		return err
	}

	err = r.RuntimeManifest.ResolveImages(rtb, reloMap)
	if err != nil {
		return errors.Wrap(err, "unable to resolve bundle images")
	}

	err = r.FileSystem.MkdirAll(portercontext.MixinOutputsDir, pkg.FileModeDirectory)
	if err != nil {
		return errors.Wrapf(err, "could not create outputs directory %s", portercontext.MixinOutputsDir)
	}

	var bigErr *multierror.Error
	for _, step := range r.RuntimeManifest.GetSteps() {
		err = r.executeStep(ctx, step)
		if err != nil {
			bigErr = multierror.Append(bigErr, err)
			break
		}
	}

	err = r.RuntimeManifest.Finalize()
	if err != nil {
		bigErr = multierror.Append(bigErr, err)
	}

	return bigErr.ErrorOrNil()
}

func (r *PorterRuntime) executeStep(ctx context.Context, step *manifest.Step) error {
	if step == nil {
		return nil
	}
	err := r.RuntimeManifest.ResolveStep(step)
	if err != nil {
		return errors.Wrap(err, "unable to resolve step")
	}

	description, _ := step.GetDescription()
	if len(description) > 0 {
		fmt.Fprintln(r.Out, description)
	}

	// Hand over values needing masking in context output streams
	r.Context.SetSensitiveValues(r.RuntimeManifest.GetSensitiveValues())

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
	err = r.mixins.Run(ctx, r.Context, step.GetMixinName(), cmd)
	if err != nil {
		return errors.Wrap(err, "mixin execution failed")
	}

	outputs, err := r.readMixinOutputs()
	if err != nil {
		return errors.Wrap(err, "could not read step outputs")
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

			err := r.FileSystem.WriteFile(outpath, []byte(outputValue), pkg.FileModeWritable)
			if err != nil {
				return errors.Wrapf(err, "unable to write output file %s", outpath)
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

	outfiles, err := r.FileSystem.ReadDir(portercontext.MixinOutputsDir)
	if err != nil {
		return nil, errors.Wrapf(err, "could not list %s", portercontext.MixinOutputsDir)
	}

	for _, outfile := range outfiles {
		if outfile.IsDir() {
			continue
		}
		outpath := filepath.Join(portercontext.MixinOutputsDir, outfile.Name())
		contents, err := r.FileSystem.ReadFile(outpath)
		if err != nil {
			return nil, errors.Wrapf(err, "could not read output file %s", outpath)
		}

		outputs[outfile.Name()] = string(contents)

		err = r.FileSystem.Remove(outpath)
		if err != nil {
			return nil, err
		}
	}

	return outputs, nil
}

func (r *PorterRuntime) getImageMappingFiles() (cnab.ExtendedBundle, relocation.ImageRelocationMap, error) {
	// TODO(carolynvs): switch to returning a BundleReference
	b, err := cnab.LoadBundle(r.Context, "/cnab/bundle.json")
	if err != nil {
		return cnab.ExtendedBundle{}, nil, err
	}

	var reloMap relocation.ImageRelocationMap
	if _, err := r.FileSystem.Stat("/cnab/app/relocation-mapping.json"); err == nil {
		reloBytes, err := r.FileSystem.ReadFile("/cnab/app/relocation-mapping.json")
		if err != nil {
			return cnab.ExtendedBundle{}, nil, errors.Wrap(err, "couldn't read relocation file")
		}
		err = json.Unmarshal(reloBytes, &reloMap)
		if err != nil {
			return cnab.ExtendedBundle{}, nil, errors.Wrap(err, "couldn't load relocation file")
		}
	}
	return b, reloMap, nil
}
