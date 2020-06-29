package porter

import (
	"fmt"
	"sort"
	"strings"

	"get.porter.sh/porter/pkg/context"
	"get.porter.sh/porter/pkg/printer"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/pkg/errors"
)

type ExplainOpts struct {
	BundleLifecycleOpts
	printer.PrintOptions
}

// PrintableBundle holds a subset of pertinent values to be explained from a bundle.Bundle
type PrintableBundle struct {
	Name        string                `json:"name" yaml:"name"`
	Description string                `json:"description,omitempty" yaml:"description,omitempty"`
	Version     string                `json:"version" yaml:"version"`
	Parameters  []PrintableParameter  `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	Credentials []PrintableCredential `json:"credentials,omitempty" yaml:"credentials,omitempty"`
	Outputs     []PrintableOutput     `json:"outputs,omitempty" yaml:"outputs,omitempty"`
	Actions     []PrintableAction     `json:"customActions,omitempty" yaml:"customActions,omitempty"`
}

type PrintableCredential struct {
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description" yaml:"description"`
	Required    bool   `json:"required" yaml:"required"`
}

type SortPrintableCredential []PrintableCredential

func (s SortPrintableCredential) Len() int {
	return len(s)
}

func (s SortPrintableCredential) Less(i, j int) bool {
	return s[i].Name < s[j].Name
}

func (s SortPrintableCredential) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

type PrintableOutput struct {
	Name        string      `json:"name" yaml:"name"`
	Type        interface{} `json:"type" yaml:"type"`
	ApplyTo     string      `json:"applyTo" yaml:"applyTo"`
	Description string      `json:"description" yaml:"description"`
}

type SortPrintableOutput []PrintableOutput

func (s SortPrintableOutput) Len() int {
	return len(s)
}

func (s SortPrintableOutput) Less(i, j int) bool {
	return s[i].Name < s[j].Name
}

func (s SortPrintableOutput) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

type PrintableParameter struct {
	Name        string      `json:"name" yaml:"name"`
	Type        interface{} `json:"type" yaml:"type"`
	Default     interface{} `json:"default" yaml:"default"`
	ApplyTo     string      `json:"applyTo" yaml:"applyTo"`
	Description string      `json:"description" yaml:"description"`
	Required    bool        `json:"required" yaml:"required"`
}

type SortPrintableParameter []PrintableParameter

func (s SortPrintableParameter) Len() int {
	return len(s)
}

func (s SortPrintableParameter) Less(i, j int) bool {
	return s[i].Name < s[j].Name
}

func (s SortPrintableParameter) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

type PrintableAction struct {
	Name     string `json:"name" yaml:"name"`
	Modifies bool   `json:"modifies" yaml:"modifies"`
	// Stateless indicates that the action is purely informational, that credentials are not required, and that the runtime should not keep track of its invocation
	Stateless bool `json:"stateless" yaml:"stateless"`
	// Description describes the action as a user-readable string
	Description string `json:"description" yaml:"description"`
}

type SortPrintableAction []PrintableAction

func (s SortPrintableAction) Len() int {
	return len(s)
}

func (s SortPrintableAction) Less(i, j int) bool {
	return s[i].Name < s[j].Name
}

func (s SortPrintableAction) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Using empty structs instead of booleans cuz empty structs occupy zero memory
var porterInternalParams = map[string]struct{}{"porter-debug": {}}

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
		return errors.Wrap(err, "unable to pull bundle before invoking explain command")
	}

	err = p.applyDefaultOptions(&o.sharedOptions)
	if err != nil {
		return err
	}
	err = p.ensureLocalBundleIsUpToDate(o.bundleFileOptions)
	if err != nil {
		return err
	}
	bundle, err := p.CNAB.LoadBundle(o.CNABFile)
	// Print Bundle Details

	pb, err := generatePrintable(bundle)
	if err != nil {
		return errors.Wrap(err, "unable to print bundle")
	}
	return p.printBundleExplain(o, pb)
}

func (p *Porter) printBundleExplain(o ExplainOpts, pb *PrintableBundle) error {
	switch o.Format {
	case printer.FormatJson:
		return printer.PrintJson(p.Out, pb)
	case printer.FormatYaml:
		return printer.PrintYaml(p.Out, pb)
	case printer.FormatTable:
		return p.printBundleExplainTable(pb)
	default:
		return fmt.Errorf("invalid format: %s", o.Format)
	}
}

func isInternalParam(inputParam string) bool {
	_, ok := porterInternalParams[inputParam]
	return ok
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

	actions := []PrintableAction{}
	for a, v := range bun.Actions {
		pa := PrintableAction{}
		pa.Name = a
		pa.Description = v.Description
		pa.Modifies = v.Modifies
		pa.Stateless = v.Stateless
		actions = append(actions, pa)
	}
	sort.Sort(SortPrintableAction(actions))

	creds := []PrintableCredential{}
	for c, v := range bun.Credentials {
		pc := PrintableCredential{}
		pc.Name = c
		pc.Description = v.Description
		pc.Required = v.Required

		creds = append(creds, pc)
	}
	sort.Sort(SortPrintableCredential(creds))

	params := []PrintableParameter{}
	for p, v := range bun.Parameters {
		if isInternalParam(p) {
			continue
		}
		def, ok := bun.Definitions[v.Definition]
		if !ok {
			return nil, fmt.Errorf("unable to find definition %s", v.Definition)
		}
		if def == nil {
			return nil, fmt.Errorf("empty definition for %s", v.Definition)
		}
		pp := PrintableParameter{}
		pp.Name = p
		pp.Type = def.Type
		pp.Default = def.Default
		pp.ApplyTo = generateApplyToString(v.ApplyTo)
		pp.Required = v.Required
		pp.Description = v.Description

		params = append(params, pp)
	}
	sort.Sort(SortPrintableParameter(params))

	outputs := []PrintableOutput{}
	for o, v := range bun.Outputs {
		def, ok := bun.Definitions[v.Definition]
		if !ok {
			return nil, fmt.Errorf("unable to find definition %s", v.Definition)
		}
		if def == nil {
			return nil, fmt.Errorf("empty definition for %s", v.Definition)
		}
		po := PrintableOutput{}
		po.Name = o
		po.Type = def.Type
		po.ApplyTo = generateApplyToString(v.ApplyTo)
		po.Description = v.Description

		outputs = append(outputs, po)
	}
	sort.Sort(SortPrintableOutput(outputs))

	pb.Actions = actions
	pb.Credentials = creds
	pb.Outputs = outputs
	pb.Parameters = params

	return &pb, nil
}

