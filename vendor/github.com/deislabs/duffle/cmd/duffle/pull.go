package main

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/deislabs/duffle/pkg/bundle"
	"github.com/deislabs/duffle/pkg/loader"
	"github.com/deislabs/duffle/pkg/reference"

	"github.com/spf13/cobra"
)

var ErrNotSigned = errors.New("bundle is not signed")

func newPullCmd(w io.Writer) *cobra.Command {
	const usage = `Pulls a CNAB bundle into the cache without installing it.

Example:
	$ duffle pull duffle/example:0.1.0
`

	var insecure bool
	cmd := &cobra.Command{
		Hidden: true,
		Use:    "pull",
		Short:  "pull a CNAB bundle from a repository",
		Long:   usage,
		RunE: func(cmd *cobra.Command, args []string) error {
			return ErrUnderConstruction
		},
	}

	cmd.Flags().BoolVarP(&insecure, "insecure", "k", false, "Do not verify the bundle (INSECURE)")

	return cmd
}

func getLoader(home string, insecure bool) (loader.Loader, error) {
	var load loader.Loader
	if insecure {
		load = loader.NewDetectingLoader()
	} else {
		kr, err := loadVerifyingKeyRings(home)
		if err != nil {
			return nil, fmt.Errorf("cannot securely load bundle: %s", err)
		}
		load = loader.NewSecureLoader(kr)
	}
	return load, nil
}

func getReference(bundleName string) (reference.NamedTagged, error) {
	var (
		name string
		ref  reference.NamedTagged
	)

	parts := strings.SplitN(bundleName, "://", 2)
	if len(parts) == 2 {
		name = parts[1]
	} else {
		name = parts[0]
	}
	normalizedRef, err := reference.ParseNormalizedNamed(name)
	if err != nil {
		return nil, fmt.Errorf("%q is not a valid bundle name: %v", name, err)
	}
	if reference.IsNameOnly(normalizedRef) {
		ref, err = reference.WithTag(normalizedRef, "latest")
		if err != nil {
			// NOTE(bacongobbler): Using the default tag *must* be valid.
			// To create a NamedTagged type with non-validated
			// input, the WithTag function should be used instead.
			panic(err)
		}
	} else {
		if taggedRef, ok := normalizedRef.(reference.NamedTagged); ok {
			ref = taggedRef
		} else {
			return nil, fmt.Errorf("unsupported image name: %s", normalizedRef.String())
		}
	}

	return ref, nil
}

func loadBundle(bundleFile string, insecure bool) (*bundle.Bundle, error) {
	l, err := getLoader(homePath(), insecure)
	if err != nil {
		return nil, err
	}
	// Issue #439: Errors that come back from the loader can be
	// pretty opaque.
	var bun *bundle.Bundle
	if bun, err = l.Load(bundleFile); err != nil {
		if err.Error() == "no signature block in data" {
			return bun, ErrNotSigned
		}
		// Dear Go, Y U NO TERNARY, kthxbye
		secflag := "secure"
		if insecure {
			secflag = "insecure"
		}
		return bun, fmt.Errorf("cannot load %s bundle: %s", secflag, err)
	}
	return bun, nil
}
