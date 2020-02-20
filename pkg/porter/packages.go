package porter

import (
	"fmt"
	"strings"

	"github.com/gobuffalo/packr/v2"
	"github.com/pkg/errors"

	"get.porter.sh/porter/pkg/pkgmgmt"
	"get.porter.sh/porter/pkg/printer"
)

// SearchOptions are options for searching packages
type SearchOptions struct {
	Name string
	Type string
	List *packr.Box
	printer.PrintOptions
}

// Validate validates the arguments provided to a search command
func (o *SearchOptions) Validate(args []string) error {
	err := o.validatePackageName(args)
	if err != nil {
		return err
	}

	return o.ParseFormat()
}

// validatePackageName validates either no package name is provided or only one is
func (o *SearchOptions) validatePackageName(args []string) error {
	switch len(args) {
	case 0:
		return nil
	case 1:
		o.Name = strings.ToLower(args[0])
		return nil
	default:
		return errors.Errorf("only one positional argument may be specified, the package name, but multiple were received: %s", args)
	}
}

// SearchPackages searches the provided package list according to the provided options
func (p *Porter) SearchPackages(opts SearchOptions) error {
	pkgSearcher := pkgmgmt.NewSearcher(opts.List)

	mixinList, err := pkgSearcher.Search(opts.Name, opts.Type)
	if err != nil {
		return err
	}
	return p.PrintPackages(opts, mixinList)
}

// PrintPackages prints the provided package list according to the provided options
func (p *Porter) PrintPackages(opts SearchOptions, list pkgmgmt.PackageList) error {
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
		return printer.PrintTable(p.Out, list, printMixinRow, "Name", "Description", "Author", "URL", "URL Type")
	case printer.FormatJson:
		return printer.PrintJson(p.Out, list)
	case printer.FormatYaml:
		return printer.PrintYaml(p.Out, list)
	default:
		return fmt.Errorf("invalid format: %s", opts.Format)
	}
}
