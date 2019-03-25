package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/deislabs/duffle/pkg/duffle/home"
	"github.com/deislabs/duffle/pkg/repo"

	"github.com/Masterminds/semver"
	"github.com/spf13/cobra"
)

const bundleRemoveDesc = `Remove a bundle from the local storage.

This removes a bundle from the local storage so that it will no longer be locally
available. Bundles can be rebuilt with 'duffle build'.

Ex. $ duffle bundle remove foo  # removes all versions of foo from local store

If a SemVer range is provided with '--version'/'-r' then only releases that match
that range will be removed.
`

type bundleRemoveCmd struct {
	bundleRef string
	home      home.Home
	out       io.Writer
	versions  string
}

func newBundleRemoveCmd(w io.Writer) *cobra.Command {
	remove := &bundleRemoveCmd{out: w}

	cmd := &cobra.Command{
		Use:     "remove [BUNDLE]",
		Aliases: []string{"rm"},
		Short:   "remove a bundle from the local storage",
		Long:    bundleRemoveDesc,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			remove.bundleRef = args[0]
			remove.home = home.Home(homePath())
			return remove.run()
		},
	}
	cmd.Flags().StringVar(&remove.versions, "version", "", "A version or SemVer2 version range")

	return cmd
}

func (rm *bundleRemoveCmd) run() error {
	index, err := repo.LoadIndex(rm.home.Repositories())
	if err != nil {
		return err
	}

	vers, ok := index.GetVersions(rm.bundleRef)
	if !ok {
		fmt.Fprintf(rm.out, "Bundle %q not found. Nothing deleted.", rm.bundleRef)
		return nil
	}

	// If versions is set, we short circuit and only delete specific versions.
	if rm.versions != "" {
		fmt.Fprintln(rm.out, "Only deleting versions")
		matcher, err := semver.NewConstraint(rm.versions)
		if err != nil {
			return err
		}
		deletions := []repo.BundleVersion{}
		for _, ver := range vers {
			if ok, _ := matcher.Validate(ver.Version); ok {
				fmt.Fprintf(rm.out, "Version %s matches constraint %q\n", ver, rm.versions)
				deletions = append(deletions, ver)
				index.DeleteVersion(rm.bundleRef, ver.Version.String())
				// If there are no more versions, remove the entire entry.
				if vers, ok := index.GetVersions(rm.bundleRef); ok && len(vers) == 0 {
					index.Delete(rm.bundleRef)
				}

			}
		}

		if len(deletions) == 0 {
			return nil
		}
		if err := index.WriteFile(rm.home.Repositories(), 0644); err != nil {
			return err
		}
		deleteBundleVersions(deletions, index, rm.home, rm.out)
		return nil
	}

	// If no version was specified, delete entire record
	if !index.Delete(rm.bundleRef) {
		fmt.Fprintf(rm.out, "Bundle %q not found. Nothing deleted.", rm.bundleRef)
		return nil
	}
	if err := index.WriteFile(rm.home.Repositories(), 0644); err != nil {
		return err
	}

	deleteBundleVersions(vers, index, rm.home, rm.out)
	return nil
}

// deleteBundleVersions removes the given SHAs from bundle storage
//
// It warns, but does not fail, if a given SHA is not found.
func deleteBundleVersions(vers []repo.BundleVersion, index repo.Index, h home.Home, w io.Writer) {
	for _, ver := range vers {
		fpath := filepath.Join(h.Bundles(), ver.Digest)
		if err := os.Remove(fpath); err != nil {
			fmt.Fprintf(w, "WARNING: could not delete stake record %q", fpath)
		}
	}
}
