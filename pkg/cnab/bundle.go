package cnab

import (
	"github.com/cnabio/cnab-go/bundle/definition"
)

func IsFileType(s *definition.Schema) bool {
	return s.Type == "string" && s.ContentEncoding == "base64"
}
