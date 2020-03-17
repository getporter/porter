package porter

import (
	"fmt"
	"os"
	"strconv"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/manifest"
	"get.porter.sh/porter/pkg/runtime"
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

	err = o.defaultDind()
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

func (o *RunOptions) defaultDind() error {
	// if dind was manually set, leave it
	if o.config.Dind {
		return nil
	}

	rawDind, set := os.LookupEnv(config.EnvDIND)
	if !set {
		return nil
	}

	dind, err := strconv.ParseBool(rawDind)
	if err != nil {
		return errors.Wrapf(err, "invalid PORTER_DIND, expected a bool (true/false) but got %s", rawDind)
	}

	if dind {
		fmt.Fprintf(o.config.Err, "DEBUG: defaulting dind to %s (%t)\n", config.EnvDIND, dind)
		o.config.Dind = dind
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
