package porter

import (
	"context"
	"fmt"
	"strconv"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/manifest"
	"get.porter.sh/porter/pkg/runtime"
	"get.porter.sh/porter/pkg/schema"
	"get.porter.sh/porter/pkg/tracing"
)

type RunOptions struct {
	config *config.Config

	// Debug specifies if the bundle should be run in debug mode
	DebugMode bool

	// File is the path to the porter manifest.
	File string

	// Action name to run in the bundle, such as install.
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
	}

	return nil
}

func (o *RunOptions) defaultDebug() error {
	// if debug was manually set, leave it
	if o.DebugMode {
		return nil
	}

	rawDebug, set := o.config.LookupEnv(config.EnvDEBUG)
	if !set {
		return nil
	}

	debug, err := strconv.ParseBool(rawDebug)
	if err != nil {
		return fmt.Errorf("invalid PORTER_DEBUG, expected a bool (true/false) but got %s: %w", rawDebug, err)
	}

	if debug {
		o.DebugMode = debug
	}

	return nil
}

func (p *Porter) Run(ctx context.Context, opts RunOptions) error {
	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()

	// Once the bundle has been built, we shouldn't check the schemaVersion again when running it.
	// If the author built it with the rules loosened, then it should execute regardless of the version matching.
	// A warning is printed if it doesn't match.
	p.Config.Data.SchemaCheck = string(schema.CheckStrategyNone)

	m, err := manifest.LoadManifestFrom(ctx, p.Config, opts.File)
	if err != nil {
		return span.Error(err)
	}

	runtimeCfg := runtime.NewConfigFor(p.Context)
	runtimeCfg.DebugMode = opts.DebugMode
	r := runtime.NewPorterRuntime(runtimeCfg, p.Mixins)
	runtimeManifest := r.NewRuntimeManifest(opts.Action, m)
	err = r.Execute(ctx, runtimeManifest)
	if err != nil {
		return span.Error(err)
	}

	span.Info("execution completed successfully!")
	return nil
}
