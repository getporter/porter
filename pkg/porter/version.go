package porter

import (
	"bytes"
	"fmt"
	"runtime"
	"text/template"

	"get.porter.sh/porter/pkg"
	"get.porter.sh/porter/pkg/mixin"
	"get.porter.sh/porter/pkg/pkgmgmt"
	"get.porter.sh/porter/pkg/porter/version"
	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/pkg/printer"
	"github.com/pkg/errors"
)

type VersionOpts struct {
	version.Options
	System bool
}

type SystemInfo struct {
	OS   string
	Arch string
}

type Mixins []mixin.Metadata

type SystemDebugInfo struct {
	Version pkgmgmt.PackageMetadata `json:"version"`
	SysInfo SystemInfo              `json:"system"`
	Mixins  Mixins                  `json:"mixins"`
}

func (mixins Mixins) PrintMixinsTable() string {
	buffer := &bytes.Buffer{}
	printMixinRow :=
		func(v interface{}) []string {
			m, ok := v.(mixin.Metadata)
			if !ok {
				return nil
			}
			return []string{m.Name, m.VersionInfo.Version, m.VersionInfo.Author}
		}
	err := printer.PrintTable(buffer, mixins, printMixinRow, "Name", "Version", "Author")
	if err != nil {
		return ""
	}
	return buffer.String()
}

func (p *Porter) PrintVersion(opts VersionOpts) error {
	metadata := pkgmgmt.Metadata{
		Name: "porter",
		VersionInfo: pkgmgmt.VersionInfo{
			Version: pkg.Version,
			Commit:  pkg.Commit,
		},
	}

	if opts.System {
		return p.PrintDebugInfo(p.Context, opts, metadata)
	}

	return version.PrintVersion(p.Context, opts.Options, metadata)
}

func getSystemInfo() *SystemInfo {
	return &SystemInfo{
		OS:   runtime.GOOS,
		Arch: runtime.GOARCH,
	}
}

func (p *Porter) PrintDebugInfo(ctx *portercontext.Context, opts VersionOpts, metadata pkgmgmt.Metadata) error {
	opts.RawFormat = string(printer.FormatPlaintext)
	sysInfo := getSystemInfo()
	mixins, err := p.ListMixins()
	if err != nil {
		if p.Debug {
			fmt.Fprint(p.Err, err.Error())
		}
		return nil
	}

	sysDebugInfo := SystemDebugInfo{
		Version: metadata,
		SysInfo: *sysInfo,
		Mixins:  mixins,
	}

	switch opts.Format {
	case printer.FormatJson:
		return printer.PrintJson(ctx.Out, sysDebugInfo)
	case printer.FormatPlaintext:
		plaintextTmpl := `{{.Version.Name}} {{.Version.VersionInfo.Version}} ({{.Version.VersionInfo.Commit}})

System
-------
os: {{.SysInfo.OS}}
arch: {{.SysInfo.Arch}}
{{if .Mixins}}
Mixins
{{.Mixins.PrintMixinsTable}}{{end}}
`
		tmpl, err := template.New("systemDebugInfo").Parse(plaintextTmpl)
		if err != nil {
			return errors.Wrap(err, "Failed to parse plaintext template")
		}
		err = tmpl.Execute(ctx.Out, sysDebugInfo)
		return err
	default:
		return fmt.Errorf("unsupported format: %s", opts.Format)
	}
}
