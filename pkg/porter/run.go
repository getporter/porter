package porter

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/deislabs/cnab-go/bundle/loader"
	"github.com/deislabs/porter/pkg/config"
	"github.com/deislabs/porter/pkg/context"
	"github.com/deislabs/porter/pkg/manifest"
	"github.com/deislabs/porter/pkg/mixin"
	"github.com/docker/cnab-to-oci/relocation"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

type RunOptions struct {
	config *config.Config

	File         string
	Action       string
	parsedAction manifest.Action
}

func NewRunOptions(c *config.Config) RunOptions {
	return RunOptions{
		config: c,
	}
}

func (o *RunOptions) Validate() error {
	err := o.defaultDebug()
	if err != nil {
		return err
	}

	err = o.validateAction()
	if err != nil {
		return err
	}

	return nil
}

func (o *RunOptions) validateAction() error {
	if o.Action == "" {
		o.Action = os.Getenv(config.EnvACTION)
		if o.config.Debug {
			fmt.Fprintf(o.config.Err, "DEBUG: defaulting action to %s (%s)\n", config.EnvACTION, o.Action)
		}
	}

	o.parsedAction = manifest.Action(o.Action)
	return nil
}

func (o *RunOptions) defaultDebug() error {
	// if debug was manually set, leave it
	if o.config.Debug {
		return nil
	}

	rawDebug, set := os.LookupEnv(config.EnvDEBUG)
	if !set {
		return nil
	}

	debug, err := strconv.ParseBool(rawDebug)
	if err != nil {
		return errors.Wrapf(err, "invalid PORTER_DEBUG, expected a bool (true/false) but got %s", rawDebug)
	}

	if debug {
		fmt.Fprintf(o.config.Err, "DEBUG: defaulting debug to %s (%t)\n", config.EnvDEBUG, debug)
		o.config.Debug = debug
	}

	return nil
}

func (p *Porter) Run(opts RunOptions) error {
	claimName := os.Getenv(config.EnvClaimName)
	bundleName := os.Getenv(config.EnvBundleName)
	fmt.Fprintf(p.Out, "executing %s action from %s (bundle instance: %s) defined in %s\n", opts.parsedAction, bundleName, claimName, opts.File)

	err := p.LoadManifestFrom(opts.File)
	if err != nil {
		return err
	}
	runtimeManifest := manifest.NewRuntimeManifest(p.Context, opts.parsedAction, p.Manifest)

	err = runtimeManifest.Validate()
	if err != nil {
		return err
	}

	// Prepare prepares the runtime environment prior to step execution.
	// As an example, for parameters of type "file", we may need to decode file contents
	// on the filesystem before execution of the step/action
	err = runtimeManifest.Prepare()
	if err != nil {
		return err
	}

	//Update the runtimeManifest images with the bundle.json and relocation mapping (if it's there)
	l := loader.New()
	bunBytes, err := p.FileSystem.ReadFile("/cnab/bundle.json")
	if err != nil {
		return errors.Wrap(err, "couldn't read runtime bundle.json")
	}
	rtb, err := l.LoadData(bunBytes)
	if err != nil {
		return errors.Wrap(err, "couldn't load runtime bundle.json")
	}
	var reloMap relocation.ImageRelocationMap
	if _, err := p.FileSystem.Stat("/cnab/app/relocation-mapping.json"); err == nil {
		reloBytes, err := p.FileSystem.ReadFile("/cnab/app/relocation-mapping.json")
		if err != nil {
			return errors.Wrap(err, "couldn't read relocation file")
		}
		err = json.Unmarshal(reloBytes, reloMap)
		if err != nil {
			return errors.Wrap(err, "couldn't load relocation file")
		}
	}
	err = runtimeManifest.ResolveImages(rtb, reloMap)
	if err != nil {
		return errors.Wrap(err, "unable to resolve bundle images")
	}
	err = p.FileSystem.MkdirAll(context.MixinOutputsDir, 0755)
	if err != nil {
		return errors.Wrapf(err, "could not create outputs directory %s", context.MixinOutputsDir)
	}

	for _, step := range runtimeManifest.GetSteps() {
		if step != nil {
			err := runtimeManifest.ResolveStep(step)
			if err != nil {
				return errors.Wrap(err, "unable to resolve sourced values")
			}

			description, _ := step.GetDescription()
			fmt.Fprintln(p.Out, description)

			// Hand over values needing masking in context output streams
			p.Context.SetSensitiveValues(runtimeManifest.GetSensitiveValues())

			input := &ActionInput{
				action: opts.parsedAction,
				Steps:  []*manifest.Step{step},
			}
			inputBytes, _ := yaml.Marshal(input)
			cmd := mixin.CommandOptions{
				Command: string(opts.parsedAction),
				Input:   string(inputBytes),
				Runtime: true,
			}
			err = p.Mixins.Run(p.Context, step.GetMixinName(), cmd)
			if err != nil {
				return errors.Wrap(err, "mixin execution failed")
			}

			outputs, err := p.readMixinOutputs()
			if err != nil {
				return errors.Wrap(err, "could not read step outputs")
			}

			err = runtimeManifest.ApplyStepOutputs(step, outputs)
			if err != nil {
				return err
			}

			// Apply any Bundle Outputs declared in this step
			err = p.ApplyBundleOutputs(opts, outputs)
			if err != nil {
				return err
			}
		}
	}

	fmt.Fprintln(p.Out, "execution completed successfully!")
	return nil
}

