package porter

import (
	"fmt"
	"strings"

	"github.com/deislabs/cnab-go/bundle"
	"github.com/deislabs/porter/pkg/context"
	"github.com/deislabs/porter/pkg/printer"
	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
)

type ExplainOpts struct {
	BundleLifecycleOpts
	printer.PrintOptions
}

// PrintableBundle holds a subset of pertinent values to be explained from a bundle.Bundle
type PrintableBundle struct {
	Name        string                         `json:"name" yaml:"name"`
	Description string                         `json:"description,omitempty" yaml:"description,omitempty"`
	Version     string                         `json:"version" yaml:"version"`
	Parameters  map[string]PrintableParameter  `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	Credentials map[string]PrintableCredential `json:"credentials,omitempty" yaml:"credentials,omitempty"`
	Outputs     map[string]PrintableOutput     `json:"outputs,omitempty" yaml:"outputs,omitempty"`
	Actions     map[string]PrintableAction     `json:"customActions,omitempty" yaml:"customActions,omitempty"`
}

type PrintableCredential struct {
	Description string `json:"description" yaml:"description"`
	Required    bool   `json:"required" yaml:"required"`
}

type PrintableOutput struct {
	Type        interface{} `json:"type" yaml:"type"`
	ApplyTo     string      `json:"applyTo" yaml:"applyTo"`
	Description string      `json:"description" yaml:"description"`
}

type PrintableParameter struct {
	Type        interface{} `json:"type" yaml:"type"`
	Default     interface{} `json:"default" yaml:"default"`
	ApplyTo     string      `json:"applyTo" yaml:"applyTo"`
	Description string      `json:"description" yaml:"description"`
	Required    bool        `json:"required" yaml:"required"`
}

type PrintableAction struct {
	Modifies bool `json:"modifies" yaml:"modifies"`
	// Stateless indicates that the action is purely informational, that credentials are not required, and that the runtime should not keep track of its invocation
	Stateless bool `json:"stateless" yaml:"stateless"`
	// Description describes the action as a user-readable string
	Description string `json:"description" yaml:"description"`
}

func (o *ExplainOpts) Validate(args []string, cxt *context.Context) error {
	err := o.sharedOptions.Validate(args, cxt)
	if err != nil {
		return err
	}
	err = o.ParseFormat()
	if err != nil {
		return err
	}
	if o.Tag != "" {
		o.File = ""
		o.CNABFile = ""

		return o.validateTag()
	}
	return nil
}

func (p *Porter) Explain(o ExplainOpts) error {
	err := p.prepullBundleByTag(&o.BundleLifecycleOpts)
	if err != nil {
		return errors.Wrap(err, "unable to pull bundle before invoking credentials generate")
	}

	err = p.applyDefaultOptions(&o.sharedOptions)
	if err != nil {
		return err
	}
	err = p.ensureLocalBundleIsUpToDate(o.bundleFileOptions)
	if err != nil {
		return err
	}
	bundle, err := p.CNAB.LoadBundle(o.CNABFile, o.Insecure)
	// Print Bundle Details

	pb, err := generatePrintable(bundle)
	if err != nil {
		return errors.Wrap(err, "unable to print bundle")
	}
	switch o.Format {
	case printer.FormatJson:
		return printer.PrintJson(p.Out, pb)
	case printer.FormatYaml:
		return printer.PrintYaml(p.Out, pb)
	case printer.FormatTable:
		return p.printTable(pb)
	default:
		return fmt.Errorf("invalid format: %s", o.Format)
	}
}

func generatePrintable(bun *bundle.Bundle) (*PrintableBundle, error) {
	if bun == nil {
		return nil, fmt.Errorf("expected a bundle")
	}
	pb := PrintableBundle{
		Name:        bun.Name,
		Description: bun.Description,
		Version:     bun.Version,
	}

	actions := map[string]PrintableAction{}
	for a, v := range bun.Actions {
		pa := PrintableAction{}
		pa.Description = v.Description
		pa.Modifies = v.Modifies
		pa.Stateless = v.Stateless
		actions[a] = pa
	}

	creds := map[string]PrintableCredential{}
	for c, v := range bun.Credentials {
		pc := PrintableCredential{}
		pc.Description = v.Description
		pc.Required = v.Required

		creds[c] = pc
	}

	params := map[string]PrintableParameter{}
	for p, v := range bun.Parameters {
		def, ok := bun.Definitions[v.Definition]
		if !ok {
			return nil, fmt.Errorf("unable to find definition %s", v.Definition)
		}
		if def == nil {
			return nil, fmt.Errorf("empty definition for %s", v.Definition)
		}
		pp := PrintableParameter{}
		pp.Type = def.Type
		pp.Default = def.Default
		pp.ApplyTo = generateApplyToString(v.ApplyTo)
		pp.Required = v.Required
		pp.Description = v.Description

		params[p] = pp
	}

	outputs := map[string]PrintableOutput{}
	for p, v := range bun.Outputs {
		def, ok := bun.Definitions[v.Definition]
		if !ok {
			return nil, fmt.Errorf("unable to find definition %s", v.Definition)
		}
		if def == nil {
			return nil, fmt.Errorf("empty definition for %s", v.Definition)
		}
		pd := PrintableOutput{}
		pd.Type = def.Type
		pd.ApplyTo = generateApplyToString(v.ApplyTo)
		pd.Description = v.Description

		outputs[p] = pd
	}

	pb.Actions = actions
	pb.Credentials = creds
	pb.Outputs = outputs
	pb.Parameters = params

	return &pb, nil
}

func (p *Porter) printTable(bun *PrintableBundle) error {
	fmt.Fprintf(p.Out, "Name: %s\n", bun.Name)
	fmt.Fprintf(p.Out, "Description: %s\n", bun.Description)
	fmt.Fprintf(p.Out, "Version: %s\n", bun.Version)
	fmt.Fprintln(p.Out, "")

	p.generateCredentialsBlock(bun)

	p.generateParametersBlock(bun)

	p.generateOutputsBlock(bun)

	p.generateActionsBlock(bun)

	return nil
}

func (p *Porter) generateCredentialsBlock(bun *PrintableBundle) {
	if len(bun.Credentials) > 0 {
		fmt.Fprintln(p.Out, "Credentials:")
		p.generateCredentialsTable(bun)
	} else {
		fmt.Fprintln(p.Out, "No credentials defined")
	}
	fmt.Fprintln(p.Out, "") // force a blank line after this block
}
func (p *Porter) generateCredentialsTable(bun *PrintableBundle) {
	h := []string{"Name", "Description", "Required"}
	d := [][]string{}
	for k, v := range bun.Credentials {
		d = append(d, []string{k, v.Description, fmt.Sprintf("%v", v.Required)})
	}
	p.generateTable(h, d)
}

func (p *Porter) generateParametersBlock(bun *PrintableBundle) {
	if len(bun.Parameters) > 0 {
		fmt.Fprintln(p.Out, "Parameters:")
		p.generateParametersTable(bun)
	} else {
		fmt.Fprintln(p.Out, "No parameters defined")
	}
	fmt.Fprintln(p.Out, "") // force a blank line after this block
}
func (p *Porter) generateParametersTable(bun *PrintableBundle) {
	h := []string{"Name", "Description", "Type", "Default", "Required", "Applies To"}
	d := [][]string{}
	for k, v := range bun.Parameters {
		d = append(d, []string{
			k, v.Description,
			fmt.Sprintf("%v", v.Type),
			fmt.Sprintf("%v", v.Default),
			fmt.Sprintf("%v", v.Required),
			v.ApplyTo,
		})
	}
	p.generateTable(h, d)
}

func (p *Porter) generateOutputsBlock(bun *PrintableBundle) {
	if len(bun.Outputs) > 0 {
		fmt.Fprintln(p.Out, "Outputs:")
		p.generateOutputsTable(bun)
	} else {
		fmt.Fprintln(p.Out, "No outputs defined")
	}
	fmt.Fprintln(p.Out, "") // force a blank line after this block
}

func (p *Porter) generateOutputsTable(bun *PrintableBundle) {
	h := []string{"Name", "Description", "Type", "Applies To"}
	d := [][]string{}
	for k, v := range bun.Outputs {
		d = append(d, []string{k, v.Description, fmt.Sprintf("%v", v.Type), v.ApplyTo})
	}
	p.generateTable(h, d)
}

func generateApplyToString(appliesTo []string) string {
	if len(appliesTo) == 0 {
		return "All Actions"
	} else {
		return strings.Join(appliesTo, ",")
	}
}

func (p *Porter) generateActionsBlock(bun *PrintableBundle) {
	if len(bun.Actions) > 0 {
		fmt.Fprintln(p.Out, "Actions:")
		p.generateActionsTable(bun)
	} else {
		fmt.Fprintln(p.Out, "No custom actions defined")
	}
	fmt.Fprintln(p.Out, "") // force a blank line after this block
}

func (p *Porter) generateActionsTable(bun *PrintableBundle) {
	h := []string{"Name", "Description", "Modifies Instance", "Stateless"}
	d := [][]string{}
	for k, v := range bun.Actions {
		d = append(d, []string{k, v.Description, fmt.Sprintf("%v", v.Modifies), fmt.Sprintf("%v", v.Stateless)})
	}
	p.generateTable(h, d)
}

func (p *Porter) generateTable(h []string, d [][]string) {
	t := tablewriter.NewWriter(p.Out)
	t.SetAlignment(tablewriter.ALIGN_LEFT)
	t.SetHeader(h)
	t.AppendBulk(d)
	t.Render()
}
