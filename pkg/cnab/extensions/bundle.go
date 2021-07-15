package extensions

import (
	"fmt"

	claim "get.porter.sh/porter/pkg/cnab"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-go/bundle/definition"
)

// IsPorterBundle determines if the bundle was created by Porter.
func IsPorterBundle(b bundle.Bundle) bool {
	_, madeByPorter := b.Custom["sh.porter"]
	return madeByPorter
}

// GetParameterType determines the type of parameter accounting for
// Porter-specific parameter types like file.
func GetParameterType(b bundle.Bundle, def *definition.Schema) string {
	if IsFileType(b, def) {
		return "file"
	}

	if def.ID == claim.OutputInvocationImageLogs {
		return "string"
	}

	return fmt.Sprintf("%v", def.Type)
}

// IsFileType determines if the parameter/credential is of type "file".
func IsFileType(b bundle.Bundle, def *definition.Schema) bool {
	return SupportsFileParameters(b) &&
		def.Type == "string" && def.ContentEncoding == "base64"
}
