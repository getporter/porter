package porter

import (
	"fmt"
	"strings"

	"get.porter.sh/porter/pkg/mixin"

	"get.porter.sh/porter/pkg/pkgmgmt"
	"get.porter.sh/porter/pkg/pkgmgmt/feed"
	"get.porter.sh/porter/pkg/printer"
	"github.com/gobuffalo/packr/v2"
)

// PrintMixinsOptions represent options for the PrintMixins function
type PrintMixinsOptions struct {
	printer.PrintOptions
}

func (p *Porter) PrintMixins(opts PrintMixinsOptions) error {
	mixins, err := p.ListMixins()
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
				return []interface{}{m.Name, m.VersionInfo.Version, m.VersionInfo.Author}
			}
		return printer.PrintTable(p.Out, mixins, printMixinRow, "Name", "Version", "Author")
	case printer.FormatJson:
		return printer.PrintJson(p.Out, mixins)
	case printer.FormatYaml:
		return printer.PrintYaml(p.Out, mixins)
	default:
		return fmt.Errorf("invalid format: %s", opts.Format)
	}
}

func (p *Porter) ListMixins() ([]mixin.Metadata, error) {
	// List out what is installed on the file system
	names, err := p.Mixins.List()
	if err != nil {
		return nil, err
	}

	// Query each mixin and fill out their metadata
	mixins := make([]mixin.Metadata, len(names))
	for i, name := range names {
		m, err := p.Mixins.GetMetadata(name)
		if err != nil {
			fmt.Fprintf(p.Err, "could not get version from mixin %s: %s\n ", name, err.Error())
			continue
		}

		meta, _ := m.(*mixin.Metadata)
		mixins[i] = *meta
	}

	return mixins, nil
}

// SearchMixins searches the internal remote mixins list according to the provided options
func (p *Porter) SearchMixins(opts mixin.SearchOptions) error {
	box := packr.New("get.porter.sh/porter/pkg/mixin/remote-mixins", "../mixin/remote-mixins")
	mixinSearcher := mixin.NewSearcher(box)

	remoteMixins, err := mixinSearcher.Search(opts)
	if err != nil {
		return err
	}

	switch opts.Format {
	case printer.FormatTable:
		printMixinRow :=
			func(v interface{}) []interface{} {
				m, ok := v.(pkgmgmt.PackageListing)
				if !ok {
					return nil
				}

				var urlType string
				if strings.Contains(m.URL, ".xml") {
					urlType = "Atom Feed"
				} else if strings.Contains(m.URL, "download") {
					urlType = "Download"
				} else {
					urlType = "Unknown"
				}
				return []interface{}{m.Name, m.Description, m.Author, m.URL, urlType}
			}
		return printer.PrintTable(p.Out, remoteMixins, printMixinRow, "Name", "Description", "Author", "URL", "URL Type")
	case printer.FormatJson:
		return printer.PrintJson(p.Out, remoteMixins)
	case printer.FormatYaml:
		return printer.PrintYaml(p.Out, remoteMixins)
	default:
		return fmt.Errorf("invalid format: %s", opts.Format)
	}
}

func (p *Porter) InstallMixin(opts mixin.InstallOptions) error {
	err := p.Mixins.Install(opts.InstallOptions)
	if err != nil {
		return err
	}

	mixin, err := p.Mixins.GetMetadata(opts.Name)
	if err != nil {
		return err
	}

	v := mixin.GetVersionInfo()
	fmt.Fprintf(p.Out, "installed %s mixin %s (%s)\n", opts.Name, v.Version, v.Commit)

	return nil
}

func (p *Porter) UninstallMixin(opts pkgmgmt.UninstallOptions) error {
	err := p.Mixins.Uninstall(opts)
	if err != nil {
		return err
	}

	fmt.Fprintf(p.Out, "Uninstalled %s mixin", opts.Name)

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
