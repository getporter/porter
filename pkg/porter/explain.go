package porter

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"get.porter.sh/porter/pkg/cnab"
	configadapter "get.porter.sh/porter/pkg/cnab/config-adapter"
	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/pkg/printer"
	"github.com/cnabio/cnab-go/bundle"
)

type ExplainOpts struct {
	BundleReferenceOptions
	printer.PrintOptions

	Action string
}

// PrintableBundle holds a subset of pertinent values to be explained from a bundle
type PrintableBundle struct {
	Name          string                `json:"name" yaml:"name"`
	Description   string                `json:"description,omitempty" yaml:"description,omitempty"`
	Version       string                `json:"version" yaml:"version"`
	PorterVersion string                `json:"porterVersion,omitempty" yaml:"porterVersion,omitempty"`
	Parameters    []PrintableParameter  `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	Credentials   []PrintableCredential `json:"credentials,omitempty" yaml:"credentials,omitempty"`
	Outputs       []PrintableOutput     `json:"outputs,omitempty" yaml:"outputs,omitempty"`
	Actions       []PrintableAction     `json:"customActions,omitempty" yaml:"customActions,omitempty"`
	Dependencies  []PrintableDependency `json:"dependencies,omitempty" yaml:"dependencies,omitempty"`
	Mixins        []string              `json:"mixins" yaml:"mixins"`
}

type PrintableCredential struct {
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description" yaml:"description"`
	Required    bool   `json:"required" yaml:"required"`
	ApplyTo     string `json:"applyTo" yaml:"applyTo"`
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

type PrintableDependency struct {
	Alias     string `json:"alias" yaml:"alias"`
	Reference string `json:"reference" yaml:"reference"`
}

type PrintableParameter struct {
	param       *bundle.Parameter
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

func (o *ExplainOpts) Validate(args []string, pctx *portercontext.Context) error {
	// Allow reference to be specified as a positional argument, or using --reference
	if len(args) == 1 {
		o.Reference = args[0]
	} else if len(args) > 1 {
		return fmt.Errorf("only one positional argument may be specified, the bundle reference, but multiple were received: %s", args)
	}

	err := o.bundleFileOptions.Validate(pctx)
	if err != nil {
		return err
	}

	err = o.ParseFormat()
	if err != nil {
		return err
	}
	if o.Reference != "" {
		o.File = ""
		o.CNABFile = ""

		return o.validateReference()
	}
	return nil
}

func (p *Porter) Explain(ctx context.Context, o ExplainOpts) error {
	bundleRef, err := p.resolveBundleReference(ctx, &o.BundleReferenceOptions)
	if err != nil {
		return err
	}

	pb, err := generatePrintable(bundleRef.Definition, o.Action)
	if err != nil {
		return fmt.Errorf("unable to print bundle: %w", err)
	}
	return p.printBundleExplain(o, pb, bundleRef.Definition)
}

func (p *Porter) printBundleExplain(o ExplainOpts, pb *PrintableBundle, bun cnab.ExtendedBundle) error {
	switch o.Format {
	case printer.FormatJson:
		return printer.PrintJson(p.Out, pb)
	case printer.FormatYaml:
		return printer.PrintYaml(p.Out, pb)
	case printer.FormatPlaintext:
		return p.printBundleExplainTable(pb, o.Reference, bun)
	default:
		return fmt.Errorf("invalid format: %s", o.Format)
	}
}

func generatePrintable(bun cnab.ExtendedBundle, action string) (*PrintableBundle, error) {
	var stamp configadapter.Stamp

	stamp, err := configadapter.LoadStamp(bun)
	if err != nil {
		stamp = configadapter.Stamp{}
	}

	solver := &cnab.DependencySolver{}
	deps, err := solver.ResolveDependencies(bun)
	if err != nil {
		return nil, fmt.Errorf("error resolving bundle dependencies: %w", err)
	}

	pb := PrintableBundle{
		Name:          bun.Name,
		Description:   bun.Description,
		Version:       bun.Version,
		PorterVersion: stamp.Version,
		Actions:       make([]PrintableAction, 0, len(bun.Actions)),
		Credentials:   make([]PrintableCredential, 0, len(bun.Credentials)),
		Parameters:    make([]PrintableParameter, 0, len(bun.Parameters)),
		Outputs:       make([]PrintableOutput, 0, len(bun.Outputs)),
		Dependencies:  make([]PrintableDependency, 0, len(deps)),
		Mixins:        make([]string, 0, len(stamp.Mixins)),
	}

	for a, v := range bun.Actions {
		pa := PrintableAction{}
		pa.Name = a
		pa.Description = v.Description
		pa.Modifies = v.Modifies
		pa.Stateless = v.Stateless
		pb.Actions = append(pb.Actions, pa)
	}
	sort.Sort(SortPrintableAction(pb.Actions))

	for c, v := range bun.Credentials {
		pc := PrintableCredential{}
		pc.Name = c
		pc.Description = v.Description
		pc.Required = v.Required
		pc.ApplyTo = generateApplyToString(v.ApplyTo)

		if shouldIncludeInExplainOutput(&v, action) {
			pb.Credentials = append(pb.Credentials, pc)
		}
	}
	sort.Sort(SortPrintableCredential(pb.Credentials))

	for p, v := range bun.Parameters {
		v := v // Go closures are funny like that
		if bun.IsInternalParameter(p) || bun.ParameterHasSource(p) {
			continue
		}

		def, ok := bun.Definitions[v.Definition]
		if !ok {
			return nil, fmt.Errorf("unable to find definition %s", v.Definition)
		}
		if def == nil {
			return nil, fmt.Errorf("empty definition for %s", v.Definition)
		}
		pp := PrintableParameter{param: &v}
		pp.Name = p
		pp.Type = bun.GetParameterType(def)
		pp.Default = def.Default
		pp.ApplyTo = generateApplyToString(v.ApplyTo)
		pp.Required = v.Required
		pp.Description = v.Description

		if shouldIncludeInExplainOutput(&v, action) {
			pb.Parameters = append(pb.Parameters, pp)
		}
	}
	sort.Sort(SortPrintableParameter(pb.Parameters))

	for o, v := range bun.Outputs {
		if bun.IsInternalOutput(o) {
			continue
		}

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

		if shouldIncludeInExplainOutput(&v, action) {
			pb.Outputs = append(pb.Outputs, po)
		}
	}
	sort.Sort(SortPrintableOutput(pb.Outputs))

	for _, dep := range deps {
		pd := PrintableDependency{}
		pd.Alias = dep.Alias
		pd.Reference = dep.Reference

		pb.Dependencies = append(pb.Dependencies, pd)
	}
	// dependencies are sorted by their dependency sequence already

	for mixin := range stamp.Mixins {
		pb.Mixins = append(pb.Mixins, mixin)
	}
	sort.Strings(pb.Mixins)

	return &pb, nil
}

// shouldIncludeInExplainOutput determine if a scoped item such as a credential, parameter or output
// should be included in the explain output.
func shouldIncludeInExplainOutput(scoped bundle.Scoped, action string) bool {
	if action == "" {
		return true
	}

	return bundle.AppliesTo(scoped, action)
}

func generateApplyToString(appliesTo []string) string {
	if len(appliesTo) == 0 {
		return "All Actions"
	}
	return strings.Join(appliesTo, ",")

}

func (p *Porter) printBundleExplainTable(bun *PrintableBundle, bundleReference string, extendedBundle cnab.ExtendedBundle) error {
	fmt.Fprintf(p.Out, "Name: %s\n", bun.Name)
	fmt.Fprintf(p.Out, "Description: %s\n", bun.Description)
	fmt.Fprintf(p.Out, "Version: %s\n", bun.Version)
	if bun.PorterVersion != "" {
		fmt.Fprintf(p.Out, "Porter Version: %s\n", bun.PorterVersion)
	}
	fmt.Fprintln(p.Out, "")

	p.printCredentialsExplainBlock(bun)
	p.printParametersExplainBlock(bun)
	p.printOutputsExplainBlock(bun)
	p.printActionsExplainBlock(bun)
	p.printDependenciesExplainBlock(bun)

	if extendedBundle.IsPorterBundle() && len(bun.Mixins) > 0 {
		fmt.Fprintf(p.Out, "This bundle uses the following tools: %s.\n", strings.Join(bun.Mixins, ", "))
	}

	if extendedBundle.SupportsDocker() {
		fmt.Fprintln(p.Out, "") // force a blank line before this block
		fmt.Fprintf(p.Out, "🚨 This bundle will grant docker access to the host, make sure the publisher of this bundle is trusted.")
		fmt.Fprintln(p.Out, "") // force a blank line after this block
	}

	p.printInstallationInstructionBlock(bun, bundleReference, extendedBundle)
	return nil
}

func (p *Porter) printCredentialsExplainBlock(bun *PrintableBundle) error {
	if len(bun.Credentials) == 0 {
		return nil
	}

	fmt.Fprintln(p.Out, "Credentials:")
	err := p.printCredentialsExplainTable(bun)
	if err != nil {
		return fmt.Errorf("unable to print credentials table: %w", err)
	}

	fmt.Fprintln(p.Out, "") // force a blank line after this block
	return nil
}
func (p *Porter) printCredentialsExplainTable(bun *PrintableBundle) error {
	printCredRow :=
		func(v interface{}) []string {
			c, ok := v.(PrintableCredential)
			if !ok {
				return nil
			}
			return []string{c.Name, c.Description, strconv.FormatBool(c.Required), c.ApplyTo}
		}
	return printer.PrintTable(p.Out, bun.Credentials, printCredRow, "Name", "Description", "Required", "Applies To")
}

func (p *Porter) printParametersExplainBlock(bun *PrintableBundle) error {
	if len(bun.Parameters) == 0 {
		return nil
	}

	fmt.Fprintln(p.Out, "Parameters:")
	err := p.printParametersExplainTable(bun)
	if err != nil {
		return fmt.Errorf("unable to print parameters table: %w", err)
	}

	fmt.Fprintln(p.Out, "") // force a blank line after this block
	return nil
}
func (p *Porter) printParametersExplainTable(bun *PrintableBundle) error {
	printParamRow :=
		func(v interface{}) []string {
			p, ok := v.(PrintableParameter)
			if !ok {
				return nil
			}
			return []string{p.Name, p.Description, fmt.Sprintf("%v", p.Type), fmt.Sprintf("%v", p.Default), strconv.FormatBool(p.Required), p.ApplyTo}
		}
	return printer.PrintTable(p.Out, bun.Parameters, printParamRow, "Name", "Description", "Type", "Default", "Required", "Applies To")
}

func (p *Porter) printOutputsExplainBlock(bun *PrintableBundle) error {
	if len(bun.Outputs) == 0 {
		return nil
	}

	fmt.Fprintln(p.Out, "Outputs:")
	err := p.printOutputsExplainTable(bun)
	if err != nil {
		return fmt.Errorf("unable to print outputs table: %w", err)
	}

	fmt.Fprintln(p.Out, "") // force a blank line after this block
	return nil
}

func (p *Porter) printOutputsExplainTable(bun *PrintableBundle) error {
	printOutputRow :=
		func(v interface{}) []string {
			o, ok := v.(PrintableOutput)
			if !ok {
				return nil
			}
			return []string{o.Name, o.Description, fmt.Sprintf("%v", o.Type), o.ApplyTo}
		}
	return printer.PrintTable(p.Out, bun.Outputs, printOutputRow, "Name", "Description", "Type", "Applies To")
}

func (p *Porter) printActionsExplainBlock(bun *PrintableBundle) error {
	if len(bun.Actions) == 0 {
		return nil
	}

	fmt.Fprintln(p.Out, "Actions:")
	err := p.printActionsExplainTable(bun)
	if err != nil {
		return fmt.Errorf("unable to print actions block: %w", err)
	}

	fmt.Fprintln(p.Out, "") // force a blank line after this block
	return nil
}

func (p *Porter) printActionsExplainTable(bun *PrintableBundle) error {
	printActionRow :=
		func(v interface{}) []string {
			a, ok := v.(PrintableAction)
			if !ok {
				return nil
			}
			return []string{a.Name, a.Description, strconv.FormatBool(a.Modifies), strconv.FormatBool(a.Stateless)}
		}
	return printer.PrintTable(p.Out, bun.Actions, printActionRow, "Name", "Description", "Modifies Installation", "Stateless")
}

// Dependencies
func (p *Porter) printDependenciesExplainBlock(bun *PrintableBundle) error {
	if len(bun.Dependencies) == 0 {
		return nil
	}

	fmt.Fprintln(p.Out, "Dependencies:")
	err := p.printDependenciesExplainTable(bun)
	if err != nil {
		return fmt.Errorf("unable to print dependencies table: %w", err)
	}

	fmt.Fprintln(p.Out, "") // force a blank line after this block
	return nil
}

func (p *Porter) printDependenciesExplainTable(bun *PrintableBundle) error {
	printDependencyRow :=
		func(v interface{}) []string {
			o, ok := v.(PrintableDependency)
			if !ok {
				return nil
			}
			return []string{o.Alias, o.Reference}
		}
	return printer.PrintTable(p.Out, bun.Dependencies, printDependencyRow, "Alias", "Reference")
}

func (p *Porter) printInstallationInstructionBlock(bun *PrintableBundle, bundleReference string, extendedBundle cnab.ExtendedBundle) error {
	fmt.Fprintln(p.Out)
	fmt.Fprint(p.Out, "To install this bundle run the following command, passing --param KEY=VALUE for any parameters you want to customize:\n")

	var bundleReferenceFlag string
	if bundleReference != "" {
		bundleReferenceFlag += " --reference " + bundleReference
	}

	// Generate predefined credential set first.
	if len(bun.Credentials) > 0 {
		fmt.Fprintf(p.Out, "porter credentials generate mycreds%s\n", bundleReferenceFlag)
	}

	// Bundle installation instruction
	var requiredParameterFlags string
	for _, parameter := range bun.Parameters {
		// Only include parameters required for install
		if parameter.Required && shouldIncludeInExplainOutput(parameter.param, cnab.ActionInstall) {
			requiredParameterFlags += parameter.Name + "=TODO "
		}
	}

	if requiredParameterFlags != "" {
		requiredParameterFlags = " --param " + requiredParameterFlags
	}

	var credentialFlags string
	if len(bun.Credentials) > 0 {
		credentialFlags += " --credential-set mycreds"
	}

	porterInstallCommand := fmt.Sprintf("porter install%s%s%s", bundleReferenceFlag, requiredParameterFlags, credentialFlags)

	// Check whether the bundle requires docker socket to be mounted into the bundle.
	// Add flag for docker host access for install command if it requires to do so.
	if extendedBundle.SupportsDocker() {
		porterInstallCommand += " --allow-docker-host-access"
	}

	fmt.Fprint(p.Out, porterInstallCommand)
	fmt.Fprintln(p.Out, "") // force a blank line after this block

	return nil
}
