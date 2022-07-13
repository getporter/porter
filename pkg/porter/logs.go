package porter

import (
	"context"
	"errors"
	"fmt"

	"get.porter.sh/porter/pkg/portercontext"
)

// LogsShowOptions represent options for an installation logs show command
type LogsShowOptions struct {
	installationOptions
	RunID string
}

// Installation name passed to the command.
func (o *LogsShowOptions) Installation() string {
	return o.installationOptions.Name
}

// Validate validates the provided args, using the provided context,
// setting attributes of LogsShowOptions as applicable
func (o *LogsShowOptions) Validate(cxt *portercontext.Context) error {
	if o.Name != "" && o.RunID != "" {
		return errors.New("either --installation or --run should be specified, not both")
	}

	// Attempt to derive installation name from context
	err := o.installationOptions.defaultBundleFiles(cxt)
	if err != nil {
		return err
	}

	if o.File == "" && o.Name == "" && o.RunID == "" {
		return errors.New("either --installation or --run is required")
	}

	return nil
}

// ShowInstallationLogs shows logs for an installation, according to the provided options.
func (p *Porter) ShowInstallationLogs(ctx context.Context, opts *LogsShowOptions) error {
	logs, ok, err := p.GetInstallationLogs(ctx, opts)
	if err != nil {
		return err
	}

	if !ok {
		return errors.New("no logs found")
	}

	fmt.Fprintln(p.Out, logs)
	return nil
}

// GetInstallationLogs gets logs for an installation, according to the provided options
func (p *Porter) GetInstallationLogs(ctx context.Context, opts *LogsShowOptions) (string, bool, error) {
	err := p.applyDefaultOptions(ctx, &opts.installationOptions)
	if err != nil {
		return "", false, err
	}
	installation := opts.installationOptions.Name

	if opts.RunID != "" {
		return p.Installations.GetLogs(ctx, opts.RunID)
	}

	return p.Installations.GetLastLogs(ctx, opts.Namespace, installation)
}
