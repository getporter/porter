package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/deislabs/duffle/pkg/bundle"
	"github.com/deislabs/duffle/pkg/crypto/digest"
	"github.com/deislabs/duffle/pkg/duffle/home"
	"github.com/deislabs/duffle/pkg/signature"
)

const bundleSignDesc = `Clear-sign a bundle.

This remarshals the bundle.json into canonical form, and then clear-signs the JSON.
By default, the signed bundle is written to $DUFFLE_HOME. You can specify an output-file to save to instead using the flag.

If no key name is supplied, this uses the first signing key in the secret keyring.
`

type bundleSignCmd struct {
	out            io.Writer
	home           home.Home
	identity       string
	bundleFile     string
	outfile        string
	skipValidation bool
}

func newBundleSignCmd(w io.Writer) *cobra.Command {
	sign := &bundleSignCmd{out: w}

	cmd := &cobra.Command{
		Use:   "sign BUNDLE",
		Short: "clear-sign a bundle",
		Args:  cobra.MaximumNArgs(1),
		Long:  bundleSignDesc,
		RunE: func(cmd *cobra.Command, args []string) error {
			sign.home = home.Home(homePath())
			secring := sign.home.SecretKeyRing()
			bundle, err := bundleFileOrArg1(args, sign.bundleFile)
			if err != nil {
				return err
			}
			return sign.signBundle(bundle, secring)
		},
	}
	cmd.Flags().StringVarP(&sign.identity, "user", "u", "", "the user ID of the key to use. Format is either email address or 'NAME (COMMENT) <EMAIL>'")
	cmd.Flags().StringVarP(&sign.bundleFile, "file", "f", "", "path to bundle file to sign")
	cmd.Flags().StringVarP(&sign.outfile, "output-file", "o", "", "the name of the output file")
	cmd.Flags().BoolVar(&sign.skipValidation, "skip-validate", false, "do not validate the JSON before marshaling it.")

	return cmd
}

func bundleFileOrArg1(args []string, bundle string) (string, error) {
	switch {
	case len(args) == 1 && bundle != "":
		return "", errors.New("please use either -f or specify a BUNDLE, but not both")
	case len(args) == 0 && bundle == "":
		return "", errors.New("please specify a BUNDLE or use -f for a file")
	case len(args) == 1:
		// passing insecure: true, as currently we can only sign an unsinged bundle
		return getBundleFilepath(args[0], homePath(), true)
	}
	return bundle, nil
}

func (bs *bundleSignCmd) signBundle(bundleFile, keyring string) error {
	// Verify that file exists
	if fi, err := os.Stat(bundleFile); err != nil {
		return fmt.Errorf("cannot find bundle file to sign: %v", err)
	} else if fi.IsDir() {
		return errors.New("cannot sign a directory")
	}

	bdata, err := ioutil.ReadFile(bundleFile)
	if err != nil {
		return err
	}
	b, err := bundle.Unmarshal(bdata)
	if err != nil {
		return err
	}

	if !bs.skipValidation {
		if err := b.Validate(); err != nil {
			return err
		}
	}

	// Load keyring
	kr, err := signature.LoadKeyRing(keyring)
	if err != nil {
		return err
	}
	// Find identity
	var k *signature.Key
	if bs.identity != "" {
		k, err = kr.Key(bs.identity)
		if err != nil {
			return err
		}
	} else {
		all := kr.PrivateKeys()
		if len(all) == 0 {
			return errors.New("no private keys found")
		}
		k = all[0]
	}

	// Be sure userID is parseable before attempting to sign
	userID, err := k.UserID()
	if err != nil {
		return err
	}

	// Sign the file
	s := signature.NewSigner(k)
	data, err := s.Clearsign(b)
	if err != nil {
		return err
	}

	data = append(data, '\n')

	digest, err := digest.OfBuffer(data)
	if err != nil {
		return fmt.Errorf("cannot compute digest from bundle: %v", err)
	}

	// if --output-file is provided, write and return
	if bs.outfile != "" {
		if err := ioutil.WriteFile(bs.outfile, data, 0644); err != nil {
			return err
		}
	}

	if err := ioutil.WriteFile(filepath.Join(bs.home.Bundles(), digest), data, 0644); err != nil {
		return err
	}

	// TODO - write pkg method in bundle that writes file and records the reference
	if err := recordBundleReference(bs.home, b.Name, b.Version, digest); err != nil {
		return err
	}

	fmt.Fprintf(bs.out, "Signed by %s %s \n", userID.String(), k.Fingerprint())
	return nil
}
