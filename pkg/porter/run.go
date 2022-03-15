package porter

import (
	"fmt"
	"strconv"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/manifest"
	"get.porter.sh/porter/pkg/runtime"
	"github.com/pkg/errors"
)

type RunOptions struct {
	config *config.Config

	File   string
	Action string
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
		o.Action = o.config.Getenv(config.EnvACTION)
		if o.config.Debug {
			fmt.Fprintf(o.config.Err, "DEBUG: defaulting action to %s (%s)\n", config.EnvACTION, o.Action)
		}
	}

	return nil
}

func (o *RunOptions) defaultDebug() error {
	// if debug was manually set, leave it
	if o.config.Debug {
		return nil
	}

	rawDebug, set := o.config.LookupEnv(config.EnvDEBUG)
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
	m, err := manifest.LoadManifestFrom(p.Context, opts.File)
	if err != nil {
		return err
	}

	runtimeManifest := runtime.NewRuntimeManifest(p.Context, opts.Action, m)
	r := runtime.NewPorterRuntime(p.Context, p.Mixins)
	err = r.Execute(runtimeManifest)
	if err == nil {
		fmt.Fprintln(r.Out, "execution completed successfully!")
	}
	return err
}
