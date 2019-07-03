package porter

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/deislabs/porter/pkg/printer"

	dtprinter "github.com/carolynvs/datetime-printer"
	"github.com/pkg/errors"
)

type OutputListOptions struct {
	RawFormat string
	Format    printer.Format
	Bundle    string
}

type OutputListing struct {
	Name     string
	Modified time.Time
}

type OutputShowOptions struct {
	RawFormat      string
	Format         printer.Format
	Output, Bundle string
}

type Output struct {
	Name, Bundle, Contents string
}

// OutputList is a slice of OutputListings
type OutputList []OutputListing

func (l OutputList) Len() int {
	return len(l)
}
func (l OutputList) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}
func (l OutputList) Less(i, j int) bool {
	return l[i].Modified.Before(l[j].Modified)
}

var err error

// Validate prepares for an action and validates the options.
func (o *OutputListOptions) Validate() error {
	if o.Bundle == "" {
		return errors.New("a bundle name must be supplied")
	}

	o.Format, err = printer.ParseFormat(o.RawFormat)
	if err != nil {
		return err
	}

	return nil
}

// TODO: testssssssss
func (p *Porter) ListBundleOutputs(opts OutputListOptions) error {
	outputList, err := p.listBundleOutputs(opts.Bundle)
	if err != nil {
		return errors.Wrapf(err, "unable to list outputs for bundle %s", opts.Bundle)
	}

	switch opts.Format {
	case printer.FormatJson:
		return printer.PrintJson(p.Out, outputList)
	case printer.FormatYaml:
		return printer.PrintYaml(p.Out, outputList)
	case printer.FormatTable:
		// have every row use the same "now" starting ... NOW!
		now := time.Now()
		tp := dtprinter.DateTimePrinter{
			Now: func() time.Time { return now },
		}

		printOutputRow :=
			func(v interface{}) []interface{} {
				ol, ok := v.(OutputListing)
				if !ok {
					return nil
				}
				return []interface{}{ol.Name, tp.Format(ol.Modified)}
			}
		return printer.PrintTable(p.Out, *outputList, printOutputRow,
			"NAME", "MODIFIED")
	default:
		return fmt.Errorf("invalid format: %s", opts.Format)
	}
}

func (p *Porter) listBundleOutputs(bundle string) (*OutputList, error) {
	outputsDir, err := p.Config.GetOutputsDir()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get outputs directory")
	}
	bundleOutputsDir := filepath.Join(outputsDir, bundle)

	var outputList OutputList
	// Walk through bundleOutputsDir, if exists, and read all output filenames.
	// We don't want to present the actual values (as they may be large)
	// The `show` command should be used to show the value for a given output
	if ok, _ := p.Context.FileSystem.DirExists(bundleOutputsDir); ok {
		p.Context.FileSystem.Walk(bundleOutputsDir, func(path string, info os.FileInfo, err error) error {
			if !info.IsDir() {
				outputListing := OutputListing{
					Name:     info.Name(),
					Modified: info.ModTime(),
				}
				outputList = append(outputList, outputListing)
			}
			return nil
		})
		sort.Sort(sort.Reverse(outputList))
	} else {
		return nil, errors.New("bundle not found")
	}
	return &outputList, nil
}

// Validate prepares for an action and validates the options.
func (o *OutputShowOptions) Validate() error {
	if o.Bundle == "" {
		return errors.New("a bundle name must be supplied")
	}
	if o.Output == "" {
		return errors.New("an output name must be supplied")
	}

	o.Format, err = printer.ParseFormat(o.RawFormat)
	if err != nil {
		return err
	}

	return nil
}

func (p *Porter) ShowBundleOutput(opts OutputShowOptions) error {
	contents, err := p.readBundleOutput(opts.Output, opts.Bundle)
	if err != nil {
		return errors.Wrapf(err, "unable to read output %s for bundle %s", opts.Output, opts.Bundle)
	}

	output := Output{
		Name:     opts.Output,
		Bundle:   opts.Bundle,
		Contents: string(contents),
	}

	switch opts.Format {
	case printer.FormatJson:
		return printer.PrintJson(p.Out, output)
	case printer.FormatYaml:
		return printer.PrintYaml(p.Out, output)
	case printer.FormatPlaintext:
		fmt.Fprintln(p.Out, output.Contents)
		return nil
	default:
		return fmt.Errorf("invalid format: %s", opts.Format)
	}
}

func (p *Porter) readBundleOutput(Output, Bundle string) ([]byte, error) {
	outputsDir, err := p.Config.GetOutputsDir()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get outputs directory")
	}
	bundleOutputsDir := filepath.Join(outputsDir, Bundle)

	outputPath := filepath.Join(bundleOutputsDir, Output)

	return p.Context.FileSystem.ReadFile(outputPath)
}