func generateApplyToString(appliesTo []string) string {
	if len(appliesTo) == 0 {
		return "All Actions"
	}
	return strings.Join(appliesTo, ",")

}

func (p *Porter) printBundleExplainTable(bun *PrintableBundle) error {
	fmt.Fprintf(p.Out, "Name: %s\n", bun.Name)
	fmt.Fprintf(p.Out, "Description: %s\n", bun.Description)
	fmt.Fprintf(p.Out, "Version: %s\n", bun.Version)
	fmt.Fprintln(p.Out, "")

	p.printCredentialsExplainBlock(bun)
	p.printParametersExplainBlock(bun)
	p.printOutputsExplainBlock(bun)
	p.printActionsExplainBlock(bun)

	return nil
}

func (p *Porter) printCredentialsExplainBlock(bun *PrintableBundle) error {
	if len(bun.Credentials) > 0 {
		fmt.Fprintln(p.Out, "Credentials:")
		err := p.printCredentialsExplainTable(bun)
		if err != nil {
			return errors.Wrap(err, "unable to print credentials table")
		}
	} else {
		fmt.Fprintln(p.Out, "No credentials defined")
	}
	fmt.Fprintln(p.Out, "") // force a blank line after this block
	return nil
}
func (p *Porter) printCredentialsExplainTable(bun *PrintableBundle) error {
	printCredRow :=
		func(v interface{}) []interface{} {
			c, ok := v.(PrintableCredential)
			if !ok {
				return nil
			}
			return []interface{}{c.Name, c.Description, c.Required}
		}
	return printer.PrintTable(p.Out, bun.Credentials, printCredRow, "Name", "Description", "Required")
}

func (p *Porter) printParametersExplainBlock(bun *PrintableBundle) error {
	if len(bun.Parameters) > 0 {
		fmt.Fprintln(p.Out, "Parameters:")
		err := p.printParametersExplainTable(bun)
		if err != nil {
			return errors.Wrap(err, "unable to print parameters table")
		}
	} else {
		fmt.Fprintln(p.Out, "No parameters defined")
	}
	fmt.Fprintln(p.Out, "") // force a blank line after this block
	return nil
}
func (p *Porter) printParametersExplainTable(bun *PrintableBundle) error {
	printParamRow :=
		func(v interface{}) []interface{} {
			p, ok := v.(PrintableParameter)
			if !ok {
				return nil
			}
			return []interface{}{p.Name, p.Description, p.Type, p.Default, p.Required, p.ApplyTo}
		}
	return printer.PrintTable(p.Out, bun.Parameters, printParamRow, "Name", "Description", "Type", "Default", "Required", "Applies To")
}

func (p *Porter) printOutputsExplainBlock(bun *PrintableBundle) error {
	if len(bun.Outputs) > 0 {
		fmt.Fprintln(p.Out, "Outputs:")
		err := p.printOutputsExplainTable(bun)
		if err != nil {
			return errors.Wrap(err, "unable to print outputs table")
		}
	} else {
		fmt.Fprintln(p.Out, "No outputs defined")
	}
	fmt.Fprintln(p.Out, "") // force a blank line after this block
	return nil
}

func (p *Porter) printOutputsExplainTable(bun *PrintableBundle) error {
	printOutputRow :=
		func(v interface{}) []interface{} {
			o, ok := v.(PrintableOutput)
			if !ok {
				return nil
			}
			return []interface{}{o.Name, o.Description, o.Type, o.ApplyTo}
		}
	return printer.PrintTable(p.Out, bun.Outputs, printOutputRow, "Name", "Description", "Type", "Applies To")
}

func (p *Porter) printActionsExplainBlock(bun *PrintableBundle) error {
	if len(bun.Actions) > 0 {
		fmt.Fprintln(p.Out, "Actions:")
		err := p.printActionsExplainTable(bun)
		if err != nil {
			return errors.Wrap(err, "unable to print actions block")
		}
	} else {
		fmt.Fprintln(p.Out, "No custom actions defined")
	}
	fmt.Fprintln(p.Out, "") // force a blank line after this block
	return nil
}

func (p *Porter) printActionsExplainTable(bun *PrintableBundle) error {
	printActionRow :=
		func(v interface{}) []interface{} {
			a, ok := v.(PrintableAction)
			if !ok {
				return nil
			}
			return []interface{}{a.Name, a.Description, a.Modifies, a.Stateless}
		}
	return printer.PrintTable(p.Out, bun.Actions, printActionRow, "Name", "Description", "Modifies Instance", "Stateless")
}
