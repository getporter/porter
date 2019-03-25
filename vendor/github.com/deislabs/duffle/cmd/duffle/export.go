package main

import (
	"fmt"
	"io"

	"github.com/deislabs/duffle/pkg/duffle/home"
	"github.com/deislabs/duffle/pkg/loader"
	"github.com/deislabs/duffle/pkg/packager"

	"github.com/spf13/cobra"
)

const exportDesc = `
Packages a bundle, invocation images, and all referenced images within
a single gzipped tarfile. All images specified in the bundle metadata
are saved as tar files in the artifacts/ directory along with an 
artifacts.json file which describes the contents of artifacts/ directory.

By default, this command will use the name and version information of
the bundle to create a compressed archive file called
<name>-<version>.tgz in the current directory. This destination can be
updated by specifying a file path to save the compressed bundle to using
the --output-file flag.

Use the --thin flag to export the bundle manifest without the invocation
images and referenced images.

Pass in a path to a bundle file instead of a bundle in local storage by
using the --bundle-is-file flag like below:
$ duffle export [PATH] --bundle-is-file
`

type exportCmd struct {
	bundle       string
	dest         string
	home         home.Home
	out          io.Writer
	thin         bool
	verbose      bool
	insecure     bool
	bundleIsFile bool
}

func newExportCmd(w io.Writer) *cobra.Command {
	export := &exportCmd{out: w}

	cmd := &cobra.Command{
		Use:   "export [BUNDLE]",
		Short: "package CNAB bundle in gzipped tar file",
		Long:  exportDesc,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			export.home = home.Home(homePath())
			export.bundle = args[0]

			return export.run()
		},
	}

	f := cmd.Flags()
	f.StringVarP(&export.dest, "output-file", "o", "", "Save exported bundle to file path")
	f.BoolVarP(&export.bundleIsFile, "bundle-is-file", "f", false, "Indicates that the bundle source is a file path")
	f.BoolVarP(&export.thin, "thin", "t", false, "Export only the bundle manifest")
	f.BoolVarP(&export.verbose, "verbose", "v", false, "Verbose output")
	f.BoolVarP(&export.insecure, "insecure", "k", false, "Do not verify the bundle (INSECURE)")

	return cmd
}

func (ex *exportCmd) run() error {
	bundlefile, l, err := ex.setup()
	if err != nil {
		return err
	}
	if err := ex.Export(bundlefile, l); err != nil {
		return err
	}

	return nil
}

func (ex *exportCmd) Export(bundlefile string, l loader.Loader) error {
	exp, err := packager.NewExporter(bundlefile, ex.dest, ex.home.Logs(), l, ex.thin, ex.insecure)
	if err != nil {
		return fmt.Errorf("Unable to set up exporter: %s", err)
	}
	if err := exp.Export(); err != nil {
		return err
	}
	if ex.verbose {
		fmt.Fprintf(ex.out, "Export logs: %s\n", exp.Logs)
	}
	return nil
}

func (ex *exportCmd) setup() (string, loader.Loader, error) {
	bundlefile, err := resolveBundleFilePath(ex.bundle, ex.home.String(), ex.bundleIsFile, ex.insecure)
	if err != nil {
		return "", nil, err
	}

	l, err := getLoader(ex.home.String(), ex.insecure)
	if err != nil {
		return "", nil, err
	}

	return bundlefile, l, nil
}

func resolveBundleFilePath(bun, homePath string, bundleIsFile, insecure bool) (string, error) {

	if bundleIsFile {
		return bun, nil
	}

	bundlefile, err := getBundleFilepath(bun, homePath, insecure)
	if err != nil {
		return "", err
	}
	return bundlefile, err
}
