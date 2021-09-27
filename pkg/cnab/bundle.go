package cnab

import (
	"bytes"
	"context"
	"encoding/json"

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
