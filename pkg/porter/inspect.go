package porter

import (
	"encoding/json"
	"fmt"

	"github.com/deislabs/cnab-go/bundle"
	"github.com/deislabs/porter/pkg/printer"
	"github.com/pkg/errors"
)

type InspectableBundle struct {
	*PrintableBundle
	InvocationImages []bundle.InvocationImage `json:"invocationImages" yaml:"invocationImages"`
	Images           []PrintableImage         `json:"images,omitempty" yaml:"images,omitempty"`
}

type PrintableImage struct {
	Name string `json:"name" yaml:"name"`
	bundle.Image
}

func (p *Porter) Inspect(o ExplainOpts) error {
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
	bundle, err := p.CNAB.LoadBundle(o.CNABFile, o.Insecure)
	if err != nil {
		return errors.Wrap(err, "unable to load bundle")
	}
	if o.RelocationMapping != "" {
		var reloMap map[string]string
		reloBytes, err := p.FileSystem.ReadFile(o.RelocationMapping)
		if err != nil {
			return errors.Wrap(err, "unable to read provided relocation mapping")
		}
		err = json.Unmarshal(reloBytes, &reloMap)
		if err != nil {
			return errors.Wrap(err, "unable to load provided relocation mapping")
		}
		handleInspectRelocate(bundle, reloMap)
	}

	pb, err := generatePrintable(bundle)
	if err != nil {
		return errors.Wrap(err, "unable to print bundle")
	}

	printableImages := generateInspectImages(bundle)
	ib := &InspectableBundle{
		PrintableBundle:  pb,
		InvocationImages: bundle.InvocationImages,
		Images:           printableImages,
	}
	return p.printBundleInspect(o, ib)
}

func generateInspectImages(bun *bundle.Bundle) []PrintableImage {
	var printableImages []PrintableImage
	for alias, img := range bun.Images {
		pi := PrintableImage{
			Name:  alias,
			Image: img,
		}
		printableImages = append(printableImages, pi)
	}
	return printableImages
}

func (p *Porter) printBundleInspect(o ExplainOpts, ib *InspectableBundle) error {
	switch o.Format {
	case printer.FormatJson:
		return printer.PrintJson(p.Out, ib)
	case printer.FormatYaml:
		return printer.PrintYaml(p.Out, ib)
	case printer.FormatTable:
		return p.printBundleInspectTable(ib)
	default:
		return fmt.Errorf("invalid format: %s", o.Format)
	}
}

func (p *Porter) printBundleInspectTable(bun *InspectableBundle) error {
	err := p.printBundleExplainTable(bun.PrintableBundle)
	if err != nil {
		return errors.Wrap(err, "unable to inspect bundle")
	}
	p.printInvocationImageInspectBlock(bun)
	p.printImagesInspectBlock(bun)
	return nil
}

func handleInspectRelocate(bun *bundle.Bundle, reloMap map[string]string) {
	for idx, invoImage := range bun.InvocationImages {
		if mappedInvo, ok := reloMap[invoImage.Image]; ok {
			invoImage.Image = mappedInvo
		}
		bun.InvocationImages[idx] = invoImage
	}
	for alias, image := range bun.Images {
		if mappedImg, ok := reloMap[image.Image]; ok {
			image.Image = mappedImg
		}
		bun.Images[alias] = image
	}
}

func (p *Porter) printInvocationImageInspectBlock(bun *InspectableBundle) error {
	fmt.Fprintln(p.Out, "Invocation Images:")
	err := p.printInvocationImageInspectTable(bun)
	if err != nil {
		return errors.Wrap(err, "unable to print invocation images table")
	}
	fmt.Fprintln(p.Out, "") // force a blank line after this block
	return nil
}

func (p *Porter) printInvocationImageInspectTable(bun *InspectableBundle) error {
	printInvocationImageRow :=
		func(v interface{}) []interface{} {
			ii, ok := v.(bundle.InvocationImage)
			if !ok {
				return nil
			}
			return []interface{}{ii.Image, ii.ImageType, ii.Digest}
		}
	return printer.PrintTable(p.Out, bun.InvocationImages, printInvocationImageRow, "Image", "Type", "Digest")
}

func (p *Porter) printImagesInspectBlock(bun *InspectableBundle) error {
	fmt.Fprintln(p.Out, "Images:")
	if len(bun.Images) > 0 {
		err := p.printImagesInspectTable(bun)
		if err != nil {
			return errors.Wrap(err, "unable to print images table")
		}
		fmt.Fprintln(p.Out, "") // force a blank line after this block
	} else {
		fmt.Fprintln(p.Out, "No images defined")
	}
	return nil
}

func (p *Porter) printImagesInspectTable(bun *InspectableBundle) error {
	printImageRow :=
		func(v interface{}) []interface{} {
			pi, ok := v.(PrintableImage)
			if !ok {
				return nil
			}
			return []interface{}{pi.Name, pi.ImageType, pi.Image.Image, pi.Digest}
		}
	return printer.PrintTable(p.Out, bun.Images, printImageRow, "Name", "Type", "Image", "Digest")
}
