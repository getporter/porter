package porter

import (
	"encoding/json"
	"fmt"

	"get.porter.sh/porter/pkg/printer"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/pkg/errors"
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

func (p *Porter) Inspect(o ExplainOpts) error {
	err := p.prepullBundleByTag(&o.BundleActionOptions)
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
	if err != nil {
		return errors.Wrap(err, "unable to load bundle")
	}
	var reloMap map[string]string
	if o.RelocationMapping != "" {
		reloBytes, err := p.FileSystem.ReadFile(o.RelocationMapping)
		if err != nil {
			return errors.Wrap(err, "unable to read provided relocation mapping")
		}
		err = json.Unmarshal(reloBytes, &reloMap)
		if err != nil {
			return errors.Wrap(err, "unable to load provided relocation mapping")
		}
	}

	ib, err := generateInspectableBundle(bundle, reloMap)
	if err != nil {
		return errors.Wrap(err, "unable to inspect bundle")
	}
	return p.printBundleInspect(o, ib)
}

func generateInspectableBundle(bun bundle.Bundle, reloMap map[string]string) (*InspectableBundle, error) {
	ib := &InspectableBundle{
		Name:        bun.Name,
		Description: bun.Description,
		Version:     bun.Version,
	}
	ib.InvocationImages, ib.Images = handleInspectRelocate(bun, reloMap)
	return ib, nil
}

func handleInspectRelocate(bun bundle.Bundle, reloMap map[string]string) ([]PrintableInvocationImage, []PrintableImage) {
	invoImages := []PrintableInvocationImage{}
	for _, invoImage := range bun.InvocationImages {
		pii := PrintableInvocationImage{
			InvocationImage: invoImage,
		}
		if mappedInvo, ok := reloMap[invoImage.Image]; ok {
			pii.Original = pii.Image
			pii.Image = mappedInvo
		}
		invoImages = append(invoImages, pii)
	}
	images := []PrintableImage{}
	for alias, image := range bun.Images {
		pi := PrintableImage{
			Name:  alias,
			Image: image,
		}
		if mappedImg, ok := reloMap[image.Image]; ok {
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
	case printer.FormatTable:
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
		return errors.Wrap(err, "unable to print invocation images table")
	}
	fmt.Fprintln(p.Out, "") // force a blank line after this block
	return nil
}

func (p *Porter) printInvocationImageInspectTable(bun *InspectableBundle) error {
	printInvocationImageRow :=
		func(v interface{}) []interface{} {
			ii, ok := v.(PrintableInvocationImage)
			if !ok {
				return nil
			}
			return []interface{}{ii.Image, ii.ImageType, ii.Digest, ii.Original}
		}
	return printer.PrintTable(p.Out, bun.InvocationImages, printInvocationImageRow, "Image", "Type", "Digest", "Original Image")
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
			return []interface{}{pi.Name, pi.ImageType, pi.Image.Image, pi.Digest, pi.Original}
		}
	return printer.PrintTable(p.Out, bun.Images, printImageRow, "Name", "Type", "Image", "Digest", "Original Image")
}
