package porter

import (
	"context"
	"fmt"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/pkg/printer"
	"github.com/cnabio/cnab-go/bundle"
)

type InspectOpts struct {
	BundleReferenceOptions
	printer.PrintOptions

	// ResolveTags will resolve tags if true
	ResolveTags bool
}

func (o *InspectOpts) Validate(args []string, pctx *portercontext.Context) error {
	// Allow reference to be specified as a positional argument, or using --reference
	if len(args) == 1 {
		o.Reference = args[0]
	} else if len(args) > 1 {
		return fmt.Errorf("only one positional argument may be specified, the bundle reference, but multiple were received: %s", args)
	}

	err := o.BundleDefinitionOptions.Validate(pctx)
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

func (p *Porter) Inspect(ctx context.Context, o InspectOpts) error {
	bundleRef, err := o.GetBundleReference(ctx, p)
	if err != nil {
		return err
	}

	ib, err := generateInspectableBundle(bundleRef, &o)
	if err != nil {
		return fmt.Errorf("unable to inspect bundle: %w", err)
	}
	return p.printBundleInspect(o, ib)
}

func generateInspectableBundle(bundleRef cnab.BundleReference, opts *InspectOpts) (*InspectableBundle, error) {
	ib := &InspectableBundle{
		Name:        bundleRef.Definition.Name,
		Description: bundleRef.Definition.Description,
		Version:     bundleRef.Definition.Version,
	}
	var err error
	ib.InvocationImages, ib.Images, err = handleInspectRelocate(bundleRef, opts)
	if err != nil {
		return nil, err
	}
	return ib, nil
}

func handleInspectRelocate(bundleRef cnab.BundleReference, opts *InspectOpts) ([]PrintableInvocationImage, []PrintableImage, error) {
	invoImages := []PrintableInvocationImage{}
	for _, invoImage := range bundleRef.Definition.InvocationImages {
		pii := PrintableInvocationImage{
			InvocationImage: invoImage,
		}
		if mappedInvo, ok := bundleRef.RelocationMap[invoImage.Image]; ok {
			pii.Original = pii.Image
			pii.Image = mappedInvo
		}
		if opts.ResolveTags {
			originalRef, err := getMatchingTag(pii.Original, opts)
			if err != nil {
				return nil, nil, err
			}
			pii.Original = originalRef.String()
			imageRef, err := getMatchingTag(pii.Image, opts)
			if err != nil {
				return nil, nil, err
			}
			pii.Image = imageRef.String()
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
		if opts.ResolveTags {
			originalRef, err := getMatchingTag(pi.Original, opts)
			if err != nil {
				return nil, nil, err
			}
			pi.Original = originalRef.String()
			imageRef, err := getMatchingTag(pi.Image.Image, opts)
			if err != nil {
				return nil, nil, err
			}
			pi.Image.Image = imageRef.String()
		}
		images = append(images, pi)
	}
	return invoImages, images, nil
}

func (p *Porter) printBundleInspect(o InspectOpts, ib *InspectableBundle) error {
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
			return []string{ii.Image, ii.Digest, ii.Original}
		}
	return printer.PrintTable(p.Out, bun.InvocationImages, printInvocationImageRow, "Image", "Digest", "Original Image")
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

func getMatchingTag(ref string, opts *InspectOpts) (cnab.OCIReference, error) {
	imgRef, err := cnab.ParseOCIReference(ref)
	if err != nil {
		return cnab.OCIReference{}, err
	}

	if !imgRef.HasDigest() {
		return imgRef, nil
	}

	image, err := imgRef.FindTagMatchingDigest(opts.InsecureRegistry)
	if err != nil {
		return cnab.OCIReference{}, err
	}

	return image, nil
}
