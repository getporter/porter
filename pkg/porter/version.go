package porter

import (
	"fmt"
	"github.com/deislabs/porter/pkg"
	"github.com/deislabs/porter/pkg/mixin"
	"github.com/deislabs/porter/pkg/porter/version"
	"github.com/deislabs/porter/pkg/printer"
	"github.com/pkg/errors"
	"io"
	"runtime"
)

type SystemInfo struct {
	OS string
	Arch string
}

func (p *Porter) PrintVersion(opts version.Options) error {
	metadata := mixin.Metadata{
		Name: "porter",
		VersionInfo: mixin.VersionInfo{
			Version: pkg.Version,
			Commit:  pkg.Commit,
		},
	}
	return version.PrintVersion(p.Context, opts, metadata)
}

func getSystemInfo() *SystemInfo {
	return &SystemInfo{
		OS:            runtime.GOOS,
		Arch:          runtime.GOARCH,
	}
}

func (system *SystemInfo) printSystemInfo(out io.Writer) error {
	_, err := fmt.Fprintf(out, "os: %s\n", system.OS)
	_, err = fmt.Fprintf(out, "arch: %s\n", system.Arch)
	return err
}

func printSectionHeader(out io.Writer, header string){
	underline := "-------"
	_, _ = fmt.Fprintf(out, "\n%s\n%s\n", header, underline)

}

func (p *Porter) PrintDebugInfo(opts version.Options) error {
	// force opts to print version as plaintext
	opts.RawFormat = string(printer.FormatPlaintext)
	err := p.PrintVersion(opts)
	if err != nil {
		return errors.Wrap(err, "Failed to print version")
	}

	printSectionHeader(p.Context.Out, "System")
	systemInfo := getSystemInfo()
	err = systemInfo.printSystemInfo(p.Out)
	if err != nil {
		return errors.Wrap(err, "Failed to print system information")
	}
	printSectionHeader(p.Context.Out, "Mixins")

	mixins, err := p.ListMixins()
	if err != nil {
		return err
	}
	printMixinRow :=
		func(v interface{}) []interface{} {
			m, ok := v.(mixin.Metadata)
			if !ok {
				return nil
			}
			return []interface{}{m.Name, m.VersionInfo.Version, m.VersionInfo.Author}
		}
	err = printer.PrintTable(p.Out, mixins, printMixinRow, "Name", "Version", "Author")
	return err
}

