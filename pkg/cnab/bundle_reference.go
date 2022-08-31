package cnab

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/cnabio/cnab-to-oci/relocation"
	"github.com/opencontainers/go-digest"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type BundleReference struct {
	Reference     OCIReference
	Digest        digest.Digest
	Definition    ExtendedBundle
	RelocationMap relocation.ImageRelocationMap
}

func (r BundleReference) String() string {
	return r.Reference.String()
}

// AddToTrace appends the bundle reference attributes to the current span.
func (r BundleReference) AddToTrace(cxt context.Context) {
	span := trace.SpanFromContext(cxt)

	span.SetAttributes(attribute.String("reference", r.String()))

	var bunJson bytes.Buffer
	_, _ = r.Definition.WriteTo(&bunJson)
	span.SetAttributes(attribute.String("bundleDefinition", bunJson.String()))

	relocationMappingJson, _ := json.Marshal(r.RelocationMap)
	span.SetAttributes(attribute.String("relocationMapping", string(relocationMappingJson)))
}

// CalculateTemporaryImageTag returns the temporary tag applied to images that we
// use to push it and then retrieve the repository digest for an image.
func CalculateTemporaryImageTag(bunRef OCIReference) (OCIReference, error) {
	imageName, err := ParseOCIReference(bunRef.Repository())
	if err != nil {
		return OCIReference{}, fmt.Errorf("could not calculate temporary image tag for %s: %w", bunRef.String(), err)
	}
	referenceHash := md5.Sum([]byte(bunRef.String()))

	// prefix the temporary tag with porter- so that people can identify its purpose (otherwise it's just some random number as far as people can tell)
	imgTag := "porter-" + hex.EncodeToString(referenceHash[:])
	imageRef, err := imageName.WithTag(imgTag)
	if err != nil {
		return OCIReference{}, fmt.Errorf("could not apply tag, %s, to the image %s: %w", imgTag, imageName.String(), err)
	}

	return imageRef, nil
}
