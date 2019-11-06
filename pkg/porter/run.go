package porter

import (
	"fmt"
	"os"
	"strconv"

	"github.com/deislabs/porter/pkg/config"
	"github.com/deislabs/porter/pkg/manifest"
	"github.com/deislabs/porter/pkg/runtime"
	"github.com/pkg/errors"
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
	err := p.LoadManifestFrom(opts.File)
	if err != nil {
		return err
	}

	runtimeManifest := runtime.NewRuntimeManifest(p.Context, opts.parsedAction, p.Manifest)
	r := runtime.NewPorterRuntime(p.Context, p.Mixins)
	return r.Execute(runtimeManifest)
}