type ActionInput struct {
	action manifest.Action
	Steps  []*manifest.Step `yaml:"steps"`
}

// MarshalYAML marshals the step nested under the action
// install:
// - helm:
//   ...
// Solution from https://stackoverflow.com/a/42547226
func (a *ActionInput) MarshalYAML() (interface{}, error) {
	// encode the original
	b, err := yaml.Marshal(a.Steps)
	if err != nil {
		return nil, err
	}

	// decode it back to get a map
	var tmp interface{}
	err = yaml.Unmarshal(b, &tmp)
	if err != nil {
		return nil, err
	}
	stepMap := tmp.([]interface{})
	actionMap := map[string]interface{}{
		string(a.action): stepMap,
	}
	return actionMap, nil
}

func (p *Porter) readMixinOutputs() (map[string]string, error) {
	outputs := map[string]string{}

	outfiles, err := p.FileSystem.ReadDir(context.MixinOutputsDir)
	if err != nil {
		return nil, errors.Wrapf(err, "could not list %s", context.MixinOutputsDir)
	}

	for _, outfile := range outfiles {
		if outfile.IsDir() {
			continue
		}
		outpath := filepath.Join(context.MixinOutputsDir, outfile.Name())
		contents, err := p.FileSystem.ReadFile(outpath)
		if err != nil {
			return nil, errors.Wrapf(err, "could not read output file %s", outpath)
		}

		outputs[outfile.Name()] = string(contents)

		err = p.FileSystem.Remove(outpath)
		if err != nil {
			return nil, err
		}
	}

	return outputs, nil
}

// ApplyBundleOutputs writes the provided outputs to the proper location
// in the execution environment
func (p *Porter) ApplyBundleOutputs(opts RunOptions, outputs map[string]string) error {
	// Ensure outputs directory exists
	if err := p.FileSystem.MkdirAll(config.BundleOutputsDir, 0755); err != nil {
		return errors.Wrap(err, "unable to ensure CNAB outputs directory exists")
	}

	for outputKey, outputValue := range outputs {
		// Iterate through bundle outputs declared in the manifest
		for _, bundleOutput := range p.Manifest.Outputs {
			// If a given step output matches a bundle output, proceed
			if outputKey == bundleOutput.Name {
				doApply := true

				// If ApplyTo array non-empty, default doApply to false
				// and only set to true if at least one entry matches current Action
				if len(bundleOutput.ApplyTo) > 0 {
					doApply = false
					for _, applyTo := range bundleOutput.ApplyTo {
						if opts.Action == applyTo {
							doApply = true
						}
					}
				}

				if doApply {
					outpath := filepath.Join(config.BundleOutputsDir, bundleOutput.Name)

					err := p.FileSystem.WriteFile(outpath, []byte(outputValue), 0755)
					if err != nil {
						return errors.Wrapf(err, "unable to write output file %s", outpath)
					}
				}
			}
		}
	}
	return nil
}
