package porter

import (
	"context"
	"errors"
	"fmt"

	"get.porter.sh/porter/pkg/encoding"
	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/pkg/printer"
	"get.porter.sh/porter/pkg/storage"
	"get.porter.sh/porter/pkg/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap/zapcore"
)

type ApplyOptions struct {
	Namespace string
	File      string

	// Force the installation to be re-applied regardless of anything being changed or not
	Force bool

	// DryRun only checks if the changes would trigger a bundle run
	DryRun bool
}

const ApplyDefaultFormat = printer.FormatPlaintext

var ApplyAllowedFormats = printer.Formats{printer.FormatPlaintext, printer.FormatYaml, printer.FormatJson}

func (o *ApplyOptions) Validate(cxt *portercontext.Context, args []string) error {
	switch len(args) {
	case 0:
		return errors.New("a file argument is required")
	case 1:
		o.File = args[0]
	default:
		return errors.New("only one file argument may be specified")
	}

	info, err := cxt.FileSystem.Stat(o.File)
	if err != nil {
		return fmt.Errorf("invalid file argument %s: %w", o.File, err)
	}
	if info.IsDir() {
		return fmt.Errorf("invalid file argument %s, must be a file not a directory", o.File)
	}

	return nil
}

func (p *Porter) InstallationApply(ctx context.Context, opts ApplyOptions) error {
	ctx, log := tracing.StartSpan(ctx)
	defer log.EndSpan()

	log.Debugf("Reading input file %s", opts.File)

	namespace, err := p.getNamespaceFromFile(opts)
	if err != nil {
		return err
	}

	if log.ShouldLog(zapcore.DebugLevel) {
		// ignoring any error here, printing debug info isn't critical
		contents, _ := p.FileSystem.ReadFile(opts.File)
		log.Debug("read input file", attribute.String("contents", string(contents)))
	}

	var input DisplayInstallation
	if err := encoding.UnmarshalFile(p.FileSystem, opts.File, &input); err != nil {
		return fmt.Errorf("unable to parse %s as an installation document: %w", opts.File, err)
	}
	input.Namespace = namespace
	inputInstallation, err := input.ConvertToInstallation()
	if err != nil {
		return err
	}

	installation, err := p.Installations.GetInstallation(ctx, inputInstallation.Namespace, inputInstallation.Name)
	if err != nil {
		if !errors.Is(err, storage.ErrNotFound{}) {
			return fmt.Errorf("could not query for an existing installation document for %s: %w", inputInstallation, err)
		}

		// Create a new installation
		installation = storage.NewInstallation(input.Namespace, input.Name)
		installation.Apply(inputInstallation)

		log.Info("Creating a new installation", attribute.String("installation", installation.String()))
	} else {
		// Apply the specified changes to the installation
		installation.Apply(inputInstallation)
		if err := installation.Validate(); err != nil {
			return err
		}

		fmt.Fprintf(p.Err, "Updating %s installation\n", installation)
	}

	reconcileOpts := ReconcileOptions{
		Namespace:    input.Namespace,
		Name:         input.Name,
		Installation: installation,
		Force:        opts.Force,
		DryRun:       opts.DryRun,
	}
	return p.ReconcileInstallation(ctx, reconcileOpts)
}
