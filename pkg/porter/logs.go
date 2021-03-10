package porter

import (
	"fmt"

	"get.porter.sh/porter/pkg/context"
	"github.com/cnabio/cnab-go/claim"
	"github.com/pkg/errors"
)

// LogsShowOptions represent options for an installation logs show command
type LogsShowOptions struct {
	sharedOptions
	ClaimID string
}

// Installation name passed to the command.
func (o *LogsShowOptions) Installation() string {
	return o.sharedOptions.Name
}

// Validate validates the provided args, using the provided context,
// setting attributes of LogsShowOptions as applicable
func (o *LogsShowOptions) Validate(cxt *context.Context) error {
	if o.Name != "" && o.ClaimID != "" {
		return errors.New("either --installation or --run should be specified, not both")
	}

	// Attempt to derive installation name from context
	err := o.sharedOptions.defaultBundleFiles(cxt)
	if err != nil {
		return err
	}

	if o.File == "" && o.Name == "" && o.ClaimID == "" {
		return errors.New("either --installation or --run is required")
	}

	return nil
}

// ShowInstallationLogs shows logs for an installation, according to the provided options.
func (p *Porter) ShowInstallationLogs(opts *LogsShowOptions) error {
	logs, ok, err := p.GetInstallationLogs(opts)
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
func (p *Porter) GetInstallationLogs(opts *LogsShowOptions) (string, bool, error) {
	err := p.applyDefaultOptions(&opts.sharedOptions)
	if err != nil {
		return "", false, err
	}
	installation := opts.sharedOptions.Name

	if opts.ClaimID != "" {
		return claim.GetLogs(p.Claims, opts.ClaimID)
	}

	return claim.GetLastLogs(p.Claims, installation)
}
