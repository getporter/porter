package porter

import (
	"bytes"
	"context"
	"fmt"
	"runtime"
	"text/template"

	"get.porter.sh/porter/pkg"
	"get.porter.sh/porter/pkg/mixin"
	"get.porter.sh/porter/pkg/pkgmgmt"
	"get.porter.sh/porter/pkg/porter/version"
	"get.porter.sh/porter/pkg/printer"
	"get.porter.sh/porter/pkg/tracing"
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

func (p *Porter) PrintVersion(ctx context.Context, opts VersionOpts) error {
	metadata := pkgmgmt.Metadata{
		Name: "porter",
		VersionInfo: pkgmgmt.VersionInfo{
			Version: pkg.Version,
			Commit:  pkg.Commit,
		},
	}

	if opts.System {
		return p.PrintDebugInfo(ctx, opts, metadata)
	}

	return version.PrintVersion(p.Context, opts.Options, metadata)
}

func getSystemInfo() *SystemInfo {
	return &SystemInfo{
		OS:   runtime.GOOS,
		Arch: runtime.GOARCH,
	}
}

func (p *Porter) PrintDebugInfo(ctx context.Context, opts VersionOpts, metadata pkgmgmt.Metadata) error {
	log := tracing.LoggerFromContext(ctx)

	opts.RawFormat = string(printer.FormatPlaintext)
	sysInfo := getSystemInfo()
	mixins, err := p.ListMixins(ctx)
	if err != nil {
		log.Debug(err.Error())
		return nil
	}

	sysDebugInfo := SystemDebugInfo{
		Version: metadata,
		SysInfo: *sysInfo,
		Mixins:  mixins,
	}

	switch opts.Format {
	case printer.FormatJson:
		return printer.PrintJson(p.Out, sysDebugInfo)
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
			return log.Error(fmt.Errorf("Failed to parse plaintext template: %w", err))
		}
		err = tmpl.Execute(p.Out, sysDebugInfo)
		return log.Error(err)
	default:
		return log.Error(fmt.Errorf("unsupported format: %s", opts.Format))
	}
}
