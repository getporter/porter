package porter

import (
	"context"
	"fmt"
	"strconv"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/manifest"
	"get.porter.sh/porter/pkg/runtime"
	"get.porter.sh/porter/pkg/schema"
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

func (p *Porter) Run(ctx context.Context, opts RunOptions) error {
	// Once the bundle has been built, we shouldn't check the schemaVersion again when running it.
	// If the author built it with the rules loosened, then it should execute regardless of the version matching.
	// A warning is printed if it doesn't match.
	p.Config.Data.SchemaCheck = string(schema.CheckStrategyNone)

	m, err := manifest.LoadManifestFrom(ctx, p.Config, opts.File)
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
