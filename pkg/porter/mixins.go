package porter

import (
	"fmt"

	"github.com/deislabs/porter/pkg/mixin"
	"github.com/deislabs/porter/pkg/mixin/feed"
	"github.com/deislabs/porter/pkg/printer"
)

// MixinProvider handles searching, listing and communicating with the mixins.
type MixinProvider interface {
	List() ([]mixin.Metadata, error)
	GetSchema(m mixin.Metadata) (string, error)
	GetVersion(m mixin.Metadata) (string, error)
	Install(opts mixin.InstallOptions) (mixin.Metadata, error)
}

func (p *Porter) PrintMixins(opts printer.PrintOptions) error {
	mixins, err := p.Mixins.List()
	if err != nil {
		return err
	}

	switch opts.Format {
	case printer.FormatTable:
		printMixinRow :=
			func(v interface{}) []interface{} {
				m, ok := v.(mixin.Metadata)
				if !ok {
					return nil
				}
				return []interface{}{m.Name}
			}
		return printer.PrintTable(p.Out, mixins, printMixinRow)
	case printer.FormatJson:
		return printer.PrintJson(p.Out, mixins)
	default:
		return fmt.Errorf("invalid format: %s", opts.Format)
	}
}

func (p *Porter) InstallMixin(opts mixin.InstallOptions) error {
	m, err := p.Mixins.Install(opts)
	if err != nil {
		return err
	}

	// TODO: Once we can extract the version from the mixin with json (#263), then we can print it out as installed mixin @v1.0.0
	confirmedVersion, err := p.Mixins.GetVersion(m)
	if err != nil {
		return err
	}
	if p.Debug {
		fmt.Fprintf(p.Out, "installed %s mixin to %s\n%s", m.Name, m.Dir, confirmedVersion)
	} else {
		fmt.Fprintf(p.Out, "installed %s mixin\n%s", m.Name, confirmedVersion)
	}

	return nil
}

func (p *Porter) GenerateMixinFeed(opts feed.GenerateOptions) error {
	f := feed.NewMixinFeed(p.Context)

	err := f.Generate(opts)
	if err != nil {
		return err
	}

	return f.Save(opts)
}

func (p *Porter) CreateMixinFeedTemplate() error {
	return feed.CreateTemplate(p.Context)
}
