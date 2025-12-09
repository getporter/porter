package porter

import (
	"context"
	"fmt"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/printer"
	"github.com/cnabio/cnab-go/bundle"
)

type InspectableBundle struct {
	Name             string                     `json:"name" yaml:"name"`
	Description      string                     `json:"description,omitempty" yaml:"description,omitempty"`
	Version          string                     `json:"version" yaml:"version"`
	InvocationImages []PrintableInvocationImage `json:"invocationImages" yaml:"invocationImages"`
	Images           []PrintableImage           `json:"images,omitempty" yaml:"images,omitempty"`
	Dependencies     []InspectableDependency    `json:"dependencies,omitempty" yaml:"dependencies,omitempty"`
}

type InspectableDependency struct {
	Alias            string                  `json:"alias" yaml:"alias"`
	Reference        string                  `json:"reference" yaml:"reference"`
	Version          string                  `json:"version,omitempty" yaml:"version,omitempty"`
	Depth            int                     `json:"depth" yaml:"depth"`
	SharingMode      bool                    `json:"sharingMode,omitempty" yaml:"sharingMode,omitempty"`
	SharingGroup     string                  `json:"sharingGroup,omitempty" yaml:"sharingGroup,omitempty"`
	Parameters       map[string]string       `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	Credentials      map[string]string       `json:"credentials,omitempty" yaml:"credentials,omitempty"`
	Outputs          map[string]string       `json:"outputs,omitempty" yaml:"outputs,omitempty"`
	Dependencies     []InspectableDependency `json:"dependencies,omitempty" yaml:"dependencies,omitempty"`
	ResolutionError  string                  `json:"resolutionError,omitempty" yaml:"resolutionError,omitempty"`
	ResolutionFailed bool                    `json:"-" yaml:"-"`
}

type PrintableInvocationImage struct {
	bundle.InvocationImage
	Original string `json:"originalImage" yaml:"originalImage"`
}

type PrintableImage struct {
	Name string `json:"name" yaml:"name"`
	bundle.Image
	Original string `json:"originalImage" yaml:"originalImage"`
}

func (p *Porter) Inspect(ctx context.Context, o ExplainOpts) error {
	bundleRef, err := o.GetBundleReference(ctx, p)
	if err != nil {
		return err
	}

	ib, err := generateInspectableBundle(ctx, p, bundleRef, o)
	if err != nil {
		return fmt.Errorf("unable to inspect bundle: %w", err)
	}
	return p.printBundleInspect(o, ib)
}

func generateInspectableBundle(ctx context.Context, p *Porter, bundleRef cnab.BundleReference, opts ExplainOpts) (*InspectableBundle, error) {
	ib := &InspectableBundle{
		Name:        bundleRef.Definition.Name,
		Description: bundleRef.Definition.Description,
		Version:     bundleRef.Definition.Version,
	}
	ib.InvocationImages, ib.Images = handleInspectRelocate(bundleRef)

	// Build dependency tree when flag is set
	if opts.ShowDependencies && (bundleRef.Definition.HasDependenciesV1() || bundleRef.Definition.HasDependenciesV2()) {
		builder := NewDependencyTreeBuilder(p, opts.MaxDependencyDepth)
		deps, err := builder.BuildDependencyTree(ctx, bundleRef.Definition, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to build dependency tree: %w", err)
		}
		ib.Dependencies = deps
	}

	return ib, nil
}

func handleInspectRelocate(bundleRef cnab.BundleReference) ([]PrintableInvocationImage, []PrintableImage) {
	invoImages := []PrintableInvocationImage{}
	for _, invoImage := range bundleRef.Definition.InvocationImages {
		pii := PrintableInvocationImage{
			InvocationImage: invoImage,
		}
		if mappedInvo, ok := bundleRef.RelocationMap[invoImage.Image]; ok {
			pii.Original = pii.Image
			pii.Image = mappedInvo
		}
		invoImages = append(invoImages, pii)
	}
	images := []PrintableImage{}
	for alias, image := range bundleRef.Definition.Images {
		pi := PrintableImage{
			Name:  alias,
			Image: image,
		}
		if mappedImg, ok := bundleRef.RelocationMap[image.Image]; ok {
			pi.Original = pi.Image.Image
			pi.Image.Image = mappedImg
		}
		images = append(images, pi)
	}
	return invoImages, images
}

func (p *Porter) printBundleInspect(o ExplainOpts, ib *InspectableBundle) error {
	switch o.Format {
	case printer.FormatJson:
		return printer.PrintJson(p.Out, ib)
	case printer.FormatYaml:
		return printer.PrintYaml(p.Out, ib)
	case printer.FormatPlaintext:
		return p.printBundleInspectTable(ib)
	default:
		return fmt.Errorf("invalid format: %s", o.Format)
	}
}

func (p *Porter) printBundleInspectTable(bun *InspectableBundle) error {
	fmt.Fprintf(p.Out, "Name: %s\n", bun.Name)
	fmt.Fprintf(p.Out, "Description: %s\n", bun.Description)
	fmt.Fprintf(p.Out, "Version: %s\n", bun.Version)
	fmt.Fprintln(p.Out, "")

	err := p.printInvocationImageInspectBlock(bun)
	if err != nil {
		return err
	}
	err = p.printImagesInspectBlock(bun)
	if err != nil {
		return err
	}
	err = p.printDependenciesInspectBlock(bun)
	if err != nil {
		return err
	}
	return nil
}

func (p *Porter) printInvocationImageInspectBlock(bun *InspectableBundle) error {
	fmt.Fprintln(p.Out, "Bundle Images:")
	err := p.printInvocationImageInspectTable(bun)
	if err != nil {
		return fmt.Errorf("unable to print bundle images table: %w", err)
	}
	fmt.Fprintln(p.Out, "") // force a blank line after this block
	return nil
}

func (p *Porter) printInvocationImageInspectTable(bun *InspectableBundle) error {
	printInvocationImageRow :=
		func(v interface{}) []string {
			ii, ok := v.(PrintableInvocationImage)
			if !ok {
				return nil
			}
			return []string{ii.Image, ii.ImageType, ii.Digest, ii.Original}
		}
	return printer.PrintTable(p.Out, bun.InvocationImages, printInvocationImageRow, "Image", "Type", "Digest", "Original Image")
}

func (p *Porter) printImagesInspectBlock(bun *InspectableBundle) error {
	if len(bun.Images) == 0 {
		return nil
	}

	fmt.Fprintln(p.Out, "Images:")
	err := p.printImagesInspectTable(bun)
	if err != nil {
		return fmt.Errorf("unable to print images table: %w", err)
	}
	fmt.Fprintln(p.Out, "") // force a blank line after this block

	return nil
}

func (p *Porter) printImagesInspectTable(bun *InspectableBundle) error {
	printImageRow :=
		func(v interface{}) []string {
			pi, ok := v.(PrintableImage)
			if !ok {
				return nil
			}
			return []string{pi.Name, pi.ImageType, pi.Image.Image, pi.Digest, pi.Original}
		}
	return printer.PrintTable(p.Out, bun.Images, printImageRow, "Name", "Type", "Image", "Digest", "Original Image")
}

func (p *Porter) printDependenciesInspectBlock(bun *InspectableBundle) error {
	if len(bun.Dependencies) == 0 {
		return nil
	}

	fmt.Fprintln(p.Out, "Dependencies:")
	err := p.printDependenciesInspectTable(bun)
	if err != nil {
		return fmt.Errorf("unable to print dependencies table: %w", err)
	}

	// Check if any dependencies failed to resolve
	if hasFailedDependencies(bun.Dependencies) {
		fmt.Fprintln(p.Out, "")
		fmt.Fprintln(p.Out, "Some dependencies failed to resolve. Use --output json or --output yaml to see detailed error messages.")
	}

	fmt.Fprintln(p.Out, "") // force a blank line after this block

	return nil
}

func (p *Porter) printDependenciesInspectTable(bun *InspectableBundle) error {
	// Flatten the tree for table display
	flatDeps := flattenDependencyTree(bun.Dependencies)

	printDependencyRow :=
		func(v interface{}) []string {
			dep, ok := v.(InspectableDependency)
			if !ok {
				return nil
			}

			// Add indentation based on depth
			indent := ""
			for i := 0; i < dep.Depth; i++ {
				indent += "  "
			}

			// Add warning emoji for failed dependencies
			prefix := ""
			if dep.ResolutionFailed {
				prefix = "⚠️ "
			}
			alias := indent + prefix + dep.Alias

			// Format sharing info
			sharing := ""
			if dep.SharingMode && dep.SharingGroup != "" {
				sharing = fmt.Sprintf("[shared:%s]", dep.SharingGroup)
			}

			return []string{alias, dep.Reference, dep.Version, sharing}
		}

	return printer.PrintTable(p.Out, flatDeps, printDependencyRow, "Alias", "Reference", "Version", "Sharing")
}

// hasFailedDependencies recursively checks if any dependency in the tree failed to resolve
func hasFailedDependencies(deps []InspectableDependency) bool {
	for _, dep := range deps {
		if dep.ResolutionFailed {
			return true
		}
		if hasFailedDependencies(dep.Dependencies) {
			return true
		}
	}
	return false
}
