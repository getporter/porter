package coras

import (
	"fmt"
	"strings"

	"github.com/deislabs/cnab-go/bundle"
	"github.com/docker/distribution/reference"
	"github.com/pivotal/image-relocation/pkg/image"
	"github.com/pivotal/image-relocation/pkg/registry"
)

// RelocateBundleImages pushes all referenced images to the a new repository.
// In the new repository, images are uniquely identified by the digest.
// Currently, each image also has a unique tag, but the human readable tag should
// never be used when referencing the image.
//
// The bundle is mutated in place, and contains the new image location (and digest, if it wasn't previously present)
func RelocateBundleImages(b *bundle.Bundle, targetRef string) error {
	rc := registry.NewRegistryClient()
	for i := range b.InvocationImages {
		ii := b.InvocationImages[i]
		_, err := relocateImage(&ii.BaseImage, targetRef, rc)
		if err != nil {
			return err
		}
		b.InvocationImages[i] = ii
	}

	for i := range b.Images {
		im := b.Images[i]
		_, err := relocateImage(&im.BaseImage, targetRef, rc)
		if err != nil {
			return err
		}
		b.Images[i] = im
	}

	return nil
}

func relocateImage(i *bundle.BaseImage, targetRef string, client registry.Client) (bool, error) {
	if !isAcceptedImageType(i.ImageType) {
		return false, fmt.Errorf("cannot relocate image of type %v", i.ImageType)
	}

	// original image FQDN
	originalImage, err := image.NewName(i.Image)
	if err != nil {
		return false, fmt.Errorf("cannot get fully qualified image name for %v: %v", i.Image, err)
	}

	// relocated image FQDN
	// the location of the new image is the same target repository as the bundle itself,
	// but a different, unique tag.
	//
	// Note that the tag is just a familiar name, and it is not used
	// All references to this image are through the SHA digest
	//
	// TODO - @radu-matei
	// make sure naming strategy is consistent

	nn, err := TransformImageName(i.Image, targetRef)
	if err != nil {
		return false, fmt.Errorf("cannot parse new image name: %v", err)
	}

	newImage, err := image.NewName(nn)
	if err != nil {
		return false, fmt.Errorf("cannot get fully qualified image name for the new image in %v: %v", targetRef, err)
	}

	dig, err := client.Copy(originalImage, newImage)
	if err != nil {
		return false, fmt.Errorf("cannot copy original image %v into new image %v: %v", originalImage.Name(), newImage.Name(), err)
	}

	// TODO - @radu-matei
	// make sure the digest is not modified, and return true if it is
	i.Digest = dig.String()
	i.OriginalImage = i.Image
	i.Image = newImage.String()

	return false, nil
}

func TransformImageName(inputImage, targetRef string) (string, error) {
	tref, err := reference.ParseNormalizedNamed(targetRef)
	if err != nil {
		return "", fmt.Errorf("cannot parse target reference: %v", err)
	}

	nt, ok := tref.(reference.NamedTagged)
	if !ok {
		return "", fmt.Errorf("cannot use target reference that doesn't contain a tag")
	}
	repo := strings.Split(nt.Name(), fmt.Sprintf(":%s", nt.Tag()))[0]

	ref, err := reference.ParseNormalizedNamed(inputImage)
	if err != nil {
		return "", fmt.Errorf("cannot parse image reference: %v", err)
	}
	return fmt.Sprintf("%v:%v", repo, removeSlashColumnAt(ref.String())), nil

}

func removeSlashColumnAt(input string) string {
	return strings.Replace(strings.Replace(strings.Replace(input, "/", "-", -1), ":", "-", -1), "@", "-", -1)
}

func isAcceptedImageType(imageType string) bool {
	return imageType == "" || imageType == "oci" || imageType == "docker"
}
