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
	bundleRef, err := p.resolveBundleReference(ctx, &o.BundleReferenceOptions)
	if err != nil {
		return err
	}

	ib, err := generateInspectableBundle(bundleRef)
	if err != nil {
		return fmt.Errorf("unable to inspect bundle: %w", err)
	}
	return p.printBundleInspect(o, ib)
}

func generateInspectableBundle(bundleRef cnab.BundleReference) (*InspectableBundle, error) {
	ib := &InspectableBundle{
		Name:        bundleRef.Definition.Name,
		Description: bundleRef.Definition.Description,
		Version:     bundleRef.Definition.Version,
	}
	ib.InvocationImages, ib.Images = handleInspectRelocate(bundleRef)
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

	p.printInvocationImageInspectBlock(bun)
	p.printImagesInspectBlock(bun)
	return nil
}

func (p *Porter) printInvocationImageInspectBlock(bun *InspectableBundle) error {
	fmt.Fprintln(p.Out, "Invocation Images:")
	err := p.printInvocationImageInspectTable(bun)
	if err != nil {
		return fmt.Errorf("unable to print invocation images table: %w", err)
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
