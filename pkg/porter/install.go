package porter

import (
	"fmt"

	"github.com/deislabs/porter/pkg/parameters"
)

type InstallOptions struct {
	// Name of the claim.
	Name string

	// File path to the CNAB bundle.
	File string

	// Insecure bundle installation allowed.
	Insecure bool

	// RawParams is the unparsed list of NAME=VALUE parameters set on the command line.
	RawParams []string

	// Params is the parsed set of parameters from RawParams.
	Params map[string]string

	// ParamFiles is a list of file paths containing lines of NAME=VALUE parameter definitions.
	ParamFiles []string

	// CredentialSets is a list of credentialset names to make available to the bundle.
	CredentialSets []string
}

func (o *InstallOptions) Prepare() error {
	return o.parseParams()
}

func (o *InstallOptions) parseParams() error {
	p, err := parameters.ParseVariableAssignments(o.RawParams)
	if err == nil {
		o.Params = p
	}
	return err
}

func (p *Porter) InstallBundle(opts InstallOptions) error {
	err := p.Config.LoadManifest()
	if err != nil {
		return err
	}
	fmt.Fprintf(p.Out, "installing %s...\n", p.Manifest.Name)
	return nil
}
